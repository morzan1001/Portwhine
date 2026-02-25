// Package server provides a shared gRPC/ConnectRPC server bootstrap for
// worker and trigger containers. It supports optional mTLS when the
// PORTWHINE_TLS_ENABLED environment variable is set.
package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// ListenAndServe starts a gRPC/ConnectRPC server on :50051. If PORTWHINE_TLS_ENABLED
// is set, it loads base64-encoded certificates from environment variables and enables
// mTLS. Otherwise, it serves plaintext HTTP/2 (h2c) as before.
//
// The function blocks until SIGINT/SIGTERM is received, then performs a graceful shutdown.
func ListenAndServe(handler http.Handler) error {
	addr := ":50051"

	if os.Getenv("PORTWHINE_TLS_ENABLED") == "true" {
		return listenTLS(handler, addr)
	}
	return listenPlaintext(handler, addr)
}

func listenPlaintext(handler http.Handler, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(handler, &http2.Server{}),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return runServer(srv, addr, false)
}

func listenTLS(handler http.Handler, addr string) error {
	tlsCfg, err := buildMTLSConfig()
	if err != nil {
		return fmt.Errorf("build mTLS config: %w", err)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		TLSConfig:         tlsCfg,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Configure HTTP/2 on the TLS server.
	if err := http2.ConfigureServer(srv, &http2.Server{}); err != nil {
		return fmt.Errorf("configure HTTP/2: %w", err)
	}

	return runServer(srv, addr, true)
}

func runServer(srv *http.Server, addr string, isTLS bool) error {
	errCh := make(chan error, 1)
	go func() {
		mode := "h2c"
		if isTLS {
			mode = "mTLS"
		}
		slog.Info("server listening", "addr", addr, "mode", mode)

		var err error
		if isTLS {
			// TLS certs are already in srv.TLSConfig; pass empty strings.
			err = srv.ListenAndServeTLS("", "")
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		slog.Info("received shutdown signal", "signal", sig)
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	return nil
}

// buildMTLSConfig creates a tls.Config with mutual TLS authentication from
// base64-encoded PEM certificates in environment variables.
func buildMTLSConfig() (*tls.Config, error) {
	caCertB64 := os.Getenv("PORTWHINE_TLS_CA_CERT")
	certB64 := os.Getenv("PORTWHINE_TLS_CERT")
	keyB64 := os.Getenv("PORTWHINE_TLS_KEY")

	if caCertB64 == "" || certB64 == "" || keyB64 == "" {
		return nil, fmt.Errorf("PORTWHINE_TLS_CA_CERT, PORTWHINE_TLS_CERT, and PORTWHINE_TLS_KEY must be set")
	}

	caCertPEM, err := base64.StdEncoding.DecodeString(caCertB64)
	if err != nil {
		return nil, fmt.Errorf("decode CA cert: %w", err)
	}
	certPEM, err := base64.StdEncoding.DecodeString(certB64)
	if err != nil {
		return nil, fmt.Errorf("decode cert: %w", err)
	}
	keyPEM, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse key pair: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCertPEM) {
		return nil, fmt.Errorf("failed to add CA cert to pool")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
		RootCAs:      caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// MustListenAndServe is like ListenAndServe but calls log.Fatal on error.
func MustListenAndServe(handler http.Handler) {
	if err := ListenAndServe(handler); err != nil {
		log.Fatal(err)
	}
}
