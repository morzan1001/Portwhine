package trigger

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"golang.org/x/net/http2"

	"connectrpc.com/connect"
	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
	triggerv1 "github.com/portwhine/portwhine/gen/go/portwhine/trigger/v1"
	"github.com/portwhine/portwhine/gen/go/portwhine/trigger/v1/triggerv1connect"
	"github.com/portwhine/portwhine/internal/pipeline"
)

// Client is a ConnectRPC client that communicates with a trigger container.
// It implements the pipeline.TriggerClient interface.
type Client struct {
	rpc    triggerv1connect.TriggerServiceClient
	logger *slog.Logger
}

// compile-time assertion that Client implements pipeline.TriggerClient.
var _ pipeline.TriggerClient = (*Client)(nil)

// NewClient creates a new trigger Client that connects to the given address
// using ConnectRPC over HTTP/2.
func NewClient(address string, logger *slog.Logger) (*Client, error) {
	if address == "" {
		return nil, errors.New("trigger address must not be empty")
	}

	httpClient := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
	}

	rpcClient := triggerv1connect.NewTriggerServiceClient(
		httpClient,
		address,
	)

	return &Client{
		rpc:    rpcClient,
		logger: logger,
	}, nil
}

// Initialize calls the trigger's Initialize RPC with the provided StageConfig.
func (c *Client) Initialize(ctx context.Context, config *portwhinev1.StageConfig) error {
	req := connect.NewRequest(&triggerv1.InitializeRequest{
		Config: config,
	})

	resp, err := c.rpc.Initialize(ctx, req)
	if err != nil {
		return fmt.Errorf("trigger initialize RPC failed: %w", err)
	}

	if !resp.Msg.GetSuccess() {
		return fmt.Errorf("trigger initialization failed: %s", resp.Msg.GetErrorMessage())
	}

	c.logger.Debug("trigger initialized successfully")
	return nil
}

// Start calls the trigger's Start RPC, which returns a server-side stream.
// It reads StartResponse messages from the stream, extracts DataItems, and
// writes them to the output channel. Heartbeats are logged. Errors are logged
// and, if fatal, cause the method to return an error. The output channel is
// closed when the stream ends.
func (c *Client) Start(ctx context.Context, output chan<- *portwhinev1.DataItem, onHeartbeat pipeline.HeartbeatFunc) error {
	defer close(output)

	req := connect.NewRequest(&triggerv1.StartRequest{})

	stream, err := c.rpc.Start(ctx, req)
	if err != nil {
		return fmt.Errorf("trigger start RPC failed: %w", err)
	}
	defer stream.Close()

	for stream.Receive() {
		resp := stream.Msg()

		switch payload := resp.GetPayload().(type) {
		case *triggerv1.StartResponse_Item:
			if payload.Item != nil {
				select {
				case output <- payload.Item:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		case *triggerv1.StartResponse_Heartbeat:
			c.logger.Debug("trigger heartbeat received",
				slog.String("trigger_id", payload.Heartbeat.GetTriggerId()),
				slog.Uint64("events_emitted", payload.Heartbeat.GetEventsEmitted()),
			)
			if onHeartbeat != nil {
				onHeartbeat(0, payload.Heartbeat.GetEventsEmitted(), 0)
			}
		case *triggerv1.StartResponse_Error:
			c.logger.Error("trigger reported error",
				slog.String("error_message", payload.Error.GetErrorMessage()),
				slog.Bool("fatal", payload.Error.GetFatal()),
			)
			if payload.Error.GetFatal() {
				return fmt.Errorf("trigger fatal error: %s", payload.Error.GetErrorMessage())
			}
		default:
			c.logger.Warn("unknown start response payload type")
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("trigger stream error: %w", err)
	}

	c.logger.Debug("trigger stream ended")
	return nil
}

// Stop calls the trigger's Stop RPC to gracefully stop event emission.
func (c *Client) Stop(ctx context.Context) error {
	req := connect.NewRequest(&triggerv1.StopRequest{})

	_, err := c.rpc.Stop(ctx, req)
	if err != nil {
		return fmt.Errorf("trigger stop RPC failed: %w", err)
	}

	c.logger.Debug("trigger stopped successfully")
	return nil
}

// NewTriggerClientFactory returns a pipeline.TriggerClientFactory that creates
// ConnectRPC-based trigger clients with the given logger.
func NewTriggerClientFactory(logger *slog.Logger) pipeline.TriggerClientFactory {
	return func(address string) (pipeline.TriggerClient, error) {
		return NewClient(address, logger)
	}
}

// NewTLSTriggerClientFactory returns a pipeline.TLSTriggerClientFactory that creates
// ConnectRPC-based trigger clients with mTLS. Certificates are provided per-call
// because they are ephemeral and change per pipeline run.
func NewTLSTriggerClientFactory(logger *slog.Logger) pipeline.TLSTriggerClientFactory {
	return func(address, serverName string, caCert, cert, key []byte) (pipeline.TriggerClient, error) {
		return NewTLSClient(address, serverName, caCert, cert, key, logger)
	}
}

// NewTLSClient creates a new trigger Client that connects using mTLS.
// serverName is the container's DNS name used for TLS certificate verification.
func NewTLSClient(address, serverName string, caCert, clientCert, clientKey []byte, logger *slog.Logger) (*Client, error) {
	if address == "" {
		return nil, errors.New("trigger address must not be empty")
	}

	tlsCfg, err := buildMTLSClientConfig(caCert, clientCert, clientKey, serverName)
	if err != nil {
		return nil, fmt.Errorf("build mTLS config: %w", err)
	}

	// Replace http:// with https:// for TLS.
	address = strings.Replace(address, "http://", "https://", 1)

	httpClient := &http.Client{
		Transport: &http2.Transport{
			TLSClientConfig: tlsCfg,
		},
	}

	rpcClient := triggerv1connect.NewTriggerServiceClient(httpClient, address)

	return &Client{
		rpc:    rpcClient,
		logger: logger,
	}, nil
}

func buildMTLSClientConfig(caCert, clientCert, clientKey []byte, serverName string) (*tls.Config, error) {
	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, fmt.Errorf("parse client key pair: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add CA cert to pool")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS13,
	}, nil
}
