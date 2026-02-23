package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
	workerv1 "github.com/portwhine/portwhine/gen/go/portwhine/worker/v1"
	"github.com/portwhine/portwhine/gen/go/portwhine/worker/v1/workerv1connect"
)

type screenshotConfig struct {
	width            int
	height           int
	timeout          int
	fullPage         bool
	userAgent        string
	delay            int
	ignoreCertErrors bool
	device           string
	quality          int
	format           string
}

type workerHandler struct {
	workerv1connect.UnimplementedWorkerServiceHandler

	mu             sync.Mutex
	streamMu       sync.Mutex
	nodeID         string
	pipelineRunID  string
	config         *screenshotConfig
	allocCtx       context.Context
	allocCancel    context.CancelFunc
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
		slog.Info("screenshot worker listening", "addr", ":50051")
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
			Name:               "screenshot-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"url"},
			OutputTypes:        []string{"screenshot"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"width": {"type": "number", "description": "Viewport width in pixels", "default": 1280},
					"height": {"type": "number", "description": "Viewport height in pixels", "default": 720},
					"timeout": {"type": "number", "description": "Navigation timeout in seconds", "default": 30},
					"full_page": {"type": "boolean", "description": "Capture full page screenshot", "default": false},
					"user_agent": {"type": "string", "description": "Custom User-Agent string"},
					"delay": {"type": "number", "description": "Seconds to wait after page load", "default": 2},
					"ignore_cert_errors": {"type": "boolean", "description": "Skip TLS certificate validation", "default": true},
					"device": {"type": "string", "description": "Device preset: 'mobile' (375x812) or 'tablet' (768x1024)"},
					"quality": {"type": "number", "description": "Screenshot quality (1-100)", "default": 90},
					"format": {"type": "string", "description": "Output format: 'png' or 'jpeg'", "default": "png"}
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

	cfg := &screenshotConfig{
		width:            1280,
		height:           720,
		timeout:          30,
		fullPage:         false,
		delay:            2,
		ignoreCertErrors: true,
		quality:          90,
		format:           "png",
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["width"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.width = int(n)
			}
		}

		if v, ok := fields["height"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.height = int(n)
			}
		}

		if v, ok := fields["timeout"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.timeout = int(n)
			}
		}

		if v, ok := fields["full_page"]; ok {
			cfg.fullPage = v.GetBoolValue()
		}

		if v, ok := fields["user_agent"]; ok {
			cfg.userAgent = v.GetStringValue()
		}

		if v, ok := fields["delay"]; ok {
			if n := v.GetNumberValue(); n >= 0 {
				cfg.delay = int(n)
			}
		}

		if v, ok := fields["ignore_cert_errors"]; ok {
			cfg.ignoreCertErrors = v.GetBoolValue()
		}

		if v, ok := fields["device"]; ok {
			cfg.device = v.GetStringValue()
		}

		if v, ok := fields["quality"]; ok {
			if n := v.GetNumberValue(); n > 0 && n <= 100 {
				cfg.quality = int(n)
			}
		}

		if v, ok := fields["format"]; ok {
			if s := v.GetStringValue(); s == "png" || s == "jpeg" {
				cfg.format = s
			}
		}
	}

	// Apply device presets
	switch cfg.device {
	case "mobile":
		cfg.width = 375
		cfg.height = 812
	case "tablet":
		cfg.width = 768
		cfg.height = 1024
	}

	h.config = cfg

	// Set up chromedp allocator with options suitable for Alpine Docker
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	if cfg.ignoreCertErrors {
		opts = append(opts, chromedp.Flag("ignore-certificate-errors", true))
	}

	if cfg.userAgent != "" {
		opts = append(opts, chromedp.UserAgent(cfg.userAgent))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	h.allocCtx = allocCtx
	h.allocCancel = allocCancel

	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("screenshot worker initialized",
		"width", cfg.width,
		"height", cfg.height,
		"timeout", cfg.timeout,
		"full_page", cfg.fullPage,
		"delay", cfg.delay,
		"ignore_cert_errors", cfg.ignoreCertErrors,
		"device", cfg.device,
		"format", cfg.format,
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

func (h *workerHandler) processItem(_ context.Context, stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], item *portwhinev1.DataItem) {
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

	// Validate URL
	parsed, err := url.Parse(targetURL)
	if err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("invalid URL %q: %v", targetURL, err), false)
		return
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		h.sendError(stream, item.GetId(), fmt.Sprintf("unsupported URL scheme %q, must be http or https", parsed.Scheme), false)
		return
	}

	// Create a new chromedp context from the allocator
	ctx, cancel := chromedp.NewContext(h.allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
	defer cancel()

	var buf []byte
	var title string
	var currentURL string
	tasks := chromedp.Tasks{
		chromedp.EmulateViewport(int64(h.config.width), int64(h.config.height)),
		chromedp.Navigate(targetURL),
		chromedp.Sleep(time.Duration(h.config.delay) * time.Second),
		chromedp.Title(&title),
		chromedp.Location(&currentURL),
	}
	if h.config.fullPage {
		tasks = append(tasks, chromedp.FullScreenshot(&buf, h.config.quality))
	} else {
		tasks = append(tasks, chromedp.CaptureScreenshot(&buf))
	}

	if err := chromedp.Run(ctx, tasks...); err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("screenshot failed for %s: %v", targetURL, err), true)
		return
	}

	// Emit screenshot DataItem
	screenshotItem := &portwhinev1.DataItem{
		Id:            uuid.New().String(),
		PipelineRunId: h.pipelineRunID,
		Type:          "screenshot",
		Data: &structpb.Struct{Fields: map[string]*structpb.Value{
			"url":       structpb.NewStringValue(targetURL),
			"title":     structpb.NewStringValue(title),
			"final_url": structpb.NewStringValue(currentURL),
			"width":     structpb.NewNumberValue(float64(h.config.width)),
			"height":    structpb.NewNumberValue(float64(h.config.height)),
			"format":    structpb.NewStringValue(h.config.format),
		}},
		RawPayload: buf,
		Metadata: &portwhinev1.DataItemMetadata{
			Source:    "screenshot-worker",
			CreatedAt: timestamppb.Now(),
			NodeId:    h.nodeID,
			Labels: map[string]string{
				"worker_type": "screenshot",
				"format":      h.config.format,
			},
		},
		ParentIds: []string{item.GetId()},
	}

	h.safeSend(stream, &workerv1.ProcessResponse{
		Payload: &workerv1.ProcessResponse_Item{Item: screenshotItem},
	})
	h.itemsProduced.Add(1)

	slog.Info("screenshot captured",
		"url", targetURL,
		"title", title,
		"size_bytes", len(buf),
		"format", h.config.format,
	)
}

func (h *workerHandler) Shutdown(_ context.Context, _ *connect.Request[workerv1.ShutdownRequest]) (*connect.Response[workerv1.ShutdownResponse], error) {
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_STOPPED

	// Cancel the chromedp allocator context
	if h.allocCancel != nil {
		h.allocCancel()
	}

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
