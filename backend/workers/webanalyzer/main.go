package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
	workerv1 "github.com/portwhine/portwhine/gen/go/portwhine/worker/v1"
	"github.com/portwhine/portwhine/gen/go/portwhine/worker/v1/workerv1connect"
)

const maxBodySize = 10 * 1024 * 1024 // 10MB

type webanalyzerConfig struct {
	timeout          int
	headers          map[string]string
	maxRedirects     int
	userAgent        string
	ignoreCertErrors bool
}

type workerHandler struct {
	workerv1connect.UnimplementedWorkerServiceHandler

	mu             sync.Mutex
	streamMu       sync.Mutex
	nodeID         string
	pipelineRunID  string
	config         webanalyzerConfig
	wappalyzer     *wappalyzer.Wappalyze
	httpClient     *http.Client
	status         portwhinev1.WorkerStatus
	itemsProcessed atomic.Uint64
	itemsProduced  atomic.Uint64
	errorsCount    atomic.Uint64
	initialized    bool
}

func defaultWebanalyzerConfig() webanalyzerConfig {
	return webanalyzerConfig{
		timeout:          30,
		headers:          make(map[string]string),
		maxRedirects:     5,
		userAgent:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		ignoreCertErrors: true,
	}
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
		slog.Info("webanalyzer worker listening", "addr", ":50051")
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
			Name:               "webanalyzer-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"url"},
			OutputTypes:        []string{"web_technology"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"timeout": {"type": "number", "description": "HTTP request timeout in seconds", "default": 30},
					"headers": {"type": "object", "description": "Custom HTTP headers"},
					"max_redirects": {"type": "number", "description": "Maximum redirects to follow", "default": 5},
					"user_agent": {"type": "string", "description": "Custom User-Agent string"},
					"ignore_cert_errors": {"type": "boolean", "description": "Skip TLS cert validation", "default": true}
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

	cfg := defaultWebanalyzerConfig()

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["timeout"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.timeout = int(n)
			}
		}

		if v, ok := fields["headers"]; ok {
			if s := v.GetStructValue(); s != nil {
				for k, val := range s.GetFields() {
					cfg.headers[k] = val.GetStringValue()
				}
			}
		}

		if v, ok := fields["max_redirects"]; ok {
			if n := v.GetNumberValue(); n >= 0 {
				cfg.maxRedirects = int(n)
			}
		}

		if v, ok := fields["user_agent"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.userAgent = s
			}
		}

		if v, ok := fields["ignore_cert_errors"]; ok {
			cfg.ignoreCertErrors = v.GetBoolValue()
		}
	}

	h.config = cfg

	wappalyzerClient, err := wappalyzer.New()
	if err != nil {
		return connect.NewResponse(&workerv1.InitializeResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to initialize wappalyzer: %v", err),
		}), nil
	}
	h.wappalyzer = wappalyzerClient

	h.httpClient = &http.Client{
		Timeout: time.Duration(cfg.timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.ignoreCertErrors,
			},
			DialContext: (&net.Dialer{
				Timeout: time.Duration(cfg.timeout) * time.Second,
			}).DialContext,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= cfg.maxRedirects {
				return fmt.Errorf("stopped after %d redirects", cfg.maxRedirects)
			}
			return nil
		},
	}

	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("webanalyzer worker initialized",
		"node_id", h.nodeID,
		"pipeline_run_id", h.pipelineRunID,
		"timeout", cfg.timeout,
		"max_redirects", cfg.maxRedirects,
		"ignore_cert_errors", cfg.ignoreCertErrors,
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
	parsedURL, err := url.Parse(targetURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		h.sendError(stream, item.GetId(), fmt.Sprintf("invalid URL: %s", targetURL), false)
		return
	}

	// Fetch URL and fingerprint technologies
	headers, body, statusCode, serverHeader, err := h.fetchURL(ctx, targetURL)
	if err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("failed to fetch %s: %v", targetURL, err), true)
		return
	}

	fingerprints := h.wappalyzer.FingerprintWithInfo(headers, body)

	techListVal, err := buildTechList(fingerprints)
	if err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("failed to build technology list: %v", err), false)
		return
	}

	dataItem := &portwhinev1.DataItem{
		Id:            uuid.New().String(),
		PipelineRunId: h.pipelineRunID,
		Type:          "web_technology",
		Data: &structpb.Struct{Fields: map[string]*structpb.Value{
			"url":          structpb.NewStringValue(targetURL),
			"technologies": structpb.NewListValue(techListVal),
			"status_code":  structpb.NewNumberValue(float64(statusCode)),
			"server":       structpb.NewStringValue(serverHeader),
		}},
		Metadata: &portwhinev1.DataItemMetadata{
			Source:    "webanalyzer-worker",
			CreatedAt: timestamppb.Now(),
			NodeId:    h.nodeID,
			Labels: map[string]string{
				"worker_type":      "webanalyzer",
				"technology_count": fmt.Sprintf("%d", len(fingerprints)),
			},
		},
		ParentIds: []string{item.GetId()},
	}

	h.safeSend(stream, &workerv1.ProcessResponse{
		Payload: &workerv1.ProcessResponse_Item{Item: dataItem},
	})
	h.itemsProduced.Add(1)

	slog.Info("analyzed URL",
		"url", targetURL,
		"technologies_found", len(fingerprints),
	)
}

func (h *workerHandler) fetchURL(ctx context.Context, targetURL string) (http.Header, []byte, int, string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, nil, 0, "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("User-Agent", h.config.userAgent)

	for k, v := range h.config.headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, 0, "", fmt.Errorf("HTTP GET: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return nil, nil, 0, "", fmt.Errorf("read body: %w", err)
	}

	return resp.Header, body, resp.StatusCode, resp.Header.Get("Server"), nil
}

func buildTechList(fingerprints map[string]wappalyzer.AppInfo) (*structpb.ListValue, error) {
	var techList []interface{}
	for name, info := range fingerprints {
		tech := map[string]interface{}{
			"name": name,
		}
		if len(info.Categories) > 0 {
			tech["categories"] = info.Categories
		}
		if info.Description != "" {
			tech["description"] = info.Description
		}
		if info.Website != "" {
			tech["website"] = info.Website
		}
		techList = append(techList, tech)
	}

	return structpb.NewList(techList)
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
