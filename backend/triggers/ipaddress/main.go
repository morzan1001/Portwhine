package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math/big"
	"net/http"
	"net/netip"
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
	triggerv1 "github.com/portwhine/portwhine/gen/go/portwhine/trigger/v1"
	"github.com/portwhine/portwhine/gen/go/portwhine/trigger/v1/triggerv1connect"
)

// ipTarget represents a parsed IP target with optional network origin.
type ipTarget struct {
	IP      netip.Addr
	Network string // source CIDR if expanded, empty for single IPs
}

// triggerHandler implements the TriggerServiceHandler interface.
type triggerHandler struct {
	triggerv1connect.UnimplementedTriggerServiceHandler

	mu             sync.Mutex
	nodeID         string
	pipelineRunID  string
	targets        []ipTarget
	intervalSec    int
	eventsEmitted  atomic.Uint64
	stopCh         chan struct{}
	initialized    bool
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	handler := &triggerHandler{}

	mux := http.NewServeMux()
	path, h := triggerv1connect.NewTriggerServiceHandler(handler)
	mux.Handle(path, h)

	srv := &http.Server{
		Addr:              ":50051",
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("ipaddress trigger listening", "addr", ":50051")
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

func (h *triggerHandler) GetCapabilities(_ context.Context, _ *connect.Request[triggerv1.GetCapabilitiesRequest]) (*connect.Response[triggerv1.GetCapabilitiesResponse], error) {
	return connect.NewResponse(&triggerv1.GetCapabilitiesResponse{
		Capability: &portwhinev1.TriggerCapability{
			Name:        "ipaddress-trigger",
			Version:     "1.0.0",
			OutputTypes: []string{"ip_address"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"targets": {"type": "array", "items": {"type": "string"}, "description": "List of IPs, CIDRs, or IP ranges (start-end)"},
					"interval_seconds": {"type": "integer", "description": "Repetition interval in seconds (0 = emit once)"}
				},
				"required": ["targets"]
			}`,
		},
	}), nil
}

func (h *triggerHandler) Initialize(_ context.Context, req *connect.Request[triggerv1.InitializeRequest]) (*connect.Response[triggerv1.InitializeResponse], error) {
	config := req.Msg.GetConfig()
	if config == nil {
		return connect.NewResponse(&triggerv1.InitializeResponse{
			Success:      false,
			ErrorMessage: "missing stage config",
		}), nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.nodeID = config.GetNodeId()
	h.pipelineRunID = config.GetPipelineRunId()

	params := config.GetParameters()
	if params == nil {
		return connect.NewResponse(&triggerv1.InitializeResponse{
			Success:      false,
			ErrorMessage: "missing parameters",
		}), nil
	}

	fields := params.GetFields()

	// Parse targets
	targetsVal, ok := fields["targets"]
	if !ok {
		return connect.NewResponse(&triggerv1.InitializeResponse{
			Success:      false,
			ErrorMessage: "missing 'targets' parameter",
		}), nil
	}

	targetList := targetsVal.GetListValue()
	if targetList == nil || len(targetList.GetValues()) == 0 {
		return connect.NewResponse(&triggerv1.InitializeResponse{
			Success:      false,
			ErrorMessage: "'targets' must be a non-empty list",
		}), nil
	}

	var targets []ipTarget
	for _, v := range targetList.GetValues() {
		raw := v.GetStringValue()
		if raw == "" {
			continue
		}

		parsed, err := parseTarget(raw)
		if err != nil {
			return connect.NewResponse(&triggerv1.InitializeResponse{
				Success:      false,
				ErrorMessage: fmt.Sprintf("invalid target %q: %v", raw, err),
			}), nil
		}
		targets = append(targets, parsed...)
	}

	if len(targets) == 0 {
		return connect.NewResponse(&triggerv1.InitializeResponse{
			Success:      false,
			ErrorMessage: "no valid targets after parsing",
		}), nil
	}

	h.targets = targets

	// Parse interval_seconds
	if intervalVal, ok := fields["interval_seconds"]; ok {
		h.intervalSec = int(intervalVal.GetNumberValue())
	}

	h.stopCh = make(chan struct{})
	h.initialized = true

	slog.Info("ipaddress trigger initialized",
		"targets", len(h.targets),
		"interval_seconds", h.intervalSec,
	)

	return connect.NewResponse(&triggerv1.InitializeResponse{Success: true}), nil
}

func (h *triggerHandler) Start(_ context.Context, _ *connect.Request[triggerv1.StartRequest], stream *connect.ServerStream[triggerv1.StartResponse]) error {
	h.mu.Lock()
	if !h.initialized {
		h.mu.Unlock()
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("trigger not initialized"))
	}
	stopCh := h.stopCh
	targets := h.targets
	intervalSec := h.intervalSec
	nodeID := h.nodeID
	runID := h.pipelineRunID
	h.mu.Unlock()

	slog.Info("ipaddress trigger started", "targets", len(targets), "interval", intervalSec)

	// Heartbeat ticker
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		// Emit all targets
		for _, t := range targets {
			select {
			case <-stopCh:
				slog.Info("ipaddress trigger stopped during emission")
				return nil
			default:
			}

			item := h.buildDataItem(t, runID, nodeID)
			if err := stream.Send(&triggerv1.StartResponse{
				Payload: &triggerv1.StartResponse_Item{Item: item},
			}); err != nil {
				return fmt.Errorf("send item: %w", err)
			}
			h.eventsEmitted.Add(1)
		}

		// Send heartbeat after emission batch
		_ = stream.Send(&triggerv1.StartResponse{
			Payload: &triggerv1.StartResponse_Heartbeat{
				Heartbeat: &triggerv1.TriggerHeartbeat{
					TriggerId:     nodeID,
					EventsEmitted: h.eventsEmitted.Load(),
				},
			},
		})

		// If no repetition, end the stream
		if intervalSec <= 0 {
			slog.Info("ipaddress trigger completed (single emission)",
				"events_emitted", h.eventsEmitted.Load())
			return nil
		}

		// Wait for interval or stop signal
		slog.Info("ipaddress trigger waiting for next cycle",
			"interval_seconds", intervalSec,
			"events_emitted", h.eventsEmitted.Load())

		timer := time.NewTimer(time.Duration(intervalSec) * time.Second)
		select {
		case <-stopCh:
			timer.Stop()
			slog.Info("ipaddress trigger stopped during wait")
			return nil
		case <-timer.C:
			// Continue to next emission cycle
		case <-heartbeatTicker.C:
			_ = stream.Send(&triggerv1.StartResponse{
				Payload: &triggerv1.StartResponse_Heartbeat{
					Heartbeat: &triggerv1.TriggerHeartbeat{
						TriggerId:     nodeID,
						EventsEmitted: h.eventsEmitted.Load(),
					},
				},
			})
		}
	}
}

func (h *triggerHandler) Stop(_ context.Context, _ *connect.Request[triggerv1.StopRequest]) (*connect.Response[triggerv1.StopResponse], error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopCh != nil {
		close(h.stopCh)
		h.stopCh = nil
	}

	return connect.NewResponse(&triggerv1.StopResponse{
		TotalEventsEmitted: h.eventsEmitted.Load(),
	}), nil
}

func (h *triggerHandler) HealthCheck(_ context.Context, _ *connect.Request[triggerv1.HealthCheckRequest]) (*connect.Response[triggerv1.HealthCheckResponse], error) {
	return connect.NewResponse(&triggerv1.HealthCheckResponse{Healthy: true}), nil
}

func (h *triggerHandler) buildDataItem(t ipTarget, runID, nodeID string) *portwhinev1.DataItem {
	ip := t.IP.String()
	version := "4"
	if t.IP.Is6() {
		version = "6"
	}

	dataFields := map[string]*structpb.Value{
		"ip":      structpb.NewStringValue(ip),
		"version": structpb.NewStringValue(version),
	}
	if t.Network != "" {
		dataFields["network"] = structpb.NewStringValue(t.Network)
	}

	return &portwhinev1.DataItem{
		Id:            uuid.New().String(),
		PipelineRunId: runID,
		Type:          "ip_address",
		Data:          &structpb.Struct{Fields: dataFields},
		Metadata: &portwhinev1.DataItemMetadata{
			Source:    "ipaddress-trigger",
			CreatedAt: timestamppb.Now(),
			NodeId:   nodeID,
			Labels: map[string]string{
				"trigger_type": "ipaddress",
			},
		},
	}
}

// parseTarget parses a single target string into IP targets.
// Supports: single IP, CIDR notation, IP ranges (start-end).
func parseTarget(raw string) ([]ipTarget, error) {
	raw = strings.TrimSpace(raw)

	// Check for IP range (contains "-" but not "/")
	if strings.Contains(raw, "-") && !strings.Contains(raw, "/") {
		return parseIPRange(raw)
	}

	// Check for CIDR
	if strings.Contains(raw, "/") {
		return parseCIDR(raw)
	}

	// Single IP
	addr, err := netip.ParseAddr(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid IP address: %w", err)
	}
	return []ipTarget{{IP: addr}}, nil
}

func parseCIDR(cidr string) ([]ipTarget, error) {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}

	var targets []ipTarget
	addr := prefix.Addr()
	for {
		if !prefix.Contains(addr) {
			break
		}
		targets = append(targets, ipTarget{
			IP:      addr,
			Network: cidr,
		})
		addr = addr.Next()
	}
	return targets, nil
}

func parseIPRange(rangeStr string) ([]ipTarget, error) {
	parts := strings.SplitN(rangeStr, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid IP range format")
	}

	startAddr, err := netip.ParseAddr(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid start IP: %w", err)
	}

	endAddr, err := netip.ParseAddr(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid end IP: %w", err)
	}

	if startAddr.Is4() != endAddr.Is4() {
		return nil, fmt.Errorf("start and end IPs must be the same version")
	}

	// Convert to big.Int for comparison
	startBytes := startAddr.As16()
	endBytes := endAddr.As16()
	startInt := new(big.Int).SetBytes(startBytes[:])
	endInt := new(big.Int).SetBytes(endBytes[:])

	if startInt.Cmp(endInt) > 0 {
		return nil, fmt.Errorf("start IP must be <= end IP")
	}

	// Limit to prevent excessive expansion
	diff := new(big.Int).Sub(endInt, startInt)
	maxRange := big.NewInt(1 << 20) // 1M IPs max
	if diff.Cmp(maxRange) > 0 {
		return nil, fmt.Errorf("IP range too large (max 1048576 IPs)")
	}

	var targets []ipTarget
	current := startAddr
	for {
		targets = append(targets, ipTarget{IP: current})
		if current == endAddr {
			break
		}
		current = current.Next()
	}
	return targets, nil
}
