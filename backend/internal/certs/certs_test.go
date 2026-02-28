package certs

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"
)

func TestGenerateRunCA(t *testing.T) {
	caCertPEM, caKeyPEM, err := GenerateRunCA("test-run-12345678")
	if err != nil {
		t.Fatalf("GenerateRunCA failed: %v", err)
	}

	// Decode and parse CA cert.
	block, _ := pem.Decode(caCertPEM)
	if block == nil {
		t.Fatal("failed to decode CA cert PEM")
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse CA cert: %v", err)
	}

	if !caCert.IsCA {
		t.Error("CA cert should have IsCA=true")
	}
	if caCert.Subject.CommonName != "portwhine-run-test-run-ca" {
		t.Errorf("unexpected CN: %s", caCert.Subject.CommonName)
	}
	if time.Until(caCert.NotAfter) > 2*time.Hour {
		t.Errorf("CA cert lifetime too long: expires at %v", caCert.NotAfter)
	}

	// Decode and parse CA key.
	keyBlock, _ := pem.Decode(caKeyPEM)
	if keyBlock == nil {
		t.Fatal("failed to decode CA key PEM")
	}
	caKey, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse CA key: %v", err)
	}
	if caKey.Curve.Params().BitSize != 256 {
		t.Errorf("expected P-256, got %d-bit curve", caKey.Curve.Params().BitSize)
	}
}

func TestGenerateNodeCert(t *testing.T) {
	caCertPEM, caKeyPEM, err := GenerateRunCA("test-run-12345678")
	if err != nil {
		t.Fatalf("GenerateRunCA failed: %v", err)
	}

	certPEM, keyPEM, err := GenerateNodeCert(caCertPEM, caKeyPEM, "portwhine-test-worker")
	if err != nil {
		t.Fatalf("GenerateNodeCert failed: %v", err)
	}

	// Decode and parse node cert.
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("failed to decode node cert PEM")
	}
	nodeCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse node cert: %v", err)
	}

	if nodeCert.IsCA {
		t.Error("node cert should not be CA")
	}
	if nodeCert.Subject.CommonName != "portwhine-test-worker" {
		t.Errorf("unexpected CN: %s", nodeCert.Subject.CommonName)
	}

	// Check SANs.
	expectedDNS := map[string]bool{"portwhine-test-worker": false, "localhost": false}
	for _, name := range nodeCert.DNSNames {
		expectedDNS[name] = true
	}
	for name, found := range expectedDNS {
		if !found {
			t.Errorf("missing DNS SAN: %s", name)
		}
	}

	// Check ExtKeyUsage includes both server and client auth.
	hasServer, hasClient := false, false
	for _, usage := range nodeCert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth {
			hasServer = true
		}
		if usage == x509.ExtKeyUsageClientAuth {
			hasClient = true
		}
	}
	if !hasServer || !hasClient {
		t.Errorf("node cert should have both ServerAuth and ClientAuth, got server=%v client=%v", hasServer, hasClient)
	}

	// Verify the cert is signed by the CA.
	caBlock, _ := pem.Decode(caCertPEM)
	caCert, _ := x509.ParseCertificate(caBlock.Bytes)
	pool := x509.NewCertPool()
	pool.AddCert(caCert)

	_, err = nodeCert.Verify(x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		t.Errorf("node cert verification failed: %v", err)
	}

	// Verify the key matches.
	keyBlock, _ := pem.Decode(keyPEM)
	nodeKey, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse node key: %v", err)
	}
	pubKey := nodeCert.PublicKey.(*ecdsa.PublicKey)
	if pubKey.X.Cmp(nodeKey.PublicKey.X) != 0 || pubKey.Y.Cmp(nodeKey.PublicKey.Y) != 0 {
		t.Error("node cert public key does not match private key")
	}
}

func TestTLSHandshake(t *testing.T) {
	// Generate a full set of certs and verify a real TLS handshake works.
	caCertPEM, caKeyPEM, err := GenerateRunCA("handshake-test-1234")
	if err != nil {
		t.Fatalf("GenerateRunCA failed: %v", err)
	}

	serverCertPEM, serverKeyPEM, err := GenerateNodeCert(caCertPEM, caKeyPEM, "localhost")
	if err != nil {
		t.Fatalf("GenerateNodeCert (server) failed: %v", err)
	}

	clientCertPEM, clientKeyPEM, err := GenerateNodeCert(caCertPEM, caKeyPEM, "operator")
	if err != nil {
		t.Fatalf("GenerateNodeCert (client) failed: %v", err)
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCertPEM)

	serverCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		t.Fatalf("server key pair: %v", err)
	}

	clientCert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
	if err != nil {
		t.Fatalf("client key pair: %v", err)
	}

	serverCfg := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
	}

	clientCfg := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caPool,
		ServerName:   "localhost",
		MinVersion:   tls.VersionTLS13,
	}

	// Create a TLS listener and dial it.
	ln, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatalf("TLS listen: %v", err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		// Force the handshake.
		if tlsConn, ok := conn.(*tls.Conn); ok {
			done <- tlsConn.Handshake()
		} else {
			done <- nil
		}
	}()

	clientConn, err := tls.Dial("tcp", ln.Addr().String(), clientCfg)
	if err != nil {
		t.Fatalf("TLS dial failed: %v", err)
	}
	defer clientConn.Close()

	if err := <-done; err != nil {
		t.Fatalf("server handshake failed: %v", err)
	}
}

func TestGenerateNodeCertInvalidCA(t *testing.T) {
	_, _, err := GenerateNodeCert([]byte("not-a-cert"), []byte("not-a-key"), "test")
	if err == nil {
		t.Error("expected error for invalid CA cert PEM")
	}
}
