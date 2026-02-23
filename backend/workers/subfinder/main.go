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
	"os"
	"os/exec"
	"os/signal"
	"regexp"
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
)

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

type subfinderResult struct {
	Host   string `json:"host"`
	Source string `json:"source"`
	Input  string `json:"input"`
}

type workerConfig struct {
	sources   string
	threads   int
	timeout   int
	extraArgs string
	recursive bool
	maxDepth  int
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
		slog.Info("subfinder worker listening", "addr", ":50051")
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
			Name:               "subfinder-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"domain"},
			OutputTypes:        []string{"domain"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"sources": {"type": "string", "description": "Comma-separated list of sources to use"},
					"threads": {"type": "number", "description": "Number of concurrent threads", "default": 10},
					"timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 120},
					"extra_args": {"type": "string", "description": "Additional subfinder CLI arguments"},
					"recursive": {"type": "boolean", "description": "Enable recursive subdomain enumeration", "default": false},
					"max_depth": {"type": "number", "description": "Max recursion depth when recursive is true", "default": 3}
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
		threads:  10,
		timeout:  120,
		maxDepth: 3,
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["sources"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.sources = s
			}
		}

		if v, ok := fields["threads"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.threads = int(n)
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

		if v, ok := fields["recursive"]; ok {
			cfg.recursive = v.GetBoolValue()
		}

		if v, ok := fields["max_depth"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.maxDepth = int(n)
			}
		}
	}

	h.config = cfg
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("subfinder worker initialized",
		"sources", cfg.sources,
		"threads", cfg.threads,
		"timeout", cfg.timeout,
		"recursive", cfg.recursive,
		"max_depth", cfg.maxDepth,
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

	if item.GetType() != "domain" {
		h.sendError(stream, item.GetId(), fmt.Sprintf("unsupported item type: %s", item.GetType()), false)
		return
	}

	var domain string
	if item.GetData() != nil {
		if v, ok := item.GetData().GetFields()["domain"]; ok {
			domain = v.GetStringValue()
		}
	}
	if domain == "" {
		h.sendError(stream, item.GetId(), "missing 'domain' field in data", false)
		return
	}

	// Validate domain: must not contain spaces or special characters
	if !domainRegex.MatchString(domain) {
		h.sendError(stream, item.GetId(), fmt.Sprintf("invalid domain: %s", domain), false)
		return
	}

	// Build subfinder command arguments
	args := []string{"-d", domain, "-silent", "-oJ"}
	if h.config.sources != "" {
		args = append(args, "-sources", h.config.sources)
	}
	if h.config.threads > 0 {
		args = append(args, "-t", strconv.Itoa(h.config.threads))
	}
	if h.config.recursive {
		args = append(args, "-recursive")
		if h.config.maxDepth > 0 {
			args = append(args, "-max-depth", strconv.Itoa(h.config.maxDepth))
		}
	}
	if h.config.extraArgs != "" {
		extraParts := strings.Fields(h.config.extraArgs)
		args = append(args, extraParts...)
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
	defer cancel()

	slog.Debug("running subfinder", "args", args, "domain", domain)

	cmd := exec.CommandContext(execCtx, "subfinder", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("failed to create stdout pipe: %v", err), true)
		return
	}

	if err := cmd.Start(); err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("subfinder execution failed for %s: %v", domain, err), true)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var result subfinderResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			slog.Warn("failed to parse subfinder JSON line", "line", line, "error", err)
			continue
		}

		if result.Host == "" {
			continue
		}

		dataItem := &portwhinev1.DataItem{
			Id:            uuid.New().String(),
			PipelineRunId: h.pipelineRunID,
			Type:          "domain",
			Data: &structpb.Struct{Fields: map[string]*structpb.Value{
				"domain":           structpb.NewStringValue(result.Host),
				"source":           structpb.NewStringValue("subfinder"),
				"parent_domain":    structpb.NewStringValue(domain),
				"discovery_source": structpb.NewStringValue(result.Source),
			}},
			Metadata: &portwhinev1.DataItemMetadata{
				Source:    "subfinder-worker",
				CreatedAt: timestamppb.Now(),
				NodeId:    h.nodeID,
				Labels: map[string]string{
					"worker_type":      "subfinder",
					"discovery_source": result.Source,
				},
			},
			ParentIds: []string{item.GetId()},
		}

		h.safeSend(stream, &workerv1.ProcessResponse{
			Payload: &workerv1.ProcessResponse_Item{Item: dataItem},
		})
		h.itemsProduced.Add(1)
	}

	if err := scanner.Err(); err != nil {
		slog.Warn("error reading subfinder output", "error", err, "domain", domain)
	}

	if err := cmd.Wait(); err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("subfinder process exited with error for %s: %v", domain, err), true)
		return
	}

	slog.Info("subfinder completed",
		"domain", domain,
		"items_produced", h.itemsProduced.Load(),
	)
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
