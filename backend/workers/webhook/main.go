package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
	workerv1 "github.com/portwhine/portwhine/gen/go/portwhine/worker/v1"
	"github.com/portwhine/portwhine/gen/go/portwhine/worker/v1/workerv1connect"
)

type webhookConfig struct {
	url     string
	method  string
	headers map[string]string
	timeout time.Duration
}

type workerHandler struct {
	workerv1connect.UnimplementedWorkerServiceHandler

	mu             sync.Mutex
	streamMu       sync.Mutex
	nodeID         string
	pipelineRunID  string
	config         *webhookConfig
	httpClient     *http.Client
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
		slog.Info("webhook-output worker listening", "addr", ":50051")
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
			Name:               "webhook-output",
			Version:            "1.0.0",
			AcceptedInputTypes: []string{"service", "url", "vulnerability", "ssl_result", "http_headers", "web_technology", "screenshot", "ssh_audit_result", "whois_result", "ip_address", "dns_record", "domain", "report"},
			OutputTypes:        []string{"webhook_delivery"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "url": {"type": "string", "description": "Destination webhook URL"},
    "method": {"type": "string", "description": "HTTP method (POST or PUT)", "default": "POST"},
    "headers": {"type": "object", "description": "Custom HTTP headers (e.g. Authorization)"},
    "timeout": {"type": "number", "description": "HTTP request timeout in seconds", "default": 30}
  },
  "required": ["url"]
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

	cfg := &webhookConfig{
		method:  "POST",
		headers: make(map[string]string),
		timeout: 30 * time.Second,
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["url"]; ok {
			cfg.url = v.GetStringValue()
		}
		if v, ok := fields["method"]; ok {
			if s := v.GetStringValue(); s == "POST" || s == "PUT" {
				cfg.method = s
			}
		}
		if v, ok := fields["headers"]; ok {
			if sv := v.GetStructValue(); sv != nil {
				for k, val := range sv.GetFields() {
					cfg.headers[k] = val.GetStringValue()
				}
			}
		}
		if v, ok := fields["timeout"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.timeout = time.Duration(n) * time.Second
			}
		}
	}

	if cfg.url == "" {
		return connect.NewResponse(&workerv1.InitializeResponse{
			Success:      false,
			ErrorMessage: "url parameter is required",
		}), nil
	}

	h.config = cfg
	h.httpClient = &http.Client{Timeout: cfg.timeout}
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("webhook-output initialized",
		"url", cfg.url,
		"method", cfg.method,
		"timeout", cfg.timeout,
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
			slog.Debug("flush received")
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

	// Serialize the DataItem to JSON using protojson for a well-defined format.
	jsonBytes, err := protojson.Marshal(item)
	if err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("failed to serialize item: %v", err), false)
		return
	}

	req, err := http.NewRequestWithContext(ctx, h.config.method, h.config.url, bytes.NewReader(jsonBytes))
	if err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("failed to build request: %v", err), false)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Portwhine-Run-ID", h.pipelineRunID)
	req.Header.Set("X-Portwhine-Node-ID", h.nodeID)
	req.Header.Set("X-Portwhine-Item-Type", item.GetType())
	for k, v := range h.config.headers {
		req.Header.Set(k, v)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("webhook request failed: %v", err), true)
		return
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success: emit a delivery confirmation item.
		h.emitDelivery(stream, item, resp.StatusCode)
	} else if resp.StatusCode >= 500 || resp.StatusCode == 429 {
		h.sendError(stream, item.GetId(), fmt.Sprintf("webhook returned %d", resp.StatusCode), true)
	} else {
		h.sendError(stream, item.GetId(), fmt.Sprintf("webhook returned %d", resp.StatusCode), false)
	}
}

func (h *workerHandler) emitDelivery(stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], item *portwhinev1.DataItem, statusCode int) {
	deliveryData, _ := structpb.NewStruct(map[string]interface{}{
		"url":         h.config.url,
		"method":      h.config.method,
		"status_code": float64(statusCode),
		"item_id":     item.GetId(),
		"item_type":   item.GetType(),
	})

	deliveryItem := &portwhinev1.DataItem{
		Id:            uuid.New().String(),
		PipelineRunId: h.pipelineRunID,
		Type:          "webhook_delivery",
		Data:          deliveryData,
		Metadata: &portwhinev1.DataItemMetadata{
			Source:    "webhook-output",
			CreatedAt: timestamppb.Now(),
			NodeId:    h.nodeID,
		},
		ParentIds: []string{item.GetId()},
	}

	h.safeSend(stream, &workerv1.ProcessResponse{
		Payload: &workerv1.ProcessResponse_Item{Item: deliveryItem},
	})
	h.itemsProduced.Add(1)

	slog.Debug("webhook delivered",
		"item_id", item.GetId(),
		"status_code", statusCode,
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
	slog.Warn("webhook error", "item_id", sourceItemID, "error", msg, "retryable", retryable)
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
