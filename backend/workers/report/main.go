package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
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

type reportEntry struct {
	ID        string
	Type      string
	Data      map[string]interface{}
	Source    string
	CreatedAt string
	Labels    map[string]string
	ParentIDs []string
}

type workerConfig struct {
	format     string
	includeRaw bool
	groupBy    string
	title      string
}

type workerHandler struct {
	workerv1connect.UnimplementedWorkerServiceHandler

	mu             sync.Mutex
	streamMu       sync.Mutex
	nodeID         string
	pipelineRunID  string
	config         *workerConfig
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
			Name:               "report-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"service", "url", "vulnerability", "ssl_result", "http_headers", "web_technology", "screenshot", "ssh_audit_result", "whois_result", "ip_address", "dns_record", "domain"},
			OutputTypes:        []string{"report"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"format": {"type": "string", "description": "Output format: json or summary", "default": "json"},
					"include_raw": {"type": "boolean", "description": "Include raw payloads in report", "default": false},
					"group_by": {"type": "string", "description": "How to group results: type, host, or severity", "default": "type"},
					"title": {"type": "string", "description": "Report title", "default": "Portwhine Scan Report"}
				}
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

	cfg := &workerConfig{
		format:     "json",
		includeRaw: false,
		groupBy:    "type",
		title:      "Portwhine Scan Report",
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["format"]; ok {
			if s := v.GetStringValue(); s == "json" || s == "summary" {
				cfg.format = s
			}
		}

		if v, ok := fields["include_raw"]; ok {
			cfg.includeRaw = v.GetBoolValue()
		}

		if v, ok := fields["group_by"]; ok {
			if s := v.GetStringValue(); s == "type" || s == "host" || s == "severity" {
				cfg.groupBy = s
			}
		}

		if v, ok := fields["title"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.title = s
			}
		}
	}

	h.config = cfg
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("report worker initialized",
		"format", cfg.format,
		"include_raw", cfg.includeRaw,
		"group_by", cfg.groupBy,
		"title", cfg.title,
	)

	return connect.NewResponse(&workerv1.InitializeResponse{Success: true}), nil
}

func (h *workerHandler) Process(ctx context.Context, stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse]) error {
	if !h.initialized {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("not initialized"))
	}
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_PROCESSING

	// Start heartbeat goroutine
	heartbeatDone := make(chan struct{})
	go h.sendHeartbeats(stream, heartbeatDone)
	defer close(heartbeatDone)

	var entries []reportEntry

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
			slog.Info("flush received, generating report", "buffered_count", len(entries))
			if len(entries) > 0 {
				h.generateReport(ctx, stream, entries)
				entries = nil
			}
		case *workerv1.ProcessRequest_Cancel:
			slog.Info("cancel received", "reason", payload.Cancel.GetReason())
			return nil
		}
	}

	// Generate report from remaining items on EOF
	if len(entries) > 0 {
		slog.Info("EOF reached, generating report from remaining items", "buffered_count", len(entries))
		h.generateReport(ctx, stream, entries)
	}

	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	return nil
}

func convertToEntry(item *portwhinev1.DataItem) reportEntry {
	entry := reportEntry{
		ID:        item.GetId(),
		Type:      item.GetType(),
		Data:      make(map[string]interface{}),
		ParentIDs: item.GetParentIds(),
	}
	if item.GetData() != nil {
		for k, v := range item.GetData().GetFields() {
			entry.Data[k] = v.AsInterface()
		}
	}
	if item.GetMetadata() != nil {
		entry.Source = item.GetMetadata().GetSource()
		if item.GetMetadata().GetCreatedAt() != nil {
			entry.CreatedAt = item.GetMetadata().GetCreatedAt().AsTime().Format(time.RFC3339)
		}
		entry.Labels = item.GetMetadata().GetLabels()
	}
	return entry
}

func (h *workerHandler) generateReport(_ context.Context, stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], entries []reportEntry) {
	// Count items by type
	itemCounts := make(map[string]interface{})
	for _, entry := range entries {
		if count, ok := itemCounts[entry.Type]; ok {
			itemCounts[entry.Type] = count.(float64) + 1
		} else {
			itemCounts[entry.Type] = float64(1)
		}
	}

	// Collect unique hosts
	hosts := make(map[string]bool)
	for _, entry := range entries {
		if ip, ok := entry.Data["ip"]; ok {
			if s, ok := ip.(string); ok && s != "" {
				hosts[s] = true
			}
		}
		if hostname, ok := entry.Data["hostname"]; ok {
			if s, ok := hostname.(string); ok && s != "" {
				hosts[s] = true
			}
		}
		if domain, ok := entry.Data["domain"]; ok {
			if s, ok := domain.(string); ok && s != "" {
				hosts[s] = true
			}
		}
		if u, ok := entry.Data["url"]; ok {
			if s, ok := u.(string); ok && s != "" {
				if parsed, err := url.Parse(s); err == nil && parsed.Hostname() != "" {
					hosts[parsed.Hostname()] = true
				}
			}
		}
	}
	hostList := make([]interface{}, 0, len(hosts))
	for host := range hosts {
		hostList = append(hostList, host)
	}

	// Vulnerability severity counts
	vulnSeverities := make(map[string]interface{})
	for _, entry := range entries {
		if entry.Type == "vulnerability" {
			severity := "unknown"
			if s, ok := entry.Data["severity"]; ok {
				if str, ok := s.(string); ok && str != "" {
					severity = strings.ToLower(str)
				}
			}
			if count, ok := vulnSeverities[severity]; ok {
				vulnSeverities[severity] = count.(float64) + 1
			} else {
				vulnSeverities[severity] = float64(1)
			}
		}
	}

	// Collect discovered services
	var services []interface{}
	for _, entry := range entries {
		if entry.Type == "service" {
			svc := make(map[string]interface{})
			if ip, ok := entry.Data["ip"]; ok {
				svc["ip"] = ip
			}
			if port, ok := entry.Data["port"]; ok {
				svc["port"] = port
			}
			if protocol, ok := entry.Data["protocol"]; ok {
				svc["protocol"] = protocol
			}
			if name, ok := entry.Data["service_name"]; ok {
				svc["service_name"] = name
			}
			services = append(services, svc)
		}
	}

	// Group entries based on configured strategy
	grouped := h.groupEntries(entries)

	// Build summary
	summary := map[string]interface{}{
		"hosts":   hostList,
		"grouped": grouped,
	}
	if len(vulnSeverities) > 0 {
		summary["vulnerability_severities"] = vulnSeverities
	}
	if len(services) > 0 {
		summary["services_discovered"] = services
	}

	// Build the report data map
	reportData := map[string]interface{}{
		"title":        h.config.title,
		"format":       h.config.format,
		"total_items":  float64(len(entries)),
		"item_counts":  itemCounts,
		"summary":      summary,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	}

	// Determine overall risk level based on vulnerability severities
	overallRisk := "low"
	{
		totalVulns := float64(0)
		hasCriticalHigh := false
		for sev, count := range vulnSeverities {
			c := count.(float64)
			totalVulns += c
			if sev == "critical" || sev == "high" {
				hasCriticalHigh = true
			}
		}
		if hasCriticalHigh {
			overallRisk = "critical"
		} else if totalVulns > 10 {
			overallRisk = "high"
		} else if totalVulns > 0 {
			overallRisk = "medium"
		}
	}
	reportData["overall_risk"] = overallRisk

	// Add text summary for "summary" format
	if h.config.format == "summary" {
		reportData["text_summary"] = buildTextSummary(entries, hostList, itemCounts, vulnSeverities, services)
	}

	// Include raw entries if configured
	if h.config.includeRaw {
		rawEntries := make([]interface{}, 0, len(entries))
		for _, entry := range entries {
			rawEntry := map[string]interface{}{
				"id":         entry.ID,
				"type":       entry.Type,
				"data":       entry.Data,
				"source":     entry.Source,
				"created_at": entry.CreatedAt,
				"parent_ids": toInterfaceSlice(entry.ParentIDs),
			}
			if entry.Labels != nil {
				labels := make(map[string]interface{})
				for k, v := range entry.Labels {
					labels[k] = v
				}
				rawEntry["labels"] = labels
			}
			rawEntries = append(rawEntries, rawEntry)
		}
		reportData["entries"] = rawEntries
	}

	// Convert to protobuf struct
	dataStruct, err := structpb.NewValue(reportData)
	if err != nil {
		slog.Error("failed to convert report data to protobuf value", "error", err)
		h.sendError(stream, "", "failed to build report: "+err.Error(), false)
		return
	}

	reportItem := &portwhinev1.DataItem{
		Id:            uuid.New().String(),
		PipelineRunId: h.pipelineRunID,
		Type:          "report",
		Data:          dataStruct.GetStructValue(),
		Metadata: &portwhinev1.DataItemMetadata{
			Source:    "report-worker",
			CreatedAt: timestamppb.Now(),
			NodeId:    h.nodeID,
			Labels: map[string]string{
				"worker_type": "report",
				"format":      h.config.format,
				"group_by":    h.config.groupBy,
			},
		},
	}

	h.safeSend(stream, &workerv1.ProcessResponse{
		Payload: &workerv1.ProcessResponse_Item{Item: reportItem},
	})
	h.itemsProduced.Add(1)

	slog.Info("report generated",
		"total_items", len(entries),
		"format", h.config.format,
		"group_by", h.config.groupBy,
		"report_id", reportItem.Id,
	)
}

func (h *workerHandler) groupEntries(entries []reportEntry) map[string]interface{} {
	switch h.config.groupBy {
	case "host":
		return h.groupByHost(entries)
	case "severity":
		return h.groupBySeverity(entries)
	default:
		return h.groupByType(entries)
	}
}

func (h *workerHandler) groupByType(entries []reportEntry) map[string]interface{} {
	groups := make(map[string][]interface{})
	for _, entry := range entries {
		simplified := map[string]interface{}{
			"id":     entry.ID,
			"type":   entry.Type,
			"data":   entry.Data,
			"source": entry.Source,
		}
		groups[entry.Type] = append(groups[entry.Type], simplified)
	}

	result := make(map[string]interface{})
	for k, v := range groups {
		result[k] = v
	}
	return result
}

func (h *workerHandler) groupByHost(entries []reportEntry) map[string]interface{} {
	groups := make(map[string][]interface{})
	for _, entry := range entries {
		host := extractHost(entry)
		simplified := map[string]interface{}{
			"id":     entry.ID,
			"type":   entry.Type,
			"data":   entry.Data,
			"source": entry.Source,
		}
		groups[host] = append(groups[host], simplified)
	}

	result := make(map[string]interface{})
	for k, v := range groups {
		result[k] = v
	}
	return result
}

func (h *workerHandler) groupBySeverity(entries []reportEntry) map[string]interface{} {
	groups := make(map[string][]interface{})
	for _, entry := range entries {
		severity := "info"
		if entry.Type == "vulnerability" {
			if s, ok := entry.Data["severity"]; ok {
				if str, ok := s.(string); ok && str != "" {
					severity = strings.ToLower(str)
				}
			}
		}
		simplified := map[string]interface{}{
			"id":     entry.ID,
			"type":   entry.Type,
			"data":   entry.Data,
			"source": entry.Source,
		}
		groups[severity] = append(groups[severity], simplified)
	}

	result := make(map[string]interface{})
	for k, v := range groups {
		result[k] = v
	}
	return result
}

func extractHost(entry reportEntry) string {
	if ip, ok := entry.Data["ip"]; ok {
		if s, ok := ip.(string); ok && s != "" {
			return s
		}
	}
	if hostname, ok := entry.Data["hostname"]; ok {
		if s, ok := hostname.(string); ok && s != "" {
			return s
		}
	}
	if domain, ok := entry.Data["domain"]; ok {
		if s, ok := domain.(string); ok && s != "" {
			return s
		}
	}
	if u, ok := entry.Data["url"]; ok {
		if s, ok := u.(string); ok && s != "" {
			if parsed, err := url.Parse(s); err == nil && parsed.Hostname() != "" {
				return parsed.Hostname()
			}
		}
	}
	return "unknown"
}

func buildTextSummary(entries []reportEntry, hosts []interface{}, itemCounts map[string]interface{}, vulnSeverities map[string]interface{}, services []interface{}) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Scan completed: %d items collected across %d hosts.\n\n", len(entries), len(hosts)))

	// Services summary
	if len(services) > 0 {
		sb.WriteString(fmt.Sprintf("Services discovered: %d\n", len(services)))
		for _, svc := range services {
			if m, ok := svc.(map[string]interface{}); ok {
				name := ""
				port := ""
				if n, ok := m["service_name"]; ok {
					name = fmt.Sprintf("%v", n)
				}
				if p, ok := m["port"]; ok {
					port = fmt.Sprintf("%v", p)
				}
				if name != "" {
					sb.WriteString(fmt.Sprintf("  - %s (port %s)\n", name, port))
				}
			}
		}
		sb.WriteString("\n")
	}

	// Vulnerability summary
	if len(vulnSeverities) > 0 {
		sb.WriteString("Vulnerabilities found:\n")
		for sev, count := range vulnSeverities {
			sb.WriteString(fmt.Sprintf("  - %s: %.0f\n", strings.ToUpper(sev), count))
		}
		sb.WriteString("\n")
	}

	// Item type breakdown
	sb.WriteString("Item breakdown:\n")
	for itemType, count := range itemCounts {
		sb.WriteString(fmt.Sprintf("  - %s: %.0f\n", itemType, count))
	}
	sb.WriteString("\n")

	// Overall risk assessment
	totalVulns := float64(0)
	criticalHigh := float64(0)
	for sev, count := range vulnSeverities {
		c := count.(float64)
		totalVulns += c
		if sev == "critical" || sev == "high" {
			criticalHigh += c
		}
	}

	if totalVulns == 0 {
		sb.WriteString("Risk assessment: LOW - No vulnerabilities detected.\n")
	} else if criticalHigh > 0 {
		sb.WriteString(fmt.Sprintf("Risk assessment: CRITICAL - %.0f critical/high severity vulnerabilities require immediate attention.\n", criticalHigh))
	} else if totalVulns > 10 {
		sb.WriteString(fmt.Sprintf("Risk assessment: HIGH - %.0f vulnerabilities found, review recommended.\n", totalVulns))
	} else {
		sb.WriteString(fmt.Sprintf("Risk assessment: MEDIUM - %.0f vulnerabilities found.\n", totalVulns))
	}

	// Grades from other workers
	grades := make(map[string]string)
	for _, entry := range entries {
		if g, ok := entry.Data["grade"]; ok {
			if gs, ok := g.(string); ok && gs != "" {
				source := entry.Source
				if source == "" {
					source = entry.Type
				}
				grades[source] = gs
			}
		}
	}
	if len(grades) > 0 {
		sb.WriteString("\nWorker grades:\n")
		for source, grade := range grades {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", source, grade))
		}
	}

	return sb.String()
}

func toInterfaceSlice(ss []string) []interface{} {
	result := make([]interface{}, len(ss))
	for i, s := range ss {
		result[i] = s
	}
	return result
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

func (h *workerHandler) sendError(stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], sourceItemID, msg string, retryable bool) {
	h.errorsCount.Add(1)
	h.safeSend(stream, &workerv1.ProcessResponse{
		Payload: &workerv1.ProcessResponse_Error{
			Error: &workerv1.ProcessError{
				SourceItemId: sourceItemID,
				ErrorMessage: msg,
				Retryable:    retryable,
			},
		},
	})
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
