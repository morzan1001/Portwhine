package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
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

type workerConfig struct {
	server   string
	timeout  int
	extraArgs string
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
			Name:               "whois-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"domain", "ip_address"},
			OutputTypes:        []string{"whois_result"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"server": {"type": "string", "description": "Specific WHOIS server to query"},
					"timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 30},
					"extra_args": {"type": "string", "description": "Additional whois CLI arguments"}
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
		timeout: 30,
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["server"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.server = s
			}
		}

		if v, ok := fields["timeout"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.timeout = int(n)
			}
		}

		if v, ok := fields["extra_args"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.extraArgs = s
			}
		}
	}

	h.config = cfg
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("whois worker initialized",
		"server", cfg.server,
		"timeout", cfg.timeout,
		"extra_args", cfg.extraArgs,
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
			h.processItem(ctx, stream, payload.Item)
		case *workerv1.ProcessRequest_Flush:
			// No buffering; no-op
		case *workerv1.ProcessRequest_Cancel:
			slog.Info("cancel received", "reason", payload.Cancel.GetReason())
			return nil
		}
	}

	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	return nil
}

func (h *workerHandler) processItem(ctx context.Context, stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], item *portwhinev1.DataItem) {
	h.itemsProcessed.Add(1)

	var target string
	var targetType string

	switch item.GetType() {
	case "domain":
		if item.GetData() != nil {
			if v, ok := item.GetData().GetFields()["domain"]; ok {
				target = v.GetStringValue()
			}
		}
		targetType = "domain"

	case "ip_address":
		if item.GetData() != nil {
			if v, ok := item.GetData().GetFields()["ip"]; ok {
				target = v.GetStringValue()
			}
		}
		targetType = "ip_address"

	default:
		h.sendError(stream, item.GetId(), fmt.Sprintf("unsupported item type: %s", item.GetType()), false)
		return
	}

	if target == "" {
		h.sendError(stream, item.GetId(), fmt.Sprintf("missing target in %s item", item.GetType()), false)
		return
	}

	// Build whois command arguments
	args := []string{}
	if h.config.server != "" {
		args = append(args, "-h", h.config.server)
	}
	if h.config.extraArgs != "" {
		extraParts := strings.Fields(h.config.extraArgs)
		args = append(args, extraParts...)
	}
	args = append(args, target)

	slog.Debug("running whois", "args", args, "target", target)

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "whois", args...)
	output, err := cmd.Output()
	if err != nil {
		// whois may exit with non-zero even on partial success; check if we got output
		if len(output) == 0 {
			h.sendError(stream, item.GetId(), fmt.Sprintf("whois execution failed for %s: %v", target, err), true)
			return
		}
		slog.Warn("whois exited with error but produced output", "target", target, "error", err)
	}

	rawOutput := string(output)
	parsed := parseWhoisOutput(rawOutput)
	analysis := analyzeWhois(parsed)

	// Build data fields for the whois_result DataItem
	dataFields := map[string]*structpb.Value{
		"target":      structpb.NewStringValue(target),
		"target_type": structpb.NewStringValue(targetType),
		"raw_output":  structpb.NewStringValue(rawOutput),
	}

	// Add all parsed fields as additional string fields
	for key, value := range parsed {
		dataFields[key] = structpb.NewStringValue(value)
	}

	// Add analysis fields
	for key, value := range analysis {
		dataFields[key] = structpb.NewStringValue(value)
	}

	labels := map[string]string{
		"worker_type": "whois",
		"target_type": targetType,
	}
	if status, ok := analysis["expiry_status"]; ok {
		labels["expiry_status"] = status
	}

	resultItem := &portwhinev1.DataItem{
		Id:            uuid.New().String(),
		PipelineRunId: h.pipelineRunID,
		Type:          "whois_result",
		Data:          &structpb.Struct{Fields: dataFields},
		Metadata: &portwhinev1.DataItemMetadata{
			Source:    "whois-worker",
			CreatedAt: timestamppb.Now(),
			NodeId:    h.nodeID,
			Labels:    labels,
		},
		ParentIds: []string{item.GetId()},
	}

	h.safeSend(stream, &workerv1.ProcessResponse{
		Payload: &workerv1.ProcessResponse_Item{Item: resultItem},
	})
	h.itemsProduced.Add(1)
}

func parseWhoisOutput(output string) map[string]string {
	result := make(map[string]string)
	// Common WHOIS fields to extract
	fieldMappings := map[string]string{
		"domain name":              "domain_name",
		"registrar":                "registrar",
		"registrant":               "registrant",
		"creation date":            "creation_date",
		"updated date":             "updated_date",
		"expiration date":          "expiration_date",
		"registry expiry date":     "expiration_date",
		"name server":              "name_servers",
		"dnssec":                   "dnssec",
		"registrant organization":  "registrant_org",
		"registrant country":       "registrant_country",
		"admin email":              "admin_email",
		"tech email":               "tech_email",
		"status":                   "status",
		"domain status":            "status",
		// IP WHOIS fields
		"netname":      "netname",
		"netrange":     "netrange",
		"cidr":         "cidr",
		"orgname":      "org_name",
		"organization": "org_name",
		"country":      "country",
		"descr":        "description",
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])
		if mapped, ok := fieldMappings[key]; ok {
			if existing, exists := result[mapped]; exists {
				result[mapped] = existing + ", " + value
			} else {
				result[mapped] = value
			}
		}
	}
	return result
}

func analyzeWhois(parsed map[string]string) map[string]string {
	analysis := make(map[string]string)

	// Parse expiration date and calculate days until expiry
	if expStr, ok := parsed["expiration_date"]; ok && expStr != "" {
		if expTime, err := parseWhoisDate(expStr); err == nil {
			daysUntilExpiry := int(time.Until(expTime).Hours() / 24)
			analysis["days_until_expiry"] = strconv.Itoa(daysUntilExpiry)

			if daysUntilExpiry < 0 {
				analysis["expiry_status"] = "expired"
			} else if daysUntilExpiry <= 30 {
				analysis["expiry_status"] = "critical"
			} else if daysUntilExpiry <= 90 {
				analysis["expiry_status"] = "warning"
			} else {
				analysis["expiry_status"] = "ok"
			}
		}
	}

	// Parse creation date and calculate domain age
	if createStr, ok := parsed["creation_date"]; ok && createStr != "" {
		if createTime, err := parseWhoisDate(createStr); err == nil {
			ageDays := int(time.Since(createTime).Hours() / 24)
			analysis["domain_age_days"] = strconv.Itoa(ageDays)

			if ageDays < 30 {
				analysis["age_risk"] = "high"
			} else if ageDays < 365 {
				analysis["age_risk"] = "medium"
			} else {
				analysis["age_risk"] = "low"
			}
		}
	}

	// Check DNSSEC
	if dnssec, ok := parsed["dnssec"]; ok {
		lower := strings.ToLower(dnssec)
		if strings.Contains(lower, "unsigned") || strings.Contains(lower, "no") {
			analysis["dnssec_status"] = "unsigned"
		} else if strings.Contains(lower, "signed") || strings.Contains(lower, "yes") {
			analysis["dnssec_status"] = "signed"
		}
	}

	return analysis
}

func parseWhoisDate(dateStr string) (time.Time, error) {
	// Try common WHOIS date formats
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02-Jan-2006",
		"January 02 2006",
		"2006/01/02",
		"02/01/2006",
		"20060102",
	}

	// Clean the date string
	cleaned := strings.TrimSpace(dateStr)
	// Handle "before" or "until" prefixes sometimes in WHOIS
	for _, prefix := range []string{"before ", "until "} {
		cleaned = strings.TrimPrefix(cleaned, prefix)
	}

	for _, format := range formats {
		if t, err := time.Parse(format, cleaned); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
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
