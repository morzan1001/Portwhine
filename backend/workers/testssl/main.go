package main

import (
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

// testsslFinding represents a single finding from testssl.sh JSON output.
type testsslFinding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Finding  string `json:"finding"`
	IP       string `json:"ip"`
	Port     string `json:"port"`
}

type workerConfig struct {
	checks              string
	extraArgs           string
	protocolsOnly       bool
	vulnerabilitiesOnly bool
	ciphers             bool
	headers             bool
	sni                 string
	timeout             int
	parallel            bool
	starttls            string
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
		slog.Info("testssl worker listening", "addr", ":50051")
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
			Name:               "testssl-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"service"},
			OutputTypes:        []string{"ssl_result"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"checks": {"type": "string", "description": "Specific testssl.sh checks to run (e.g. --protocols --ciphers)"},
					"extra_args": {"type": "string", "description": "Additional arguments to pass to testssl.sh"},
					"protocols_only": {"type": "boolean", "description": "Only test protocols (--protocols)", "default": false},
					"vulnerabilities_only": {"type": "boolean", "description": "Only test vulns (--vulnerable)", "default": false},
					"ciphers": {"type": "boolean", "description": "Test cipher suites (--ciphers)", "default": false},
					"headers": {"type": "boolean", "description": "Check HTTP headers (--headers)", "default": false},
					"sni": {"type": "string", "description": "Server Name Indication value"},
					"timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 300},
					"parallel": {"type": "boolean", "description": "Run checks in parallel (--parallel)", "default": false},
					"starttls": {"type": "string", "description": "STARTTLS protocol override (smtp, ftp, etc.). Auto-detected from service_name if not set."}
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
		timeout: 300,
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["checks"]; ok {
			cfg.checks = v.GetStringValue()
		}

		if v, ok := fields["extra_args"]; ok {
			cfg.extraArgs = v.GetStringValue()
		}

		if v, ok := fields["protocols_only"]; ok {
			cfg.protocolsOnly = v.GetBoolValue()
		}

		if v, ok := fields["vulnerabilities_only"]; ok {
			cfg.vulnerabilitiesOnly = v.GetBoolValue()
		}

		if v, ok := fields["ciphers"]; ok {
			cfg.ciphers = v.GetBoolValue()
		}

		if v, ok := fields["headers"]; ok {
			cfg.headers = v.GetBoolValue()
		}

		if v, ok := fields["sni"]; ok {
			cfg.sni = v.GetStringValue()
		}

		if v, ok := fields["timeout"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.timeout = int(n)
			}
		}

		if v, ok := fields["parallel"]; ok {
			cfg.parallel = v.GetBoolValue()
		}

		if v, ok := fields["starttls"]; ok {
			cfg.starttls = v.GetStringValue()
		}
	}

	h.config = cfg
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("testssl worker initialized",
		"checks", cfg.checks,
		"timeout", cfg.timeout,
		"protocols_only", cfg.protocolsOnly,
		"vulnerabilities_only", cfg.vulnerabilitiesOnly,
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

	if item.GetType() != "service" {
		h.sendError(stream, item.GetId(), fmt.Sprintf("unsupported item type: %s", item.GetType()), false)
		return
	}

	if item.GetData() == nil {
		h.sendError(stream, item.GetId(), "missing data in service item", false)
		return
	}

	fields := item.GetData().GetFields()
	host := dataitem.ExtractServiceHost(fields)
	port := dataitem.ExtractServicePort(fields)
	serviceName := dataitem.ExtractServiceName(fields)

	if host == "" {
		h.sendError(stream, item.GetId(), "missing host in service item", false)
		return
	}
	if port == 0 {
		h.sendError(stream, item.GetId(), "missing port in service item", false)
		return
	}

	// Build the target string host:port
	target := dataitem.FormatHostPort(host, port)

	// Build the testssl.sh command arguments
	args := []string{
		"--jsonfile", "/dev/stdout",
		"--fast",
		"--quiet",
	}

	if h.config.protocolsOnly {
		args = append(args, "--protocols")
	}

	if h.config.vulnerabilitiesOnly {
		args = append(args, "--vulnerable")
	}

	if h.config.ciphers {
		args = append(args, "--ciphers")
	}

	if h.config.headers {
		args = append(args, "--headers")
	}

	if h.config.sni != "" {
		args = append(args, "--sni", h.config.sni)
	}

	if h.config.parallel {
		args = append(args, "--parallel")
	}

	// STARTTLS: use explicit config override, or auto-detect from service_name.
	starttls := h.config.starttls
	if starttls == "" {
		starttls = dataitem.StarttlsProtocol(serviceName)
	}
	if starttls != "" {
		args = append(args, "--starttls", starttls)
	}

	if h.config.checks != "" {
		checkParts := strings.Fields(h.config.checks)
		args = append(args, checkParts...)
	}

	if h.config.extraArgs != "" {
		extraParts := strings.Fields(h.config.extraArgs)
		args = append(args, extraParts...)
	}

	args = append(args, target)

	slog.Info("running testssl.sh", "target", target, "args", args)

	// Run testssl.sh with configurable timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "testssl.sh", args...)
	output, err := cmd.Output()
	if err != nil {
		// testssl.sh may exit with non-zero even on partial success; check if we got output
		if len(output) == 0 {
			h.sendError(stream, item.GetId(), fmt.Sprintf("testssl.sh failed for %s: %v", target, err), true)
			return
		}
		slog.Warn("testssl.sh exited with error but produced output", "target", target, "error", err)
	}

	// Parse the JSON output - testssl outputs a JSON array of findings
	var findings []testsslFinding
	if err := json.Unmarshal(output, &findings); err != nil {
		// testssl.sh sometimes wraps output; try to find the JSON array
		startIdx := strings.Index(string(output), "[")
		endIdx := strings.LastIndex(string(output), "]")
		if startIdx >= 0 && endIdx > startIdx {
			jsonSlice := output[startIdx : endIdx+1]
			if err2 := json.Unmarshal(jsonSlice, &findings); err2 != nil {
				h.sendError(stream, item.GetId(), fmt.Sprintf("failed to parse testssl.sh output for %s: %v", target, err2), true)
				return
			}
		} else {
			h.sendError(stream, item.GetId(), fmt.Sprintf("failed to parse testssl.sh output for %s: %v", target, err), true)
			return
		}
	}

	if len(findings) == 0 {
		h.sendError(stream, item.GetId(), fmt.Sprintf("testssl.sh produced no findings for %s", target), true)
		return
	}

	// Aggregate findings
	resultItem := h.aggregateFindings(findings, host, port, item)

	h.safeSend(stream, &workerv1.ProcessResponse{
		Payload: &workerv1.ProcessResponse_Item{Item: resultItem},
	})
	h.itemsProduced.Add(1)
}

func (h *workerHandler) aggregateFindings(findings []testsslFinding, host string, port int, parentItem *portwhinev1.DataItem) *portwhinev1.DataItem {
	// Extract protocol versions
	protocolIDs := map[string]string{
		"SSLv2":  "SSLv2",
		"SSLv3":  "SSLv3",
		"TLS1":   "TLSv1.0",
		"TLS1_1": "TLSv1.1",
		"TLS1_2": "TLSv1.2",
		"TLS1_3": "TLSv1.3",
	}

	var offeredProtocols []string
	for _, f := range findings {
		if label, ok := protocolIDs[f.ID]; ok {
			lower := strings.ToLower(f.Finding)
			if strings.Contains(lower, "offered") && !strings.Contains(lower, "not offered") {
				offeredProtocols = append(offeredProtocols, label)
			}
		}
	}

	// Collect vulnerabilities (CRITICAL, HIGH, MEDIUM, WARN)
	var vulnerabilities []*structpb.Value
	severityCounts := map[string]int{
		"CRITICAL": 0,
		"HIGH":     0,
		"MEDIUM":   0,
		"WARN":     0,
	}

	for _, f := range findings {
		sev := strings.ToUpper(f.Severity)
		if sev == "CRITICAL" || sev == "HIGH" || sev == "MEDIUM" || sev == "WARN" {
			severityCounts[sev]++
			vulnStruct, err := structpb.NewStruct(map[string]interface{}{
				"id":       f.ID,
				"severity": f.Severity,
				"finding":  f.Finding,
			})
			if err == nil {
				vulnerabilities = append(vulnerabilities, structpb.NewStructValue(vulnStruct))
			}
		}
	}

	// Extract certificate info from findings with IDs starting with "cert_"
	certInfo := map[string]string{
		"subject":    "",
		"issuer":     "",
		"valid_from": "",
		"valid_to":   "",
	}
	for _, f := range findings {
		if !strings.HasPrefix(f.ID, "cert_") {
			continue
		}
		switch f.ID {
		case "cert_commonName", "cert_subjectAltName":
			if certInfo["subject"] == "" {
				certInfo["subject"] = f.Finding
			}
		case "cert_caIssuers":
			if certInfo["issuer"] == "" {
				certInfo["issuer"] = f.Finding
			}
		case "cert_notBefore":
			certInfo["valid_from"] = f.Finding
		case "cert_notAfter":
			certInfo["valid_to"] = f.Finding
		}
	}

	// Compute grade based on severity counts
	grade := "A"
	if severityCounts["CRITICAL"] > 0 {
		grade = "F"
	} else if severityCounts["HIGH"] > 0 {
		grade = "C"
	} else if severityCounts["MEDIUM"] > 0 {
		grade = "B"
	} else if severityCounts["WARN"] > 0 {
		grade = "B"
	}

	// Build protocol_versions list
	protoValues := make([]*structpb.Value, len(offeredProtocols))
	for i, p := range offeredProtocols {
		protoValues[i] = structpb.NewStringValue(p)
	}

	// Build certificate struct
	certStruct, _ := structpb.NewStruct(map[string]interface{}{
		"subject":    certInfo["subject"],
		"issuer":     certInfo["issuer"],
		"valid_from": certInfo["valid_from"],
		"valid_to":   certInfo["valid_to"],
	})

	// Build the vulnerabilities list
	vulnList, _ := structpb.NewList(make([]interface{}, 0))
	if len(vulnerabilities) > 0 {
		vulnList = &structpb.ListValue{Values: vulnerabilities}
	}

	data := &structpb.Struct{Fields: map[string]*structpb.Value{
		"host":              structpb.NewStringValue(host),
		"port":              structpb.NewNumberValue(float64(port)),
		"protocol_versions": structpb.NewListValue(&structpb.ListValue{Values: protoValues}),
		"vulnerabilities":   structpb.NewListValue(vulnList),
		"certificate":       structpb.NewStructValue(certStruct),
		"grade":             structpb.NewStringValue(grade),
		"findings_count":    structpb.NewNumberValue(float64(len(findings))),
	}}

	return &portwhinev1.DataItem{
		Id:            uuid.New().String(),
		PipelineRunId: h.pipelineRunID,
		Type:          "ssl_result",
		Data:          data,
		Metadata: &portwhinev1.DataItemMetadata{
			Source:    "testssl-worker",
			CreatedAt: timestamppb.Now(),
			NodeId:    h.nodeID,
			Labels: map[string]string{
				"worker_type": "testssl",
				"grade":       grade,
			},
		},
		ParentIds: []string{parentItem.GetId()},
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
