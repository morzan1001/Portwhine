package worker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/http2"

	"connectrpc.com/connect"
	portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"
	workerv1 "github.com/portwhine/portwhine/gen/go/portwhine/worker/v1"
	"github.com/portwhine/portwhine/gen/go/portwhine/worker/v1/workerv1connect"
	"github.com/portwhine/portwhine/internal/pipeline"
)

// Client is a ConnectRPC client that communicates with a worker container.
// It implements the pipeline.WorkerClient interface.
type Client struct {
	rpc    workerv1connect.WorkerServiceClient
	logger *slog.Logger
}

// compile-time assertion that Client implements pipeline.WorkerClient.
var _ pipeline.WorkerClient = (*Client)(nil)

// NewClient creates a new worker Client that connects to the given address
// using ConnectRPC over HTTP/2.
func NewClient(address string, logger *slog.Logger) (*Client, error) {
	if address == "" {
		return nil, errors.New("worker address must not be empty")
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

	rpcClient := workerv1connect.NewWorkerServiceClient(
		httpClient,
		address,
	)

	return &Client{
		rpc:    rpcClient,
		logger: logger,
	}, nil
}

// Initialize calls the worker's Initialize RPC with the provided StageConfig.
func (c *Client) Initialize(ctx context.Context, config *portwhinev1.StageConfig) error {
	req := connect.NewRequest(&workerv1.InitializeRequest{
		Config: config,
	})

	resp, err := c.rpc.Initialize(ctx, req)
	if err != nil {
		return fmt.Errorf("worker initialize RPC failed: %w", err)
	}

	if !resp.Msg.GetSuccess() {
		return fmt.Errorf("worker initialization failed: %s", resp.Msg.GetErrorMessage())
	}

	c.logger.Debug("worker initialized successfully")
	return nil
}

// Process opens a bidirectional Process stream to the worker. It concurrently
// sends DataItems from the input channel and receives results into the output
// channel. The output channel is closed when the stream ends or an error occurs.
func (c *Client) Process(ctx context.Context, input <-chan *portwhinev1.DataItem, output chan<- *portwhinev1.DataItem, onHeartbeat pipeline.HeartbeatFunc) error {
	stream := c.rpc.Process(ctx)

	var (
		sendErr error
		recvErr error
		wg      sync.WaitGroup
	)

	// Sender goroutine: reads items from input channel and sends them on the stream.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			// Close the send side of the stream to signal the worker
			// that no more input is coming.
			if err := stream.CloseRequest(); err != nil {
				c.logger.Warn("failed to close send side of process stream",
					slog.Any("error", err),
				)
			}
		}()

		for {
			select {
			case item, ok := <-input:
				if !ok {
					// Input channel closed; no more items to send.
					return
				}
				req := &workerv1.ProcessRequest{
					Payload: &workerv1.ProcessRequest_Item{
						Item: item,
					},
				}
				if err := stream.Send(req); err != nil {
					sendErr = fmt.Errorf("failed to send item to worker: %w", err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Receiver goroutine: reads responses from the stream and dispatches
	// DataItems to the output channel. The caller owns the output channel
	// lifecycle — we never close it here.
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			resp, err := stream.Receive()
			if err != nil {
				if errors.Is(err, io.EOF) {
					// Stream ended normally.
					return
				}
				recvErr = fmt.Errorf("failed to receive from worker stream: %w", err)
				return
			}

			switch payload := resp.GetPayload().(type) {
			case *workerv1.ProcessResponse_Item:
				if payload.Item != nil {
					select {
					case output <- payload.Item:
					case <-ctx.Done():
						return
					}
				}
			case *workerv1.ProcessResponse_Heartbeat:
				c.logger.Debug("worker heartbeat received",
					slog.String("worker_id", payload.Heartbeat.GetWorkerId()),
					slog.Uint64("items_processed", payload.Heartbeat.GetItemsProcessed()),
				)
				if onHeartbeat != nil {
					onHeartbeat(
						payload.Heartbeat.GetItemsProcessed(),
						payload.Heartbeat.GetItemsProduced(),
						payload.Heartbeat.GetErrorsCount(),
					)
				}
			case *workerv1.ProcessResponse_Error:
				c.logger.Error("worker reported processing error",
					slog.String("source_item_id", payload.Error.GetSourceItemId()),
					slog.String("error_message", payload.Error.GetErrorMessage()),
					slog.Bool("retryable", payload.Error.GetRetryable()),
				)
			default:
				c.logger.Warn("unknown process response payload type")
			}
		}
	}()

	// Wait for both goroutines to finish.
	wg.Wait()

	// Close the response side of the stream.
	if err := stream.CloseResponse(); err != nil {
		c.logger.Warn("failed to close response side of process stream",
			slog.Any("error", err),
		)
	}

	// Return the first error encountered.
	if sendErr != nil {
		return sendErr
	}
	if recvErr != nil {
		return recvErr
	}
	return nil
}

// Shutdown calls the worker's Shutdown RPC to gracefully terminate it.
func (c *Client) Shutdown(ctx context.Context) error {
	req := connect.NewRequest(&workerv1.ShutdownRequest{})

	_, err := c.rpc.Shutdown(ctx, req)
	if err != nil {
		return fmt.Errorf("worker shutdown RPC failed: %w", err)
	}

	c.logger.Debug("worker shut down successfully")
	return nil
}

// NewWorkerClientFactory returns a pipeline.WorkerClientFactory that creates
// ConnectRPC-based worker clients with the given logger.
func NewWorkerClientFactory(logger *slog.Logger) pipeline.WorkerClientFactory {
	return func(address string) (pipeline.WorkerClient, error) {
		return NewClient(address, logger)
	}
}

// NewTLSWorkerClientFactory returns a pipeline.TLSWorkerClientFactory that creates
// ConnectRPC-based worker clients with mTLS. Certificates are provided per-call
// because they are ephemeral and change per pipeline run.
func NewTLSWorkerClientFactory(logger *slog.Logger) pipeline.TLSWorkerClientFactory {
	return func(address, serverName string, caCert, cert, key []byte) (pipeline.WorkerClient, error) {
		return NewTLSClient(address, serverName, caCert, cert, key, logger)
	}
}

// NewTLSClient creates a new worker Client that connects using mTLS.
// serverName is the container's DNS name used for TLS certificate verification.
func NewTLSClient(address, serverName string, caCert, clientCert, clientKey []byte, logger *slog.Logger) (*Client, error) {
	if address == "" {
		return nil, errors.New("worker address must not be empty")
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

	rpcClient := workerv1connect.NewWorkerServiceClient(httpClient, address)

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
