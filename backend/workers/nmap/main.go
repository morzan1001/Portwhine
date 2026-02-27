package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
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
	"github.com/portwhine/portwhine/pkg/dataitem"
	"github.com/portwhine/portwhine/pkg/server"
)

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// nmap XML output types

type nmapRun struct {
	XMLName xml.Name   `xml:"nmaprun"`
	Hosts   []nmapHost `xml:"host"`
}

type nmapHost struct {
	Addresses []nmapAddress `xml:"address"`
	Ports     nmapPorts     `xml:"ports"`
}

type nmapAddress struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

type nmapPorts struct {
	Ports []nmapPort `xml:"port"`
}

type nmapPort struct {
	Protocol string      `xml:"protocol,attr"`
	PortID   int         `xml:"portid,attr"`
	State    nmapState   `xml:"state"`
	Service  nmapService `xml:"service"`
}

type nmapState struct {
	State string `xml:"state,attr"`
}

type nmapService struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
}

// Worker config

type workerConfig struct {
	ports            string
	scanType         string
	extraArgs        string
	serviceDetection bool
	topPorts         int
	timeout          int
	maxRetries       int
	minRate          int
	timing           string
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
			Name:               "nmap-worker",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"ip_address", "domain"},
			OutputTypes:        []string{"service", "url"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"ports": {"type": "string", "description": "Port range (e.g. 1-1000, 80,443)", "default": "1-1000"},
					"scan_type": {"type": "string", "description": "Nmap scan type flag (e.g. -sT, -sS)", "default": "-sT"},
					"extra_args": {"type": "string", "description": "Additional nmap arguments"},
					"service_detection": {"type": "boolean", "description": "Enable -sV service fingerprinting", "default": true},
					"top_ports": {"type": "number", "description": "Scan top N ports (overrides ports)"},
					"timeout": {"type": "number", "description": "Per-host timeout in seconds", "default": 300},
					"max_retries": {"type": "number", "description": "Max retries per probe", "default": 1},
					"min_rate": {"type": "number", "description": "Min packets per second"},
					"timing": {"type": "string", "description": "Timing template (e.g. -T4)"}
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
		ports:            "1-1000",
		scanType:         "-sT",
		serviceDetection: true,
		timeout:          300,
		maxRetries:       1,
	}

	params := config.GetParameters()
	if params != nil {
		fields := params.GetFields()

		if v, ok := fields["ports"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.ports = s
			}
		}

		if v, ok := fields["scan_type"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.scanType = s
			}
		}

		if v, ok := fields["extra_args"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.extraArgs = s
			}
		}

		if v, ok := fields["service_detection"]; ok {
			cfg.serviceDetection = v.GetBoolValue()
		}

		if v, ok := fields["top_ports"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.topPorts = int(n)
			}
		}

		if v, ok := fields["timeout"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.timeout = int(n)
			}
		}

		if v, ok := fields["max_retries"]; ok {
			if n := v.GetNumberValue(); n >= 0 {
				cfg.maxRetries = int(n)
			}
		}

		if v, ok := fields["min_rate"]; ok {
			if n := v.GetNumberValue(); n > 0 {
				cfg.minRate = int(n)
			}
		}

		if v, ok := fields["timing"]; ok {
			if s := v.GetStringValue(); s != "" {
				cfg.timing = s
			}
		}
	}

	h.config = cfg
	h.status = portwhinev1.WorkerStatus_WORKER_STATUS_READY
	h.initialized = true

	slog.Info("nmap worker initialized",
		"ports", cfg.ports,
		"scan_type", cfg.scanType,
		"service_detection", cfg.serviceDetection,
		"extra_args", cfg.extraArgs,
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

	var target, hostname, domain string

	switch item.GetType() {
	case "ip_address":
		if item.GetData() != nil {
			if v, ok := item.GetData().GetFields()["ip"]; ok {
				target = v.GetStringValue()
			}
			if v, ok := item.GetData().GetFields()["domain"]; ok {
				domain = v.GetStringValue()
			}
		}
		if target == "" {
			h.sendError(stream, item.GetId(), "missing 'ip' field in data", false)
			return
		}
		if net.ParseIP(target) == nil {
			h.sendError(stream, item.GetId(), fmt.Sprintf("invalid IP address: %s", target), false)
			return
		}
		hostname = target
		if domain != "" {
			hostname = domain
		}

	case "domain":
		if item.GetData() != nil {
			if v, ok := item.GetData().GetFields()["domain"]; ok {
				target = v.GetStringValue()
			}
		}
		if target == "" {
			h.sendError(stream, item.GetId(), "missing 'domain' field in data", false)
			return
		}
		if !domainRegex.MatchString(target) {
			h.sendError(stream, item.GetId(), fmt.Sprintf("invalid domain: %s", target), false)
			return
		}
		hostname = target
		domain = target

	default:
		h.sendError(stream, item.GetId(), fmt.Sprintf("unsupported item type: %s", item.GetType()), false)
		return
	}

	// Build nmap command arguments
	args := []string{h.config.scanType}

	if h.config.serviceDetection {
		args = append(args, "-sV")
	}

	if h.config.topPorts > 0 {
		args = append(args, "--top-ports", strconv.Itoa(h.config.topPorts))
	} else {
		args = append(args, "-p", h.config.ports)
	}

	if h.config.maxRetries > 0 {
		args = append(args, "--max-retries", strconv.Itoa(h.config.maxRetries))
	}

	if h.config.minRate > 0 {
		args = append(args, "--min-rate", strconv.Itoa(h.config.minRate))
	}

	if h.config.timing != "" {
		args = append(args, h.config.timing)
	}

	if h.config.extraArgs != "" {
		extraParts := strings.Fields(h.config.extraArgs)
		args = append(args, extraParts...)
	}

	args = append(args, "-oX", "-", target)

	scanCtx, cancel := context.WithTimeout(ctx, time.Duration(h.config.timeout)*time.Second)
	defer cancel()

	slog.Debug("running nmap", "args", args, "target", target)

	cmd := exec.CommandContext(scanCtx, "nmap", args...)
	output, err := cmd.Output()
	if err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("nmap execution failed for %s: %v", target, err), true)
		return
	}

	var result nmapRun
	if err := xml.Unmarshal(output, &result); err != nil {
		h.sendError(stream, item.GetId(), fmt.Sprintf("failed to parse nmap XML output for %s: %v", target, err), true)
		return
	}

	for _, host := range result.Hosts {
		ip := target
		for _, addr := range host.Addresses {
			if addr.AddrType == "ipv4" || addr.AddrType == "ipv6" {
				ip = addr.Addr
				break
			}
		}

		for _, port := range host.Ports.Ports {
			if port.State.State != "open" {
				continue
			}

			// Build and emit a service DataItem for every open port.
			serviceFields := dataitem.BuildServiceData(
				ip, port.PortID, port.Protocol,
				port.Service.Name, port.Service.Product, port.Service.Version,
				hostname,
			)

			serviceItem := &portwhinev1.DataItem{
				Id:            uuid.New().String(),
				PipelineRunId: h.pipelineRunID,
				Type:          "service",
				Data:          &structpb.Struct{Fields: serviceFields},
				Metadata: &portwhinev1.DataItemMetadata{
					Source:    "nmap-worker",
					CreatedAt: timestamppb.Now(),
					NodeId:    h.nodeID,
					Labels: map[string]string{
						"worker_type":  "nmap",
						"scan_type":    h.config.scanType,
						"service_name": port.Service.Name,
					},
				},
				ParentIds: []string{item.GetId()},
			}

			h.safeSend(stream, &workerv1.ProcessResponse{
				Payload: &workerv1.ProcessResponse_Item{Item: serviceItem},
			})
			h.itemsProduced.Add(1)

			// For HTTP services, also emit a url DataItem with the base URL.
			if dataitem.IsHTTPService(port.Service.Name) {
				scheme := dataitem.DetermineScheme(port.Service.Name, port.PortID)
				fullURL := dataitem.BuildURL(scheme, hostname, port.PortID)

				urlItem := &portwhinev1.DataItem{
					Id:            uuid.New().String(),
					PipelineRunId: h.pipelineRunID,
					Type:          "url",
					Data: &structpb.Struct{Fields: map[string]*structpb.Value{
						"url": structpb.NewStringValue(fullURL),
					}},
					Metadata: &portwhinev1.DataItemMetadata{
						Source:    "nmap-worker",
						CreatedAt: timestamppb.Now(),
						NodeId:    h.nodeID,
						Labels: map[string]string{
							"worker_type": "nmap",
							"source_type": "port_scan",
						},
					},
					ParentIds: []string{item.GetId()},
				}

				h.safeSend(stream, &workerv1.ProcessResponse{
					Payload: &workerv1.ProcessResponse_Item{Item: urlItem},
				})
				h.itemsProduced.Add(1)
			}
		}
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
