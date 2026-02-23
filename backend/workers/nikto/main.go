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
)

// niktoOutput represents the JSON output from nikto for a single host.
type niktoOutput struct {
	Host            string              `json:"host"`
	IP              string              `json:"ip"`
	Port            string              `json:"port"`
	Banner          string              `json:"banner"`
	Vulnerabilities []niktoVulnerability `json:"vulnerabilities"`
}

// niktoVulnerability represents a single finding from nikto.
type niktoVulnerability struct {
	ID         string `json:"id"`
	OSVDBID    string `json:"OSVDB"`
	Method     string `json:"method"`
	URL        string `json:"url"`
	Message    string `json:"msg"`
	References string `json:"references"`
}

type workerConfig struct {
	tuning    string
	timeout   int
	plugins   string
	extraArgs string
	maxTime   int
	userAgent string
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
		slog.Info("nikto worker listening", "addr", ":50051")
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
			Name:               "nikto-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"url"},
			OutputTypes:        []string{"vulnerability"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"tuning": {"type": "string", "description": "Nikto tuning options (e.g. '123bde' to select specific tests)"},
					"timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 300},
					"plugins": {"type": "string", "description": "Specific plugins to run"},
					"extra_args": {"type": "string", "description": "Additional nikto CLI arguments"},
					"max_time": {"type": "number", "description": "Max time per host in seconds (0 = unlimited)", "default": 0},
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
		timeout: 300,
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["tuning"]; ok {
			cfg.tuning = v.GetStringValue()
		}

		if v, ok := fields["timeout"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.timeout = int(n)
			}
		}

		if v, ok := fields["plugins"]; ok {
			cfg.plugins = v.GetStringValue()
		}

		if v, ok := fields["extra_args"]; ok {
			cfg.extraArgs = v.GetStringValue()
		}

		if v, ok := fields["max_time"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.maxTime = int(n)
			}
		}

		if v, ok := fields["user_agent"]; ok {
			cfg.userAgent = v.GetStringValue()
		}
	}

	h.config = cfg
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("nikto worker initialized",
		"tuning", cfg.tuning,
		"timeout", cfg.timeout,
		"plugins", cfg.plugins,
		"max_time", cfg.maxTime,
		"user_agent", cfg.userAgent,
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

	if item.GetType() != "url" {
		h.sendError(stream, item.GetId(), fmt.Sprintf("unsupported item type: %s", item.GetType()), false)
		return
	}

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

	slog.Info("running nikto", "url", targetURL)

	// Create a temp file path for nikto JSON output
	tmpFile := fmt.Sprintf("/tmp/nikto-%s.json", uuid.New().String())

	// Build command arguments
	args := []string{"-h", targetURL, "-Format", "json", "-output", tmpFile}
	if h.config.tuning != "" {
		args = append(args, "-Tuning", h.config.tuning)
	}
	if h.config.plugins != "" {
		args = append(args, "-Plugins", h.config.plugins)
	}
	if h.config.maxTime > 0 {
		args = append(args, "-maxtime", strconv.Itoa(h.config.maxTime))
	}
	if h.config.userAgent != "" {
		args = append(args, "-useragent", h.config.userAgent)
	}
	if h.config.extraArgs != "" {
		extraParts := strings.Fields(h.config.extraArgs)
		args = append(args, extraParts...)
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "nikto", args...)
	cmdOutput, cmdErr := cmd.CombinedOutput()

	// Try to read and parse the JSON output file
	var vulnerabilities []niktoVulnerability
	jsonParsed := false

	jsonData, readErr := os.ReadFile(tmpFile)
	// Clean up temp file regardless
	_ = os.Remove(tmpFile)

	if readErr == nil && len(jsonData) > 0 {
		// Try parsing as a single niktoOutput object
		var singleOutput niktoOutput
		if err := json.Unmarshal(jsonData, &singleOutput); err == nil {
			vulnerabilities = singleOutput.Vulnerabilities
			jsonParsed = true
		} else {
			// Try parsing as an array of niktoOutput
			var arrayOutput []niktoOutput
			if err := json.Unmarshal(jsonData, &arrayOutput); err == nil {
				for _, o := range arrayOutput {
					vulnerabilities = append(vulnerabilities, o.Vulnerabilities...)
				}
				jsonParsed = true
			}
		}
	}

	if !jsonParsed {
		// If JSON parsing failed entirely, check if nikto itself failed
		if cmdErr != nil && len(cmdOutput) == 0 {
			h.sendError(stream, item.GetId(), fmt.Sprintf("nikto failed for %s: %v", targetURL, cmdErr), true)
			return
		}

		// Fall back to storing raw text output as a single vulnerability item
		slog.Warn("nikto JSON parsing failed, falling back to raw output",
			"url", targetURL,
			"read_err", readErr,
			"cmd_err", cmdErr,
		)

		rawOutput := string(cmdOutput)
		if rawOutput == "" {
			rawOutput = "nikto produced no parseable output"
		}

		resultItem := &portwhinev1.DataItem{
			Id:            uuid.New().String(),
			PipelineRunId: h.pipelineRunID,
			Type:          "vulnerability",
			Data: &structpb.Struct{Fields: map[string]*structpb.Value{
				"scanner":    structpb.NewStringValue("nikto"),
				"target_url": structpb.NewStringValue(targetURL),
				"raw_output": structpb.NewStringValue(rawOutput),
				"message":    structpb.NewStringValue("nikto scan completed but JSON output could not be parsed"),
				"severity":   structpb.NewStringValue("info"),
			}},
			Metadata: &portwhinev1.DataItemMetadata{
				Source:    "nikto-worker",
				CreatedAt: timestamppb.Now(),
				NodeId:    h.nodeID,
				Labels: map[string]string{
					"worker_type": "nikto",
				},
			},
			ParentIds: []string{item.GetId()},
		}

		h.safeSend(stream, &workerv1.ProcessResponse{
			Payload: &workerv1.ProcessResponse_Item{Item: resultItem},
		})
		h.itemsProduced.Add(1)
		return
	}

	if cmdErr != nil {
		slog.Warn("nikto exited with error but produced output", "url", targetURL, "error", cmdErr)
	}

	slog.Info("nikto completed",
		"url", targetURL,
		"vulnerabilities_count", len(vulnerabilities),
	)

	// Emit each vulnerability as a separate DataItem
	for _, vuln := range vulnerabilities {
		severity := classifySeverity(vuln)

		// Build the full URL for this finding
		findingURL := vuln.URL
		if findingURL != "" && !strings.HasPrefix(findingURL, "http") {
			findingURL = strings.TrimRight(targetURL, "/") + "/" + strings.TrimLeft(findingURL, "/")
		}

		resultItem := &portwhinev1.DataItem{
			Id:            uuid.New().String(),
			PipelineRunId: h.pipelineRunID,
			Type:          "vulnerability",
			Data: &structpb.Struct{Fields: map[string]*structpb.Value{
				"scanner":    structpb.NewStringValue("nikto"),
				"nikto_id":   structpb.NewStringValue(vuln.ID),
				"osvdb_id":   structpb.NewStringValue(vuln.OSVDBID),
				"method":     structpb.NewStringValue(vuln.Method),
				"url":        structpb.NewStringValue(findingURL),
				"message":    structpb.NewStringValue(vuln.Message),
				"target_url": structpb.NewStringValue(targetURL),
				"severity":   structpb.NewStringValue(severity),
			}},
			Metadata: &portwhinev1.DataItemMetadata{
				Source:    "nikto-worker",
				CreatedAt: timestamppb.Now(),
				NodeId:    h.nodeID,
				Labels: map[string]string{
					"worker_type": "nikto",
					"severity":    severity,
				},
			},
			ParentIds: []string{item.GetId()},
		}

		h.safeSend(stream, &workerv1.ProcessResponse{
			Payload: &workerv1.ProcessResponse_Item{Item: resultItem},
		})
		h.itemsProduced.Add(1)
	}
}

// classifySeverity computes severity from OSVDB presence, references, and message content.
func classifySeverity(vuln niktoVulnerability) string {
	msg := strings.ToLower(vuln.Message)
	refs := strings.ToLower(vuln.References)

	// Critical severity: actively exploitable, direct compromise
	criticalKeywords := []string{
		"remote code execution", "rce", "command injection",
		"arbitrary file upload", "backdoor", "webshell", "web shell",
		"unauthenticated admin", "default credentials",
		"authentication bypass", "auth bypass",
		"server-side request forgery", "ssrf",
		"deserializ",
	}
	for _, kw := range criticalKeywords {
		if strings.Contains(msg, kw) || strings.Contains(refs, kw) {
			return "critical"
		}
	}

	// High severity: significant vulnerabilities
	highKeywords := []string{
		"sql injection", "sqli", "xss", "cross-site scripting",
		"cross-site request forgery", "csrf",
		"local file inclusion", "lfi", "remote file inclusion", "rfi",
		"arbitrary file", "path traversal", "directory traversal",
		"file disclosure", "source code disclosure",
		"shell", "root access", "admin password",
		"buffer overflow", "heap overflow", "stack overflow",
		"xml external entity", "xxe",
		"insecure direct object", "idor",
		"cve-", "privilege escalation",
	}
	for _, kw := range highKeywords {
		if strings.Contains(msg, kw) || strings.Contains(refs, kw) {
			return "high"
		}
	}

	// Medium severity: OSVDB references or security misconfigurations
	if vuln.OSVDBID != "" && vuln.OSVDBID != "0" {
		return "medium"
	}
	mediumKeywords := []string{
		"directory listing", "directory indexing", "index of /",
		"phpinfo", "server-status", "server-info",
		"backup file", ".bak", ".old", ".orig", ".save",
		"configuration file", "config file exposed",
		"version disclosure", "banner grabbing",
		"clickjacking", "x-frame-options",
		"cors misconfiguration", "open redirect",
		"session fixation", "cookie without",
		"http verb tampering", "trace method", "track method",
	}
	for _, kw := range mediumKeywords {
		if strings.Contains(msg, kw) {
			return "medium"
		}
	}

	// Low severity: informational security findings
	lowKeywords := []string{
		"uncommon header", "unusual header",
		"allowed http methods", "options method",
		"x-powered-by", "server leaks", "server header",
		"etag", "retrieved", "debug",
		"default file", "default page", "welcome page",
		"robots.txt", "sitemap.xml",
		"favicon", "icon found",
	}
	for _, kw := range lowKeywords {
		if strings.Contains(msg, kw) {
			return "low"
		}
	}

	return "info"
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
