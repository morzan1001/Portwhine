package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
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

type workerConfig struct {
	dnsServer   string
	recordTypes []string
	timeout     int
	retries     int
}

type workerHandler struct {
	workerv1connect.UnimplementedWorkerServiceHandler

	mu             sync.Mutex
	streamMu       sync.Mutex
	nodeID         string
	pipelineRunID  string
	config         *workerConfig
	resolver       *net.Resolver
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
		slog.Info("resolver worker listening", "addr", ":50051")
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
			Name:               "resolver-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"domain"},
			OutputTypes:        []string{"ip_address", "dns_record"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"dns_server": {"type": "string", "description": "Custom DNS server (e.g. 8.8.8.8:53)"},
					"record_types": {"type": "array", "items": {"type": "string"}, "description": "DNS record types to resolve (A, AAAA, CNAME, MX, TXT, NS)", "default": ["A", "AAAA"]},
					"timeout": {"type": "number", "description": "DNS lookup timeout in seconds", "default": 10},
					"retries": {"type": "number", "description": "Retry count on failure", "default": 2}
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
		recordTypes: []string{"A", "AAAA"},
		timeout:     10,
		retries:     2,
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["dns_server"]; ok {
			cfg.dnsServer = v.GetStringValue()
		}

		if v, ok := fields["record_types"]; ok {
			if list := v.GetListValue(); list != nil {
				cfg.recordTypes = nil
				for _, item := range list.GetValues() {
					if s := item.GetStringValue(); s != "" {
						cfg.recordTypes = append(cfg.recordTypes, strings.ToUpper(s))
					}
				}
			}
		}

		if v, ok := fields["timeout"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.timeout = int(n)
			}
		}

		if v, ok := fields["retries"]; ok {
			if n := v.GetNumberValue(); n >= 0 {
				cfg.retries = int(n)
			}
		}
	}

	h.config = cfg

	// Set up resolver
	dialTimeout := time.Duration(cfg.timeout) * time.Second
	h.resolver = &net.Resolver{}
	if cfg.dnsServer != "" {
		addr := cfg.dnsServer
		if !strings.Contains(addr, ":") {
			addr += ":53"
		}
		h.resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: dialTimeout}
				return d.DialContext(ctx, "udp", addr)
			},
		}
	}

	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("resolver worker initialized",
		"dns_server", cfg.dnsServer,
		"record_types", cfg.recordTypes,
		"timeout", cfg.timeout,
		"retries", cfg.retries,
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

	domain := ""
	if item.GetData() != nil {
		if v, ok := item.GetData().GetFields()["domain"]; ok {
			domain = v.GetStringValue()
		}
	}
	if domain == "" {
		h.sendError(stream, item.GetId(), "missing 'domain' field in data", false)
		return
	}

	resolveCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
	defer cancel()

	for _, recType := range h.config.recordTypes {
		switch recType {
		case "A", "AAAA":
			h.resolveAddresses(resolveCtx, stream, item, domain, recType)
		case "CNAME":
			h.resolveCNAME(resolveCtx, stream, item, domain)
		case "MX":
			h.resolveMX(resolveCtx, stream, item, domain)
		case "TXT":
			h.resolveTXT(resolveCtx, stream, item, domain)
		case "NS":
			h.resolveNS(resolveCtx, stream, item, domain)
		default:
			slog.Warn("unsupported record type", "type", recType)
		}
	}
}

func (h *workerHandler) resolveAddresses(ctx context.Context, stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], parentItem *portwhinev1.DataItem, domain, recType string) bool {
	var ips []string
	var err error

	for attempt := 0; attempt <= h.config.retries; attempt++ {
		ips, err = h.resolver.LookupHost(ctx, domain)
		if err == nil {
			break
		}
		if attempt < h.config.retries {
			slog.Debug("DNS lookup retry", "domain", domain, "attempt", attempt+1, "error", err)
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}
	if err != nil {
		h.sendError(stream, parentItem.GetId(), fmt.Sprintf("DNS lookup failed for %s: %v", domain, err), true)
		return false
	}

	emitted := false
	for _, ip := range ips {
		parsed := net.ParseIP(ip)
		if parsed == nil {
			continue
		}

		// Filter by record type
		isIPv4 := parsed.To4() != nil
		if recType == "A" && !isIPv4 {
			continue
		}
		if recType == "AAAA" && isIPv4 {
			continue
		}

		version := "4"
		if !isIPv4 {
			version = "6"
		}

		item := &portwhinev1.DataItem{
			Id:            uuid.New().String(),
			PipelineRunId: h.pipelineRunID,
			Type:          "ip_address",
			Data: &structpb.Struct{Fields: map[string]*structpb.Value{
				"ip":      structpb.NewStringValue(ip),
				"version": structpb.NewStringValue(version),
				"domain":  structpb.NewStringValue(domain),
			}},
			Metadata: &portwhinev1.DataItemMetadata{
				Source:    "resolver-worker",
				CreatedAt: timestamppb.Now(),
				NodeId:    h.nodeID,
				Labels: map[string]string{
					"worker_type": "resolver",
					"record_type": recType,
				},
			},
			ParentIds: []string{parentItem.GetId()},
		}

		h.safeSend(stream, &workerv1.ProcessResponse{
			Payload: &workerv1.ProcessResponse_Item{Item: item},
		})
		h.itemsProduced.Add(1)
		emitted = true
	}
	return emitted
}

func (h *workerHandler) resolveCNAME(ctx context.Context, stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], parentItem *portwhinev1.DataItem, domain string) {
	var cname string
	var err error

	for attempt := 0; attempt <= h.config.retries; attempt++ {
		cname, err = h.resolver.LookupCNAME(ctx, domain)
		if err == nil {
			break
		}
		if attempt < h.config.retries {
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}
	if err != nil {
		h.sendError(stream, parentItem.GetId(), fmt.Sprintf("CNAME lookup failed for %s: %v", domain, err), true)
		return
	}

	item := &portwhinev1.DataItem{
		Id:            uuid.New().String(),
		PipelineRunId: h.pipelineRunID,
		Type:          "dns_record",
		Data: &structpb.Struct{Fields: map[string]*structpb.Value{
			"domain":      structpb.NewStringValue(domain),
			"record_type": structpb.NewStringValue("CNAME"),
			"value":       structpb.NewStringValue(cname),
		}},
		Metadata: &portwhinev1.DataItemMetadata{
			Source:    "resolver-worker",
			CreatedAt: timestamppb.Now(),
			NodeId:    h.nodeID,
			Labels:    map[string]string{"worker_type": "resolver", "record_type": "CNAME"},
		},
		ParentIds: []string{parentItem.GetId()},
	}

	h.safeSend(stream, &workerv1.ProcessResponse{
		Payload: &workerv1.ProcessResponse_Item{Item: item},
	})
	h.itemsProduced.Add(1)
}

func (h *workerHandler) resolveMX(ctx context.Context, stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], parentItem *portwhinev1.DataItem, domain string) {
	var mxRecords []*net.MX
	var err error

	for attempt := 0; attempt <= h.config.retries; attempt++ {
		mxRecords, err = h.resolver.LookupMX(ctx, domain)
		if err == nil {
			break
		}
		if attempt < h.config.retries {
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}
	if err != nil {
		h.sendError(stream, parentItem.GetId(), fmt.Sprintf("MX lookup failed for %s: %v", domain, err), true)
		return
	}

	for _, mx := range mxRecords {
		item := &portwhinev1.DataItem{
			Id:            uuid.New().String(),
			PipelineRunId: h.pipelineRunID,
			Type:          "dns_record",
			Data: &structpb.Struct{Fields: map[string]*structpb.Value{
				"domain":      structpb.NewStringValue(domain),
				"record_type": structpb.NewStringValue("MX"),
				"value":       structpb.NewStringValue(mx.Host),
				"priority":    structpb.NewNumberValue(float64(mx.Pref)),
			}},
			Metadata: &portwhinev1.DataItemMetadata{
				Source:    "resolver-worker",
				CreatedAt: timestamppb.Now(),
				NodeId:    h.nodeID,
				Labels:    map[string]string{"worker_type": "resolver", "record_type": "MX"},
			},
			ParentIds: []string{parentItem.GetId()},
		}

		h.safeSend(stream, &workerv1.ProcessResponse{
			Payload: &workerv1.ProcessResponse_Item{Item: item},
		})
		h.itemsProduced.Add(1)
	}
}

func (h *workerHandler) resolveTXT(ctx context.Context, stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], parentItem *portwhinev1.DataItem, domain string) {
	var records []string
	var err error

	for attempt := 0; attempt <= h.config.retries; attempt++ {
		records, err = h.resolver.LookupTXT(ctx, domain)
		if err == nil {
			break
		}
		if attempt < h.config.retries {
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}
	if err != nil {
		h.sendError(stream, parentItem.GetId(), fmt.Sprintf("TXT lookup failed for %s: %v", domain, err), true)
		return
	}

	for _, txt := range records {
		item := &portwhinev1.DataItem{
			Id:            uuid.New().String(),
			PipelineRunId: h.pipelineRunID,
			Type:          "dns_record",
			Data: &structpb.Struct{Fields: map[string]*structpb.Value{
				"domain":      structpb.NewStringValue(domain),
				"record_type": structpb.NewStringValue("TXT"),
				"value":       structpb.NewStringValue(txt),
			}},
			Metadata: &portwhinev1.DataItemMetadata{
				Source:    "resolver-worker",
				CreatedAt: timestamppb.Now(),
				NodeId:    h.nodeID,
				Labels:    map[string]string{"worker_type": "resolver", "record_type": "TXT"},
			},
			ParentIds: []string{parentItem.GetId()},
		}

		h.safeSend(stream, &workerv1.ProcessResponse{
			Payload: &workerv1.ProcessResponse_Item{Item: item},
		})
		h.itemsProduced.Add(1)
	}
}

func (h *workerHandler) resolveNS(ctx context.Context, stream *connect.BidiStream[workerv1.ProcessRequest, workerv1.ProcessResponse], parentItem *portwhinev1.DataItem, domain string) {
	var nsRecords []*net.NS
	var err error

	for attempt := 0; attempt <= h.config.retries; attempt++ {
		nsRecords, err = h.resolver.LookupNS(ctx, domain)
		if err == nil {
			break
		}
		if attempt < h.config.retries {
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}
	if err != nil {
		h.sendError(stream, parentItem.GetId(), fmt.Sprintf("NS lookup failed for %s: %v", domain, err), true)
		return
	}

	for _, ns := range nsRecords {
		item := &portwhinev1.DataItem{
			Id:            uuid.New().String(),
			PipelineRunId: h.pipelineRunID,
			Type:          "dns_record",
			Data: &structpb.Struct{Fields: map[string]*structpb.Value{
				"domain":      structpb.NewStringValue(domain),
				"record_type": structpb.NewStringValue("NS"),
				"value":       structpb.NewStringValue(ns.Host),
			}},
			Metadata: &portwhinev1.DataItemMetadata{
				Source:    "resolver-worker",
				CreatedAt: timestamppb.Now(),
				NodeId:    h.nodeID,
				Labels:    map[string]string{"worker_type": "resolver", "record_type": "NS"},
			},
			ParentIds: []string{parentItem.GetId()},
		}

		h.safeSend(stream, &workerv1.ProcessResponse{
			Payload: &workerv1.ProcessResponse_Item{Item: item},
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
