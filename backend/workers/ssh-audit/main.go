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

// sshAuditOutput represents the top-level JSON output from ssh-audit.
type sshAuditOutput struct {
	Banner       string            `json:"banner"`
	Compression  []string          `json:"compression"`
	Enc          []sshAuditAlgo    `json:"enc"`
	Fingerprints []sshFingerprint  `json:"fingerprints"`
	Kex          []sshAuditAlgo    `json:"kex"`
	Key          []sshAuditAlgo    `json:"key"`
	Mac          []sshAuditAlgo    `json:"mac"`
	Target       string            `json:"target"`
	CVEs         []sshAuditCVE     `json:"cves"`
}

// sshAuditAlgo represents a single algorithm entry in ssh-audit output.
type sshAuditAlgo struct {
	Algorithm string `json:"algorithm"`
	Notes     string `json:"notes"`
}

// sshFingerprint represents a host key fingerprint in ssh-audit output.
type sshFingerprint struct {
	Hash    string `json:"hash"`
	HashAlg string `json:"hash_alg"`
	Hostkey string `json:"hostkey"`
}

// sshAuditCVE represents a CVE entry in ssh-audit output.
type sshAuditCVE struct {
	CVSSv2      float64 `json:"cvssv2"`
	Description string  `json:"description"`
	Name        string  `json:"name"`
}

type workerConfig struct {
	timeout   int
	extraArgs string
	policy    string
	level     string
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
		slog.Info("ssh-audit worker listening", "addr", ":50051")
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
			Name:               "ssh-audit-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"service"},
			OutputTypes:        []string{"ssh_audit_result"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 60},
					"extra_args": {"type": "string", "description": "Additional ssh-audit CLI arguments"},
					"policy": {"type": "string", "description": "Path to custom policy file for compliance checking"},
					"level": {"type": "string", "description": "Minimum output level: info, warn, fail", "default": "info"}
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
		timeout: 60,
		level:   "info",
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

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

		if v, ok := fields["policy"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.policy = s
			}
		}

		if v, ok := fields["level"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.level = s
			}
		}
	}

	h.config = cfg
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("ssh-audit worker initialized",
		"timeout", cfg.timeout,
		"level", cfg.level,
		"policy", cfg.policy,
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

	// Only process SSH services; silently skip non-SSH services
	if !strings.EqualFold(serviceName, "ssh") {
		slog.Debug("skipping non-SSH service", "host", host, "port", port, "service_name", serviceName)
		return
	}

	// Build the ssh-audit command arguments
	args := []string{"-j", host, "-p", strconv.Itoa(port)}

	if h.config.level != "" && h.config.level != "info" {
		args = append(args, "-l", h.config.level)
	}

	if h.config.policy != "" {
		args = append(args, "-P", h.config.policy)
	}

	if h.config.extraArgs != "" {
		extraParts := strings.Fields(h.config.extraArgs)
		args = append(args, extraParts...)
	}

	slog.Info("running ssh-audit", "host", host, "port", port, "args", args)

	// Run ssh-audit with configurable timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "ssh-audit", args...)
	output, err := cmd.Output()
	if err != nil {
		// ssh-audit may exit with non-zero even on partial success; check if we got output
		if len(output) == 0 {
			h.sendError(stream, item.GetId(), fmt.Sprintf("ssh-audit failed for %s:%d: %v", host, port, err), true)
			return
		}
		slog.Warn("ssh-audit exited with error but produced output", "host", host, "port", port, "error", err)
	}

	// Try to parse JSON output
	var auditResult sshAuditOutput
	if err := json.Unmarshal(output, &auditResult); err != nil {
		slog.Warn("failed to parse ssh-audit JSON output, falling back to raw output", "host", host, "port", port, "error", err)

		// Fallback: run without -j flag and store raw text output
		rawItem := h.buildRawResultItem(ctx, host, port, item)
		if rawItem != nil {
			h.safeSend(stream, &workerv1.ProcessResponse{
				Payload: &workerv1.ProcessResponse_Item{Item: rawItem},
			})
			h.itemsProduced.Add(1)
		} else {
			h.sendError(stream, item.GetId(), fmt.Sprintf("ssh-audit failed to produce parseable output for %s:%d", host, port), true)
		}
		return
	}

	// Build result DataItem from parsed JSON
	resultItem := h.buildResultItem(&auditResult, host, port, item)

	h.safeSend(stream, &workerv1.ProcessResponse{
		Payload: &workerv1.ProcessResponse_Item{Item: resultItem},
	})
	h.itemsProduced.Add(1)
}

func (h *workerHandler) buildRawResultItem(ctx context.Context, host string, port int, parentItem *portwhinev1.DataItem) *portwhinev1.DataItem {
	// Re-run ssh-audit without -j flag to get raw text output
	args := []string{host, "-p", strconv.Itoa(port)}

	if h.config.level != "" && h.config.level != "info" {
		args = append(args, "-l", h.config.level)
	}

	if h.config.policy != "" {
		args = append(args, "-P", h.config.policy)
	}

	if h.config.extraArgs != "" {
		extraParts := strings.Fields(h.config.extraArgs)
		args = append(args, extraParts...)
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "ssh-audit", args...)
	output, err := cmd.Output()
	if err != nil && len(output) == 0 {
		slog.Error("ssh-audit raw fallback also failed", "host", host, "port", port, "error", err)
		return nil
	}

	data := &structpb.Struct{Fields: map[string]*structpb.Value{
		"host":       structpb.NewStringValue(host),
		"port":       structpb.NewNumberValue(float64(port)),
		"raw_output": structpb.NewStringValue(string(output)),
		"grade":      structpb.NewStringValue("unknown"),
	}}

	return &portwhinev1.DataItem{
		Id:            uuid.New().String(),
		PipelineRunId: h.pipelineRunID,
		Type:          "ssh_audit_result",
		Data:          data,
		Metadata: &portwhinev1.DataItemMetadata{
			Source:    "ssh-audit-worker",
			CreatedAt: timestamppb.Now(),
			NodeId:    h.nodeID,
			Labels: map[string]string{
				"worker_type": "ssh-audit",
				"grade":       "unknown",
				"raw_output":  "true",
			},
		},
		ParentIds: []string{parentItem.GetId()},
	}
}

func (h *workerHandler) buildResultItem(audit *sshAuditOutput, host string, port int, parentItem *portwhinev1.DataItem) *portwhinev1.DataItem {
	// Build fingerprints list: "hash_alg:hash"
	fingerprintValues := make([]*structpb.Value, 0, len(audit.Fingerprints))
	for _, fp := range audit.Fingerprints {
		fingerprintValues = append(fingerprintValues, structpb.NewStringValue(fmt.Sprintf("%s:%s", fp.HashAlg, fp.Hash)))
	}

	// Build kex_algorithms list
	kexValues := make([]*structpb.Value, 0, len(audit.Kex))
	for _, k := range audit.Kex {
		kexValues = append(kexValues, structpb.NewStringValue(k.Algorithm))
	}

	// Build encryption_algorithms list
	encValues := make([]*structpb.Value, 0, len(audit.Enc))
	for _, e := range audit.Enc {
		encValues = append(encValues, structpb.NewStringValue(e.Algorithm))
	}

	// Build mac_algorithms list
	macValues := make([]*structpb.Value, 0, len(audit.Mac))
	for _, m := range audit.Mac {
		macValues = append(macValues, structpb.NewStringValue(m.Algorithm))
	}

	// Build compression list
	compressionValues := make([]*structpb.Value, 0, len(audit.Compression))
	for _, c := range audit.Compression {
		compressionValues = append(compressionValues, structpb.NewStringValue(c))
	}

	// Build CVEs list: "CVE-name (cvss: X.X): description"
	cveValues := make([]*structpb.Value, 0, len(audit.CVEs))
	for _, cve := range audit.CVEs {
		cveStr := fmt.Sprintf("%s (cvss: %.1f): %s", cve.Name, cve.CVSSv2, cve.Description)
		cveValues = append(cveValues, structpb.NewStringValue(cveStr))
	}

	// Compute grade
	grade := computeGrade(audit)

	data := &structpb.Struct{Fields: map[string]*structpb.Value{
		"host":                  structpb.NewStringValue(host),
		"port":                  structpb.NewNumberValue(float64(port)),
		"banner":                structpb.NewStringValue(audit.Banner),
		"fingerprints":          structpb.NewListValue(&structpb.ListValue{Values: fingerprintValues}),
		"kex_algorithms":        structpb.NewListValue(&structpb.ListValue{Values: kexValues}),
		"encryption_algorithms": structpb.NewListValue(&structpb.ListValue{Values: encValues}),
		"mac_algorithms":        structpb.NewListValue(&structpb.ListValue{Values: macValues}),
		"compression":           structpb.NewListValue(&structpb.ListValue{Values: compressionValues}),
		"cves":                  structpb.NewListValue(&structpb.ListValue{Values: cveValues}),
		"grade":                 structpb.NewStringValue(grade),
	}}

	return &portwhinev1.DataItem{
		Id:            uuid.New().String(),
		PipelineRunId: h.pipelineRunID,
		Type:          "ssh_audit_result",
		Data:          data,
		Metadata: &portwhinev1.DataItemMetadata{
			Source:    "ssh-audit-worker",
			CreatedAt: timestamppb.Now(),
			NodeId:    h.nodeID,
			Labels: map[string]string{
				"worker_type": "ssh-audit",
				"grade":       grade,
			},
		},
		ParentIds: []string{parentItem.GetId()},
	}
}

// computeGrade determines the SSH security grade based on CVEs and algorithm quality.
// "A" if no CVEs and no weak algos, "B" if minor issues, "C" if CVEs present, "F" if critical CVEs.
func computeGrade(audit *sshAuditOutput) string {
	hasCriticalCVE := false
	hasCVE := false

	for _, cve := range audit.CVEs {
		hasCVE = true
		if cve.CVSSv2 >= 9.0 {
			hasCriticalCVE = true
			break
		}
	}

	if hasCriticalCVE {
		return "F"
	}
	if hasCVE {
		return "C"
	}

	// Check for weak algorithms by inspecting notes fields
	hasWeakAlgo := false
	allAlgos := make([]sshAuditAlgo, 0, len(audit.Kex)+len(audit.Enc)+len(audit.Mac)+len(audit.Key))
	allAlgos = append(allAlgos, audit.Kex...)
	allAlgos = append(allAlgos, audit.Enc...)
	allAlgos = append(allAlgos, audit.Mac...)
	allAlgos = append(allAlgos, audit.Key...)

	for _, algo := range allAlgos {
		notesLower := strings.ToLower(algo.Notes)
		if strings.Contains(notesLower, "fail") || strings.Contains(notesLower, "weak") ||
			strings.Contains(notesLower, "deprecated") || strings.Contains(notesLower, "insecure") {
			hasWeakAlgo = true
			break
		}
	}

	if hasWeakAlgo {
		return "B"
	}

	return "A"
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
