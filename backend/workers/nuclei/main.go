package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	neturl "net/url"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
	workerv1 "github.com/portwhine/portwhine/gen/go/portwhine/worker/v1"
	"github.com/portwhine/portwhine/gen/go/portwhine/worker/v1/workerv1connect"
	"github.com/portwhine/portwhine/pkg/dataitem"
)

type nucleiResult struct {
	TemplateID  string     `json:"template-id"`
	Info        nucleiInfo `json:"info"`
	MatchedAt   string     `json:"matched-at"`
	MatcherName string     `json:"matcher-name"`
	Type        string     `json:"type"`
	Host        string     `json:"host"`
	IP          string     `json:"ip"`
	URL         string     `json:"url"`
}

type nucleiInfo struct {
	Name        string   `json:"name"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type workerConfig struct {
	templates   string
	severity    string
	tags        string
	rateLimit   int
	concurrency int
	timeout     int
	extraArgs   string
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

	srv := &http.Server{
		Addr:              ":50051",
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("nuclei worker listening", "addr", ":50051")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func (h *workerHandler) GetCapabilities(_ context.Context, _ *connect.Request[workerv1.GetCapabilitiesRequest]) (*connect.Response[workerv1.GetCapabilitiesResponse], error) {
	return connect.NewResponse(&workerv1.GetCapabilitiesResponse{
		Capability: &portwhinev1.WorkerCapability{
			Name:               "nuclei-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"url", "service"},
			OutputTypes:        []string{"vulnerability"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"templates": {"type": "string", "description": "Specific template paths/IDs to use (comma-separated)"},
					"severity": {"type": "string", "description": "Filter by severity (comma-separated)", "default": "medium,high,critical"},
					"tags": {"type": "string", "description": "Filter templates by tags (comma-separated)"},
					"rate_limit": {"type": "number", "description": "Max requests per second", "default": 150},
					"concurrency": {"type": "number", "description": "Number of concurrent templates", "default": 25},
					"timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 300},
					"extra_args": {"type": "string", "description": "Additional nuclei CLI arguments"}
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
		severity:    "medium,high,critical",
		rateLimit:   150,
		concurrency: 25,
		timeout:     300,
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["templates"]; ok {
			cfg.templates = v.GetStringValue()
		}

		if v, ok := fields["severity"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.severity = s
			}
		}

		if v, ok := fields["tags"]; ok {
			cfg.tags = v.GetStringValue()
		}

		if v, ok := fields["rate_limit"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.rateLimit = int(n)
			}
		}

		if v, ok := fields["concurrency"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.concurrency = int(n)
			}
		}

		if v, ok := fields["timeout"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.timeout = int(n)
			}
		}

		if v, ok := fields["extra_args"]; ok {
			cfg.extraArgs = v.GetStringValue()
		}
	}

	h.config = cfg
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("nuclei worker initialized",
		"templates", cfg.templates,
		"severity", cfg.severity,
		"tags", cfg.tags,
		"rate_limit", cfg.rateLimit,
		"concurrency", cfg.concurrency,
		"timeout", cfg.timeout,
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

	switch item.GetType() {
	case "url":
		target = h.extractURLTarget(stream, item)
	case "service":
		target = h.extractServiceTarget(stream, item)
	default:
		h.sendError(stream, item.GetId(), fmt.Sprintf("unsupported item type: %s", item.GetType()), false)
		return
	}

	if target == "" {
		return
	}

	slog.Info("running nuclei", "target", target)

	// Build nuclei command arguments
	args := []string{"-u", target, "-jsonl", "-silent"}
	if h.config.severity != "" {
		args = append(args, "-severity", h.config.severity)
	}
	if h.config.templates != "" {
		args = append(args, "-t", h.config.templates)
	}
	if h.config.tags != "" {
		args = append(args, "-tags", h.config.tags)
	}
	args = append(args, "-rl", strconv.Itoa(h.config.rateLimit))
	args = append(args, "-c", strconv.Itoa(h.config.concurrency))
	if h.config.extraArgs != "" {
		extraParts := strings.Fields(h.config.extraArgs)
		args = append(args, extraParts...)
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "nuclei", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("failed to create stdout pipe: %v", err), true)
		return
	}

	if err := cmd.Start(); err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("failed to start nuclei: %v", err), true)
		return
	}

	// Parse JSONL output line by line
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	resultCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var result nucleiResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			slog.Warn("failed to parse nuclei output line", "error", err, "line", line)
			continue
		}

		resultItem := &portwhinev1.DataItem{
			Id:            uuid.New().String(),
			PipelineRunId: h.pipelineRunID,
			Type:          "vulnerability",
			Data: &structpb.Struct{Fields: map[string]*structpb.Value{
				"template_id":  structpb.NewStringValue(result.TemplateID),
				"name":         structpb.NewStringValue(result.Info.Name),
				"severity":     structpb.NewStringValue(result.Info.Severity),
				"description":  structpb.NewStringValue(result.Info.Description),
				"matched_at":   structpb.NewStringValue(result.MatchedAt),
				"matcher_name": structpb.NewStringValue(result.MatcherName),
				"tags":         structpb.NewStringValue(strings.Join(result.Info.Tags, ",")),
				"url":          structpb.NewStringValue(target),
				"type":         structpb.NewStringValue(result.Type),
			}},
			Metadata: &portwhinev1.DataItemMetadata{
				Source:    "nuclei-worker",
				CreatedAt: timestamppb.Now(),
				NodeId:    h.nodeID,
				Labels: map[string]string{
					"worker_type": "nuclei",
					"severity":    result.Info.Severity,
					"template_id": result.TemplateID,
				},
			},
			ParentIds: []string{item.GetId()},
		}

		h.safeSend(stream, &workerv1.ProcessResponse{
			Payload: &workerv1.ProcessResponse_Item{Item: resultItem},
		})
		h.itemsProduced.Add(1)
		resultCount++
	}

	if err := scanner.Err(); err != nil {
		slog.Warn("scanner error reading nuclei output", "error", err)
	}

	if err := cmd.Wait(); err != nil {
		// nuclei may exit non-zero even when it produced valid output
		if resultCount == 0 {
			h.sendError(stream, item.GetId(), fmt.Sprintf("nuclei failed for %s: %v", target, err), true)
			return
		}
		slog.Warn("nuclei exited with error but produced results", "target", target, "error", err, "results", resultCount)
	}

	slog.Info("nuclei completed", "target", target, "results", resultCount)
}

func (h *workerHandler) extractURLTarget(stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], item *portwhinev1.DataItem) string {
	targetURL := ""
	if item.GetData() != nil {
		if v, ok := item.GetData().GetFields()["url"]; ok {
			targetURL = v.GetStringValue()
		}
	}
	if targetURL == "" {
		h.sendError(stream, item.GetId(), "missing 'url' field in data", false)
		return ""
	}

	// Validate URL
	parsed, err := neturl.Parse(targetURL)
	if err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("invalid URL %q: %v", targetURL, err), false)
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		h.sendError(stream, item.GetId(), fmt.Sprintf("URL must have http or https scheme, got %q", parsed.Scheme), false)
		return ""
	}

	return targetURL
}

func (h *workerHandler) extractServiceTarget(stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], item *portwhinev1.DataItem) string {
	if item.GetData() == nil {
		h.sendError(stream, item.GetId(), "missing data in service item", false)
		return ""
	}

	fields := item.GetData().GetFields()
	host := dataitem.ExtractServiceHost(fields)
	port := dataitem.ExtractServicePort(fields)
	serviceName := dataitem.ExtractServiceName(fields)

	if host == "" {
		h.sendError(stream, item.GetId(), "missing host in service item", false)
		return ""
	}
	if port == 0 {
		h.sendError(stream, item.GetId(), "missing port in service item", false)
		return ""
	}

	// For HTTP services, extract the URL
	if dataitem.IsHTTPServiceName(serviceName) {
		if serviceURL, ok := dataitem.ExtractServiceURL(fields); ok {
			return serviceURL
		}
	}

	// For non-HTTP services, use host:port
	return dataitem.FormatHostPort(host, port)
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
