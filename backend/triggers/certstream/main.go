package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"connectrpc.com/connect"
	certstream "github.com/CaliDog/certstream-go"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
	triggerv1 "github.com/portwhine/portwhine/gen/go/portwhine/trigger/v1"
	"github.com/portwhine/portwhine/gen/go/portwhine/trigger/v1/triggerv1connect"
	"github.com/portwhine/portwhine/pkg/server"
)

// triggerHandler implements the TriggerServiceHandler interface.
type triggerHandler struct {
	triggerv1connect.UnimplementedTriggerServiceHandler

	mu            sync.Mutex
	nodeID        string
	pipelineRunID string

	// Domain matching config
	exactDomains  map[string]bool   // exact domain lookups
	wildcardSuffs []string          // wildcard patterns converted to suffix matches
	regexPatterns []*regexp.Regexp  // compiled regex patterns

	eventsEmitted atomic.Uint64
	stopCh        chan struct{}
	initialized   bool
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	handler := &triggerHandler{}

	mux := http.NewServeMux()
	path, h := triggerv1connect.NewTriggerServiceHandler(handler)
	mux.Handle(path, h)

	server.MustListenAndServe(mux)
}

func (h *triggerHandler) GetCapabilities(_ context.Context, _ *connect.Request[triggerv1.GetCapabilitiesRequest]) (*connect.Response[triggerv1.GetCapabilitiesResponse], error) {
	return connect.NewResponse(&triggerv1.GetCapabilitiesResponse{
		Capability: &portwhinev1.TriggerCapability{
			Name:        "certstream-trigger",
			Version:     "1.0.0",
			OutputTypes: []string{"domain"},
			ConfigSchema: `{
				"type": "object",
				"properties": {
					"domains": {"type": "array", "items": {"type": "string"}, "description": "Exact domains or wildcard patterns (*.example.com)"},
					"regex_patterns": {"type": "array", "items": {"type": "string"}, "description": "Go regex patterns to match against certificate domains"}
				}
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
	h.exactDomains = make(map[string]bool)
	h.wildcardSuffs = nil
	h.regexPatterns = nil

	// Parse domains
	if domainsVal, ok := fields["domains"]; ok {
		domainList := domainsVal.GetListValue()
		if domainList != nil {
			for _, v := range domainList.GetValues() {
				d := strings.TrimSpace(strings.ToLower(v.GetStringValue()))
				if d == "" {
					continue
				}
				if strings.HasPrefix(d, "*.") {
					// Wildcard: *.example.com → match any subdomain of example.com
					suffix := d[1:] // ".example.com"
					h.wildcardSuffs = append(h.wildcardSuffs, suffix)
				} else {
					h.exactDomains[d] = true
				}
			}
		}
	}

	// Parse regex patterns
	if regexVal, ok := fields["regex_patterns"]; ok {
		regexList := regexVal.GetListValue()
		if regexList != nil {
			for _, v := range regexList.GetValues() {
				pattern := strings.TrimSpace(v.GetStringValue())
				if pattern == "" {
					continue
				}
				compiled, err := regexp.Compile(pattern)
				if err != nil {
					return connect.NewResponse(&triggerv1.InitializeResponse{
						Success:      false,
						ErrorMessage: fmt.Sprintf("invalid regex pattern %q: %v", pattern, err),
					}), nil
				}
				h.regexPatterns = append(h.regexPatterns, compiled)
			}
		}
	}

	if len(h.exactDomains) == 0 && len(h.wildcardSuffs) == 0 && len(h.regexPatterns) == 0 {
		return connect.NewResponse(&triggerv1.InitializeResponse{
			Success:      false,
			ErrorMessage: "at least one of 'domains' or 'regex_patterns' must be provided",
		}), nil
	}

	h.stopCh = make(chan struct{})
	h.initialized = true

	slog.Info("certstream trigger initialized",
		"exact_domains", len(h.exactDomains),
		"wildcard_patterns", len(h.wildcardSuffs),
		"regex_patterns", len(h.regexPatterns),
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
	nodeID := h.nodeID
	runID := h.pipelineRunID
	h.mu.Unlock()

	slog.Info("certstream trigger started, connecting to certstream...")

	// Open certstream connection
	certCh, errCh := certstream.CertStreamEventStream(true) // skip heartbeats

	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-stopCh:
			slog.Info("certstream trigger stopped",
				"events_emitted", h.eventsEmitted.Load())
			return nil

		case jq := <-certCh:
			// Extract message type
			msgType, err := jq.String("message_type")
			if err != nil || msgType != "certificate_update" {
				continue
			}

			// Extract all domains from the certificate
			allDomainsInterface, err := jq.ArrayOfStrings("data", "leaf_cert", "all_domains")
			if err != nil || len(allDomainsInterface) == 0 {
				continue
			}

			// Check each domain against our filters
			var matchedDomains []string
			for _, domain := range allDomainsInterface {
				domain = strings.ToLower(strings.TrimSpace(domain))
				if h.matchesDomain(domain) {
					matchedDomains = append(matchedDomains, domain)
				}
			}

			if len(matchedDomains) == 0 {
				continue
			}

			// Extract additional cert info
			issuer, _ := jq.String("data", "leaf_cert", "subject", "O")
			fingerprint, _ := jq.String("data", "leaf_cert", "fingerprint")

			// Build list of all domains as protobuf list
			allDomainValues := make([]*structpb.Value, len(allDomainsInterface))
			for i, d := range allDomainsInterface {
				allDomainValues[i] = structpb.NewStringValue(d)
			}

			// Emit one DataItem per matched domain
			for _, domain := range matchedDomains {
				item := &portwhinev1.DataItem{
					Id:            uuid.New().String(),
					PipelineRunId: runID,
					Type:          "domain",
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"domain":      structpb.NewStringValue(domain),
							"all_domains": structpb.NewListValue(&structpb.ListValue{Values: allDomainValues}),
							"issuer":      structpb.NewStringValue(issuer),
							"source":      structpb.NewStringValue("certstream"),
							"fingerprint": structpb.NewStringValue(fingerprint),
						},
					},
					Metadata: &portwhinev1.DataItemMetadata{
						Source:    "certstream-trigger",
						CreatedAt: timestamppb.Now(),
						NodeId:   nodeID,
						Labels: map[string]string{
							"trigger_type": "certstream",
						},
					},
				}

				if err := stream.Send(&triggerv1.StartResponse{
					Payload: &triggerv1.StartResponse_Item{Item: item},
				}); err != nil {
					return fmt.Errorf("send item: %w", err)
				}
				h.eventsEmitted.Add(1)
			}

		case err := <-errCh:
			slog.Error("certstream error", "error", err)
			// Send non-fatal error to operator
			_ = stream.Send(&triggerv1.StartResponse{
				Payload: &triggerv1.StartResponse_Error{
					Error: &triggerv1.TriggerError{
						ErrorMessage: fmt.Sprintf("certstream connection error: %v", err),
						Fatal:        false,
					},
				},
			})

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

// matchesDomain checks whether a domain matches any configured filter.
func (h *triggerHandler) matchesDomain(domain string) bool {
	// Check exact match
	if h.exactDomains[domain] {
		return true
	}

	// Check wildcard suffix matches (*.example.com → .example.com suffix)
	for _, suffix := range h.wildcardSuffs {
		if strings.HasSuffix(domain, suffix) {
			return true
		}
	}

	// Check regex patterns
	for _, re := range h.regexPatterns {
		if re.MatchString(domain) {
			return true
		}
	}

	return false
}
