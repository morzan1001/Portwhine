package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
	workerv1 "github.com/portwhine/portwhine/gen/go/portwhine/worker/v1"
	"github.com/portwhine/portwhine/gen/go/portwhine/worker/v1/workerv1connect"
	"github.com/portwhine/portwhine/pkg/server"
)

type emailConfig struct {
	recipients []string
	subject    string
	format     string // "html" or "text"
	smtpHost   string
	smtpPort   int
	smtpUser   string
	smtpPass   string
}

type itemEntry struct {
	ID     string
	Type   string
	Data   map[string]interface{}
	Source string
}

type workerHandler struct {
	workerv1connect.UnimplementedWorkerServiceHandler

	mu             sync.Mutex
	streamMu       sync.Mutex
	nodeID         string
	pipelineRunID  string
	config         *emailConfig
	status         portwhinev1.WorkerStatus
	itemsProcessed atomic.Uint64
	itemsProduced  atomic.Uint64
	errorsCount    atomic.Uint64
	initialized    bool
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	handler := &workerHandler{}

	mux := http.NewServeMux()
	path, h := workerv1connect.NewWorkerServiceHandler(handler)
	mux.Handle(path, h)

	server.MustListenAndServe(mux)
}

func (h *workerHandler) GetCapabilities(_ context.Context, _ *connect.Request[workerv1.GetCapabilitiesRequest]) (*connect.Response[workerv1.GetCapabilitiesResponse], error) {
	return connect.NewResponse(&workerv1.GetCapabilitiesResponse{
		Capability: &portwhinev1.WorkerCapability{
			Name:               "email-output",
			Version:            "1.0.0",
			AcceptedInputTypes: []string{"service", "url", "vulnerability", "ssl_result", "http_headers", "web_technology", "screenshot", "ssh_audit_result", "whois_result", "ip_address", "dns_record", "domain", "report"},
			OutputTypes:        []string{"email_delivery"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "recipients": {"type": "array", "items": {"type": "string"}, "description": "E-mail addresses to send the report to"},
    "subject": {"type": "string", "description": "E-mail subject line", "default": "Portwhine Pipeline Report"},
    "format": {"type": "string", "description": "Report format: html or text", "default": "html"},
    "smtp_host": {"type": "string", "description": "SMTP server hostname"},
    "smtp_port": {"type": "number", "description": "SMTP server port", "default": 587},
    "smtp_user": {"type": "string", "description": "SMTP username"},
    "smtp_pass": {"type": "string", "description": "SMTP password"}
  },
  "required": ["recipients"]
}`,
		},
	}), nil
}

func (h *workerHandler) Initialize(_ context.Context, req *connect.Request[workerv1.InitializeRequest]) (*connect.Response[workerv1.InitializeResponse], error) {
	config := req.Msg.GetConfig()
	if config == nil {
		return connect.NewResponse(&workerv1.InitializeResponse{
			Success:      false,
			ErrorMessage: "missing stage config",
		}), nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.nodeID = config.GetNodeId()
	h.pipelineRunID = config.GetPipelineRunId()

	cfg := &emailConfig{
		subject:  "Portwhine Pipeline Report",
		format:   "html",
		smtpPort: 587,
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["recipients"]; ok {
			if lv := v.GetListValue(); lv != nil {
				for _, item := range lv.GetValues() {
					if s := item.GetStringValue(); s != "" {
						cfg.recipients = append(cfg.recipients, s)
					}
				}
			}
		}
		if v, ok := fields["subject"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.subject = s
			}
		}
		if v, ok := fields["format"]; ok {
			if s := v.GetStringValue(); s == "html" || s == "text" {
				cfg.format = s
			}
		}
		if v, ok := fields["smtp_host"]; ok {
			cfg.smtpHost = v.GetStringValue()
		}
		if v, ok := fields["smtp_port"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.smtpPort = int(n)
			}
		}
		if v, ok := fields["smtp_user"]; ok {
			cfg.smtpUser = v.GetStringValue()
		}
		if v, ok := fields["smtp_pass"]; ok {
			cfg.smtpPass = v.GetStringValue()
		}
	}

	if len(cfg.recipients) == 0 {
		return connect.NewResponse(&workerv1.InitializeResponse{
			Success:      false,
			ErrorMessage: "recipients parameter is required",
		}), nil
	}

	if cfg.smtpHost == "" {
		slog.Warn("smtp_host not configured, emails will be skipped")
	}

	h.config = cfg
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("email-output initialized",
		"recipients", cfg.recipients,
		"subject", cfg.subject,
		"format", cfg.format,
		"smtp_host", cfg.smtpHost,
	)

	return connect.NewResponse(&workerv1.InitializeResponse{Success: true}), nil
}

func (h *workerHandler) Process(ctx context.Context, stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse]) error {
	if !h.initialized {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("not initialized"))
	}
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_PROCESSING

	heartbeatDone := make(chan struct{})
	go h.sendHeartbeats(stream, heartbeatDone)
	defer close(heartbeatDone)

	var entries []itemEntry

	for {
		req, err := stream.Receive()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		switch payload := req.GetPayload().(type) {
		case *workerv1.ProcessRequest_Item:
			entries = append(entries, convertToEntry(payload.Item))
			h.itemsProcessed.Add(1)
			slog.Debug("buffered item", "type", payload.Item.GetType(), "id", payload.Item.GetId(), "buffered_count", len(entries))
		case *workerv1.ProcessRequest_Flush:
			slog.Info("flush received, sending email", "buffered_count", len(entries))
			if len(entries) > 0 {
				h.sendEmailReport(stream, entries)
				entries = nil
			}
		case *workerv1.ProcessRequest_Cancel:
			slog.Info("cancel received", "reason", payload.Cancel.GetReason())
			return nil
		}
	}

	// Send email from remaining items on EOF.
	if len(entries) > 0 {
		slog.Info("EOF reached, sending email", "buffered_count", len(entries))
		h.sendEmailReport(stream, entries)
	}

	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	return nil
}

func convertToEntry(item *portwhinev1.DataItem) itemEntry {
	entry := itemEntry{
		ID:   item.GetId(),
		Type: item.GetType(),
		Data: make(map[string]interface{}),
	}
	if item.GetData() != nil {
		for k, v := range item.GetData().GetFields() {
			entry.Data[k] = v.AsInterface()
		}
	}
	if item.GetMetadata() != nil {
		entry.Source = item.GetMetadata().GetSource()
	}
	return entry
}

func (h *workerHandler) sendEmailReport(stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], entries []itemEntry) {
	// Count items by type.
	typeCounts := make(map[string]int)
	for _, e := range entries {
		typeCounts[e.Type]++
	}

	var body string
	if h.config.format == "html" {
		body = h.buildHTMLReport(entries, typeCounts)
	} else {
		body = h.buildTextReport(entries, typeCounts)
	}

	if h.config.smtpHost == "" {
		slog.Warn("smtp_host not configured, skipping email send", "recipients", h.config.recipients, "items", len(entries))
		h.emitDelivery(stream, len(entries), "skipped (no smtp_host)")
		return
	}

	if err := h.sendSMTP(body); err != nil {
		slog.Error("failed to send email", "error", err)
		h.errorsCount.Add(1)
		h.safeSend(stream, &workerv1.ProcessResponse{
			Payload: &workerv1.ProcessResponse_Error{
				Error: &workerv1.ProcessError{
					ErrorMessage: fmt.Sprintf("email send failed: %v", err),
					Retryable:    true,
				},
			},
		})
		return
	}

	h.emitDelivery(stream, len(entries), "sent")
}

func (h *workerHandler) buildHTMLReport(entries []itemEntry, typeCounts map[string]int) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8">`)
	b.WriteString(`<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 0; padding: 20px; background: #f5f5f5; color: #333; }
.container { max-width: 800px; margin: 0 auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
.header { background: #1a1a2e; color: #fff; padding: 24px 32px; }
.header h1 { margin: 0 0 8px 0; font-size: 22px; }
.header p { margin: 0; opacity: 0.7; font-size: 14px; }
.content { padding: 24px 32px; }
table { width: 100%; border-collapse: collapse; margin: 16px 0; }
th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid #eee; font-size: 14px; }
th { background: #f8f9fa; font-weight: 600; color: #555; }
.badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
.badge-critical { background: #fee2e2; color: #dc2626; }
.badge-high { background: #ffedd5; color: #ea580c; }
.badge-medium { background: #fef3c7; color: #d97706; }
.badge-low { background: #dbeafe; color: #2563eb; }
.badge-info { background: #f0fdf4; color: #16a34a; }
.footer { padding: 16px 32px; background: #f8f9fa; font-size: 12px; color: #999; text-align: center; }
h2 { font-size: 16px; color: #1a1a2e; margin: 24px 0 8px 0; }
</style></head><body><div class="container">`)

	// Header
	b.WriteString(`<div class="header">`)
	fmt.Fprintf(&b, `<h1>%s</h1>`, h.config.subject)
	fmt.Fprintf(&b, `<p>Pipeline Run: %s &bull; %s &bull; %d items</p>`,
		h.pipelineRunID[:12], time.Now().UTC().Format("2006-01-02 15:04 UTC"), len(entries))
	b.WriteString(`</div><div class="content">`)

	// Summary table
	b.WriteString(`<h2>Summary</h2><table><tr><th>Type</th><th>Count</th></tr>`)
	sortedTypes := sortedKeys(typeCounts)
	for _, t := range sortedTypes {
		fmt.Fprintf(&b, `<tr><td>%s</td><td>%d</td></tr>`, t, typeCounts[t])
	}
	b.WriteString(`</table>`)

	// Vulnerabilities section
	var vulns []itemEntry
	for _, e := range entries {
		if e.Type == "vulnerability" {
			vulns = append(vulns, e)
		}
	}
	if len(vulns) > 0 {
		b.WriteString(`<h2>Vulnerabilities</h2><table><tr><th>Name</th><th>Severity</th><th>URL</th></tr>`)
		for _, v := range vulns {
			name := strVal(v.Data, "name")
			severity := strVal(v.Data, "severity")
			url := strVal(v.Data, "url")
			badgeClass := "badge-info"
			switch strings.ToLower(severity) {
			case "critical":
				badgeClass = "badge-critical"
			case "high":
				badgeClass = "badge-high"
			case "medium":
				badgeClass = "badge-medium"
			case "low":
				badgeClass = "badge-low"
			}
			fmt.Fprintf(&b, `<tr><td>%s</td><td><span class="badge %s">%s</span></td><td>%s</td></tr>`,
				name, badgeClass, severity, url)
		}
		b.WriteString(`</table>`)
	}

	// Services section
	var services []itemEntry
	for _, e := range entries {
		if e.Type == "service" {
			services = append(services, e)
		}
	}
	if len(services) > 0 {
		b.WriteString(`<h2>Services</h2><table><tr><th>Host</th><th>Port</th><th>Service</th><th>Product</th></tr>`)
		for _, s := range services {
			ip := strVal(s.Data, "ip")
			port := strVal(s.Data, "port")
			svcName := strVal(s.Data, "service_name")
			product := strVal(s.Data, "product")
			fmt.Fprintf(&b, `<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				ip, port, svcName, product)
		}
		b.WriteString(`</table>`)
	}

	b.WriteString(`</div>`)
	b.WriteString(`<div class="footer">Generated by Portwhine</div>`)
	b.WriteString(`</div></body></html>`)

	return b.String()
}

func (h *workerHandler) buildTextReport(entries []itemEntry, typeCounts map[string]int) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s\n", h.config.subject)
	fmt.Fprintf(&b, "Pipeline Run: %s\n", h.pipelineRunID)
	fmt.Fprintf(&b, "Date: %s\n", time.Now().UTC().Format("2006-01-02 15:04 UTC"))
	fmt.Fprintf(&b, "Total Items: %d\n\n", len(entries))

	b.WriteString("=== Summary ===\n")
	for _, t := range sortedKeys(typeCounts) {
		fmt.Fprintf(&b, "  %-20s %d\n", t, typeCounts[t])
	}
	b.WriteString("\n")

	// Vulnerabilities
	var vulns []itemEntry
	for _, e := range entries {
		if e.Type == "vulnerability" {
			vulns = append(vulns, e)
		}
	}
	if len(vulns) > 0 {
		b.WriteString("=== Vulnerabilities ===\n")
		for _, v := range vulns {
			fmt.Fprintf(&b, "  [%s] %s - %s\n",
				strVal(v.Data, "severity"), strVal(v.Data, "name"), strVal(v.Data, "url"))
		}
		b.WriteString("\n")
	}

	// Services
	var services []itemEntry
	for _, e := range entries {
		if e.Type == "service" {
			services = append(services, e)
		}
	}
	if len(services) > 0 {
		b.WriteString("=== Services ===\n")
		for _, s := range services {
			fmt.Fprintf(&b, "  %s:%s %s %s\n",
				strVal(s.Data, "ip"), strVal(s.Data, "port"),
				strVal(s.Data, "service_name"), strVal(s.Data, "product"))
		}
		b.WriteString("\n")
	}

	b.WriteString("-- Generated by Portwhine --\n")
	return b.String()
}

func (h *workerHandler) sendSMTP(body string) error {
	addr := fmt.Sprintf("%s:%d", h.config.smtpHost, h.config.smtpPort)

	contentType := "text/plain"
	if h.config.format == "html" {
		contentType = "text/html"
	}

	from := h.config.smtpUser
	if from == "" {
		from = "portwhine@localhost"
	}

	for _, recipient := range h.config.recipients {
		msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s; charset=UTF-8\r\n\r\n%s",
			from, recipient, h.config.subject, contentType, body)

		var auth smtp.Auth
		if h.config.smtpUser != "" && h.config.smtpPass != "" {
			auth = smtp.PlainAuth("", h.config.smtpUser, h.config.smtpPass, h.config.smtpHost)
		}

		if h.config.smtpPort == 465 {
			// Implicit TLS (SMTPS).
			if err := h.sendWithImplicitTLS(addr, auth, from, recipient, []byte(msg)); err != nil {
				return fmt.Errorf("send to %s: %w", recipient, err)
			}
		} else {
			// Standard SMTP with optional STARTTLS.
			if err := smtp.SendMail(addr, auth, from, []string{recipient}, []byte(msg)); err != nil {
				return fmt.Errorf("send to %s: %w", recipient, err)
			}
		}

		slog.Info("email sent", "recipient", recipient)
	}

	return nil
}

func (h *workerHandler) sendWithImplicitTLS(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	tlsConn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: h.config.smtpHost})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer tlsConn.Close()

	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(tlsConn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	return client.Quit()
}

func (h *workerHandler) emitDelivery(stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], itemCount int, status string) {
	deliveryData, _ := structpb.NewStruct(map[string]interface{}{
		"recipients":  toInterfaceSlice(h.config.recipients),
		"subject":     h.config.subject,
		"items_count": float64(itemCount),
		"format":      h.config.format,
		"status":      status,
	})

	item := &portwhinev1.DataItem{
		Id:            uuid.New().String(),
		PipelineRunId: h.pipelineRunID,
		Type:          "email_delivery",
		Data:          deliveryData,
		Metadata: &portwhinev1.DataItemMetadata{
			Source:    "email-output",
			CreatedAt: timestamppb.Now(),
			NodeId:    h.nodeID,
		},
	}

	h.safeSend(stream, &workerv1.ProcessResponse{
		Payload: &workerv1.ProcessResponse_Item{Item: item},
	})
	h.itemsProduced.Add(1)
}

func (h *workerHandler) Shutdown(_ context.Context, _ *connect.Request[workerv1.ShutdownRequest]) (*connect.Response[workerv1.ShutdownResponse], error) {
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_STOPPED
	return connect.NewResponse(&workerv1.ShutdownResponse{
		TotalItemsProcessed: h.itemsProcessed.Load(),
		TotalItemsProduced:  h.itemsProduced.Load(),
	}), nil
}

func (h *workerHandler) HealthCheck(_ context.Context, _ *connect.Request[workerv1.HealthCheckRequest]) (*connect.Response[workerv1.HealthCheckResponse], error) {
	return connect.NewResponse(&workerv1.HealthCheckResponse{
		Status: h.status,
	}), nil
}

func (h *workerHandler) safeSend(stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], resp *workerv1.ProcessResponse) {
	h.streamMu.Lock()
	defer h.streamMu.Unlock()
	if err := stream.Send(resp); err != nil {
		slog.Error("failed to send response", "error", err)
	}
}

func (h *workerHandler) sendHeartbeats(stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], done <-chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			h.safeSend(stream, &workerv1.ProcessResponse{
				Payload: &workerv1.ProcessResponse_Heartbeat{
					Heartbeat: &portwhinev1.WorkerHeartbeat{
						WorkerId:       h.nodeID,
						Status:         h.status,
						ItemsProcessed: h.itemsProcessed.Load(),
						ItemsProduced:  h.itemsProduced.Load(),
						ErrorsCount:    h.errorsCount.Load(),
						Timestamp:      timestamppb.Now(),
					},
				},
			})
		}
	}
}

// Helper functions

func strVal(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func toInterfaceSlice(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}
