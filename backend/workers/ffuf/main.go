package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
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

type ffufOutput struct {
	Results []ffufResult `json:"results"`
}

type ffufResult struct {
	Input       map[string]string `json:"input"`
	Status      int               `json:"status"`
	Length      int               `json:"length"`
	Words       int               `json:"words"`
	Lines       int               `json:"lines"`
	ContentType string            `json:"content-type"`
	URL         string            `json:"url"`
}

type workerConfig struct {
	wordlist        string
	extensions      string
	threads         int
	matchCodes      string
	extraArgs       string
	headers         map[string]string
	method          string
	postData        string
	rateLimit       int
	recursion       bool
	recursionDepth  int
	filterSize      string
	filterWords     string
	filterLines     string
	filterRegex     string
	autoCalibrate   bool
	proxy           string
	timeout         int
	followRedirects bool
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
			Name:               "ffuf-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"url"},
			OutputTypes:        []string{"url"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"wordlist": {"type": "string", "description": "Path to wordlist or 'common' for built-in", "default": "common"},
					"extensions": {"type": "string", "description": "File extensions to fuzz (e.g. '.php,.html')"},
					"threads": {"type": "number", "description": "Number of concurrent threads", "default": 40},
					"match_codes": {"type": "string", "description": "HTTP status codes to match", "default": "200,204,301,302,307,401,403"},
					"extra_args": {"type": "string", "description": "Additional ffuf command-line arguments"},
					"headers": {"type": "object", "description": "Custom HTTP headers"},
					"method": {"type": "string", "description": "HTTP method", "default": "GET"},
					"post_data": {"type": "string", "description": "POST body data"},
					"rate_limit": {"type": "number", "description": "Max requests/sec (0=unlimited)", "default": 0},
					"recursion": {"type": "boolean", "description": "Enable recursive fuzzing", "default": false},
					"recursion_depth": {"type": "number", "description": "Max recursion depth", "default": 2},
					"filter_size": {"type": "string", "description": "Filter by response size (-fs)"},
					"filter_words": {"type": "string", "description": "Filter by word count (-fw)"},
					"filter_lines": {"type": "string", "description": "Filter by line count (-fl)"},
					"filter_regex": {"type": "string", "description": "Filter by regex (-fr)"},
					"auto_calibrate": {"type": "boolean", "description": "Auto-calibrate filters (-ac)", "default": false},
					"proxy": {"type": "string", "description": "HTTP proxy URL"},
					"timeout": {"type": "number", "description": "Per-request timeout in seconds", "default": 10},
					"follow_redirects": {"type": "boolean", "description": "Follow HTTP redirects", "default": true}
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
		wordlist:        "common",
		threads:         40,
		matchCodes:      "200,204,301,302,307,401,403",
		method:          "GET",
		recursionDepth:  2,
		timeout:         10,
		followRedirects: true,
		headers:         make(map[string]string),
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["wordlist"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.wordlist = s
			}
		}

		if v, ok := fields["extensions"]; ok {
			cfg.extensions = v.GetStringValue()
		}

		if v, ok := fields["threads"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.threads = int(n)
			}
		}

		if v, ok := fields["match_codes"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.matchCodes = s
			}
		}

		if v, ok := fields["extra_args"]; ok {
			cfg.extraArgs = v.GetStringValue()
		}

		if v, ok := fields["headers"]; ok {
			if s := v.GetStructValue(); s != nil {
				for k, val := range s.GetFields() {
					cfg.headers[k] = val.GetStringValue()
				}
			}
		}

		if v, ok := fields["method"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.method = s
			}
		}

		if v, ok := fields["post_data"]; ok {
			cfg.postData = v.GetStringValue()
		}

		if v, ok := fields["rate_limit"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.rateLimit = int(n)
			}
		}

		if v, ok := fields["recursion"]; ok {
			cfg.recursion = v.GetBoolValue()
		}

		if v, ok := fields["recursion_depth"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.recursionDepth = int(n)
			}
		}

		if v, ok := fields["filter_size"]; ok {
			cfg.filterSize = v.GetStringValue()
		}

		if v, ok := fields["filter_words"]; ok {
			cfg.filterWords = v.GetStringValue()
		}

		if v, ok := fields["filter_lines"]; ok {
			cfg.filterLines = v.GetStringValue()
		}

		if v, ok := fields["filter_regex"]; ok {
			cfg.filterRegex = v.GetStringValue()
		}

		if v, ok := fields["auto_calibrate"]; ok {
			cfg.autoCalibrate = v.GetBoolValue()
		}

		if v, ok := fields["proxy"]; ok {
			cfg.proxy = v.GetStringValue()
		}

		if v, ok := fields["timeout"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.timeout = int(n)
			}
		}

		if v, ok := fields["follow_redirects"]; ok {
			cfg.followRedirects = v.GetBoolValue()
		}
	}

	h.config = cfg
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("ffuf worker initialized",
		"wordlist", cfg.wordlist,
		"extensions", cfg.extensions,
		"threads", cfg.threads,
		"match_codes", cfg.matchCodes,
		"method", cfg.method,
		"timeout", cfg.timeout,
		"follow_redirects", cfg.followRedirects,
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

	var targetURL string

	if item.GetType() != "url" {
		h.sendError(stream, item.GetId(), fmt.Sprintf("unsupported item type: %s", item.GetType()), false)
		return
	}

	if item.GetData() != nil {
		if v, ok := item.GetData().GetFields()["url"]; ok {
			targetURL = v.GetStringValue()
		}
	}
	if targetURL == "" {
		h.sendError(stream, item.GetId(), "missing 'url' field in data", false)
		return
	}
	if !strings.Contains(targetURL, "FUZZ") {
		targetURL = strings.TrimRight(targetURL, "/") + "/FUZZ"
	}

	// Validate the URL for security
	if _, err := url.Parse(targetURL); err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("invalid target URL: %v", err), false)
		return
	}

	// Resolve wordlist path
	wordlist := h.config.wordlist
	if wordlist == "common" {
		wordlist = "/wordlists/common.txt"
	}

	// Build ffuf command arguments
	args := []string{
		"-u", targetURL,
		"-w", wordlist,
		"-mc", h.config.matchCodes,
		"-t", strconv.Itoa(h.config.threads),
		"-of", "json",
		"-o", "/dev/stdout",
	}

	if h.config.extensions != "" {
		args = append(args, "-e", h.config.extensions)
	}

	if h.config.method != "GET" {
		args = append(args, "-X", h.config.method)
	}

	if h.config.postData != "" {
		args = append(args, "-d", h.config.postData)
	}

	for k, v := range h.config.headers {
		args = append(args, "-H", fmt.Sprintf("%s: %s", k, v))
	}

	if h.config.rateLimit > 0 {
		args = append(args, "-rate", strconv.Itoa(h.config.rateLimit))
	}

	if h.config.recursion {
		args = append(args, "-recursion", "-recursion-depth", strconv.Itoa(h.config.recursionDepth))
	}

	if h.config.filterSize != "" {
		args = append(args, "-fs", h.config.filterSize)
	}

	if h.config.filterWords != "" {
		args = append(args, "-fw", h.config.filterWords)
	}

	if h.config.filterLines != "" {
		args = append(args, "-fl", h.config.filterLines)
	}

	if h.config.filterRegex != "" {
		args = append(args, "-fr", h.config.filterRegex)
	}

	if h.config.autoCalibrate {
		args = append(args, "-ac")
	}

	if h.config.proxy != "" {
		args = append(args, "-x", h.config.proxy)
	}

	if h.config.timeout != 10 {
		args = append(args, "-timeout", strconv.Itoa(h.config.timeout))
	}

	if h.config.followRedirects {
		args = append(args, "-r")
	}

	if h.config.extraArgs != "" {
		extraParts := strings.Fields(h.config.extraArgs)
		args = append(args, extraParts...)
	}

	slog.Info("running ffuf",
		"target", targetURL,
		"wordlist", wordlist,
		"args", args,
	)

	// Run ffuf with a 10-minute timeout
	execCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "ffuf", args...)
	output, err := cmd.Output()
	if err != nil {
		// ffuf may return non-zero exit code but still produce valid output
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(output) > 0 {
			slog.Warn("ffuf exited with non-zero status but produced output",
				"exit_code", exitErr.ExitCode(),
				"target", targetURL,
			)
		} else {
			h.sendError(stream, item.GetId(), fmt.Sprintf("ffuf execution failed: %v", err), true)
			return
		}
	}

	// Parse ffuf JSON output
	var result ffufOutput
	if err := json.Unmarshal(output, &result); err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("failed to parse ffuf output: %v", err), true)
		return
	}

	slog.Info("ffuf completed",
		"target", targetURL,
		"results_count", len(result.Results),
	)

	// Emit results as url DataItems
	for _, r := range result.Results {
		dataItem := &portwhinev1.DataItem{
			Id:            uuid.New().String(),
			PipelineRunId: h.pipelineRunID,
			Type:          "url",
			Data: &structpb.Struct{Fields: map[string]*structpb.Value{
				"url":            structpb.NewStringValue(r.URL),
				"status_code":    structpb.NewNumberValue(float64(r.Status)),
				"content_length": structpb.NewNumberValue(float64(r.Length)),
				"content_type":   structpb.NewStringValue(r.ContentType),
				"lines":          structpb.NewNumberValue(float64(r.Lines)),
				"words":          structpb.NewNumberValue(float64(r.Words)),
			}},
			Metadata: &portwhinev1.DataItemMetadata{
				Source:    "ffuf-worker",
				CreatedAt: timestamppb.Now(),
				NodeId:    h.nodeID,
				Labels: map[string]string{
					"worker_type": "ffuf",
					"status_code": strconv.Itoa(r.Status),
				},
			},
			ParentIds: []string{item.GetId()},
		}

		h.safeSend(stream, &workerv1.ProcessResponse{
			Payload: &workerv1.ProcessResponse_Item{Item: dataItem},
		})
		h.itemsProduced.Add(1)
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
