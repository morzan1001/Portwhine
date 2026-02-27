package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	neturl "net/url"
	"os"
	"os/exec"
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

type humbleOutput struct {
	MissingHeaders []string        `json:"missing_headers"`
	Findings       []humbleFinding `json:"findings"`
}

type humbleFinding struct {
	Header   string `json:"header"`
	Issue    string `json:"issue"`
	Severity string `json:"severity"`
}

type workerConfig struct {
	extraArgs  string
	skipChecks []string
	headers    map[string]string
	timeout    int
	brief      bool
	userAgent  string
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
			Name:               "humble-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"url"},
			OutputTypes:        []string{"http_headers"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"extra_args": {"type": "string", "description": "Additional command-line arguments for humble"},
					"skip_checks": {"type": "array", "items": {"type": "string"}, "description": "Checks to skip (--skip per entry)"},
					"headers": {"type": "object", "description": "Custom HTTP headers to send"},
					"timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 120},
					"brief": {"type": "boolean", "description": "Brief output mode (-b)", "default": false},
					"user_agent": {"type": "string", "description": "Custom User-Agent string"}
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
		timeout: 120,
		headers: make(map[string]string),
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["extra_args"]; ok {
			cfg.extraArgs = v.GetStringValue()
		}

		if v, ok := fields["skip_checks"]; ok {
			if list := v.GetListValue(); list != nil {
				for _, item := range list.GetValues() {
					if s := item.GetStringValue(); s != "" {
						cfg.skipChecks = append(cfg.skipChecks, s)
					}
				}
			}
		}

		if v, ok := fields["headers"]; ok {
			if s := v.GetStructValue(); s != nil {
				for k, val := range s.GetFields() {
					cfg.headers[k] = val.GetStringValue()
				}
			}
		}

		if v, ok := fields["timeout"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.timeout = int(n)
			}
		}

		if v, ok := fields["brief"]; ok {
			cfg.brief = v.GetBoolValue()
		}

		if v, ok := fields["user_agent"]; ok {
			cfg.userAgent = v.GetStringValue()
		}
	}

	h.config = cfg
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("humble worker initialized",
		"extra_args", cfg.extraArgs,
		"timeout", cfg.timeout,
		"brief", cfg.brief,
		"skip_checks", cfg.skipChecks,
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

	targetURL := ""
	if item.GetData() != nil {
		if v, ok := item.GetData().GetFields()["url"]; ok {
			targetURL = v.GetStringValue()
		}
	}
	if targetURL == "" {
		h.sendError(stream, item.GetId(), "missing 'url' field in data", false)
		return
	}

	// Validate URL
	parsed, err := neturl.Parse(targetURL)
	if err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("invalid URL %q: %v", targetURL, err), false)
		return
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		h.sendError(stream, item.GetId(), fmt.Sprintf("URL must have http or https scheme, got %q", parsed.Scheme), false)
		return
	}

	slog.Info("running humble", "url", targetURL)

	// Build command arguments
	args := []string{"-m", "humble", "-u", targetURL, "-o", "json"}

	if h.config.brief {
		args = append(args, "-b")
	}

	for _, skip := range h.config.skipChecks {
		args = append(args, "--skip", skip)
	}

	for key, value := range h.config.headers {
		args = append(args, "-H", fmt.Sprintf("%s: %s", key, value))
	}
	if h.config.userAgent != "" {
		args = append(args, "-ua", h.config.userAgent)
	}

	if h.config.extraArgs != "" {
		extraParts := strings.Fields(h.config.extraArgs)
		args = append(args, extraParts...)
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "python3", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("humble command failed, trying without -o json",
			"error", err,
			"output", string(output),
		)

		// Fallback: run without -o json
		fallbackArgs := []string{"-m", "humble", "-u", targetURL}

		if h.config.brief {
			fallbackArgs = append(fallbackArgs, "-b")
		}

		for _, skip := range h.config.skipChecks {
			fallbackArgs = append(fallbackArgs, "--skip", skip)
		}

		for key, value := range h.config.headers {
			fallbackArgs = append(fallbackArgs, "-H", fmt.Sprintf("%s: %s", key, value))
		}
		if h.config.userAgent != "" {
			fallbackArgs = append(fallbackArgs, "-ua", h.config.userAgent)
		}

		if h.config.extraArgs != "" {
			extraParts := strings.Fields(h.config.extraArgs)
			fallbackArgs = append(fallbackArgs, extraParts...)
		}

		fallbackCtx, fallbackCancel := context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
		defer fallbackCancel()

		cmd = exec.CommandContext(fallbackCtx, "python3", fallbackArgs...)
		output, err = cmd.CombinedOutput()
		if err != nil {
			h.sendError(stream, item.GetId(), fmt.Sprintf("humble failed for %s: %v (output: %s)", targetURL, err, string(output)), true)
			return
		}
	}

	// Try parsing as JSON first
	var parsedOutput humbleOutput
	missingHeaders := []interface{}{}
	findings := []interface{}{}
	grade := ""

	if jsonErr := json.Unmarshal(output, &parsedOutput); jsonErr == nil {
		for _, header := range parsedOutput.MissingHeaders {
			missingHeaders = append(missingHeaders, header)
		}
		for _, finding := range parsedOutput.Findings {
			findings = append(findings, fmt.Sprintf("[%s] %s: %s", finding.Severity, finding.Header, finding.Issue))
		}
		grade = computeGradeWeighted(parsedOutput.Findings, len(parsedOutput.MissingHeaders))
	} else {
		// Fallback: parse text output
		slog.Debug("JSON parsing failed, falling back to text parsing", "error", jsonErr)
		missingHeaders, findings, grade = parseTextOutput(string(output))
	}

	// Build list values
	missingList, _ := structpb.NewList(missingHeaders)
	findingsList, _ := structpb.NewList(findings)

	resultItem := &portwhinev1.DataItem{
		Id:            uuid.New().String(),
		PipelineRunId: h.pipelineRunID,
		Type:          "http_headers",
		Data: &structpb.Struct{Fields: map[string]*structpb.Value{
			"url":             structpb.NewStringValue(targetURL),
			"missing_headers": structpb.NewListValue(missingList),
			"findings":        structpb.NewListValue(findingsList),
			"grade":           structpb.NewStringValue(grade),
		}},
		Metadata: &portwhinev1.DataItemMetadata{
			Source:    "humble-worker",
			CreatedAt: timestamppb.Now(),
			NodeId:    h.nodeID,
			Labels: map[string]string{
				"worker_type": "humble",
				"grade":       grade,
			},
		},
		ParentIds: []string{item.GetId()},
	}

	h.safeSend(stream, &workerv1.ProcessResponse{
		Payload: &workerv1.ProcessResponse_Item{Item: resultItem},
	})
	h.itemsProduced.Add(1)
}

func parseTextOutput(output string) (missingHeaders []interface{}, findings []interface{}, grade string) {
	lines := strings.Split(output, "\n")

	inMissing := false
	inDeprecated := false
	inFindings := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		// Detect section headers
		if strings.Contains(lower, "missing headers") {
			inMissing = true
			inDeprecated = false
			inFindings = false
			continue
		}
		if strings.Contains(lower, "deprecated headers") || strings.Contains(lower, "fingerprint headers") {
			inMissing = false
			inDeprecated = true
			inFindings = false
			continue
		}
		if strings.Contains(lower, "findings") || strings.Contains(lower, "analysis") || strings.Contains(lower, "issues") {
			inMissing = false
			inDeprecated = false
			inFindings = true
			continue
		}

		// Skip empty lines and decorative lines
		if trimmed == "" || strings.HasPrefix(trimmed, "=") || strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "[") && strings.Contains(trimmed, "humble") {
			continue
		}

		if inMissing && trimmed != "" {
			// Extract header name - lines typically contain header names
			header := strings.TrimPrefix(trimmed, "- ")
			header = strings.TrimPrefix(header, "* ")
			if header != "" {
				missingHeaders = append(missingHeaders, header)
			}
		}

		if inDeprecated && trimmed != "" {
			findings = append(findings, fmt.Sprintf("[deprecated] %s", trimmed))
		}

		if inFindings && trimmed != "" {
			findings = append(findings, trimmed)
		}
	}

	grade = computeGrade(len(missingHeaders), len(findings))
	return
}

func computeGradeWeighted(findings []humbleFinding, missingCount int) string {
	score := 0
	for _, f := range findings {
		switch strings.ToLower(f.Severity) {
		case "critical":
			score += 10
		case "high":
			score += 5
		case "medium":
			score += 3
		case "low", "warning":
			score += 1
		default:
			score += 1
		}
	}
	score += missingCount

	switch {
	case score == 0:
		return "A"
	case score <= 5:
		return "B"
	case score <= 15:
		return "C"
	case score <= 30:
		return "D"
	default:
		return "F"
	}
}

func computeGrade(missingCount, findingsCount int) string {
	total := missingCount + findingsCount
	switch {
	case total == 0:
		return "A"
	case total <= 2:
		return "B"
	case total <= 5:
		return "C"
	case total <= 8:
		return "D"
	default:
		return "F"
	}
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
