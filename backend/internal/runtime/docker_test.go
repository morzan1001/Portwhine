package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRemoteTLSConfig_AllEmpty(t *testing.T) {
	err := validateRemoteTLSConfig(DockerTLSConfig{})
	if err == nil {
		t.Error("expected error when all TLS fields are empty")
	}
}

func TestValidateRemoteTLSConfig_PartialFields(t *testing.T) {
	err := validateRemoteTLSConfig(DockerTLSConfig{
		CACert: "/some/ca.pem",
		Cert:   "",
		Key:    "",
	})
	if err == nil {
		t.Error("expected error when cert and key are missing")
	}
}

func TestValidateRemoteTLSConfig_FileNotFound(t *testing.T) {
	err := validateRemoteTLSConfig(DockerTLSConfig{
		CACert: "/nonexistent/ca.pem",
		Cert:   "/nonexistent/cert.pem",
		Key:    "/nonexistent/key.pem",
	})
	if err == nil {
		t.Error("expected error for missing files")
	}
}

func TestValidateRemoteTLSConfig_WorldReadableKey(t *testing.T) {
	dir := t.TempDir()

	caPath := filepath.Join(dir, "ca.pem")
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	os.WriteFile(caPath, []byte("ca"), 0644)
	os.WriteFile(certPath, []byte("cert"), 0644)
	os.WriteFile(keyPath, []byte("key"), 0604) // world-readable

	err := validateRemoteTLSConfig(DockerTLSConfig{
		CACert: caPath,
		Cert:   certPath,
		Key:    keyPath,
	})
	if err == nil {
		t.Error("expected error for world-readable key")
	}
}

func TestValidateRemoteTLSConfig_ValidFiles(t *testing.T) {
	dir := t.TempDir()

	caPath := filepath.Join(dir, "ca.pem")
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")

	os.WriteFile(caPath, []byte("ca"), 0644)
	os.WriteFile(certPath, []byte("cert"), 0644)
	os.WriteFile(keyPath, []byte("key"), 0600) // restricted

	err := validateRemoteTLSConfig(DockerTLSConfig{
		CACert: caPath,
		Cert:   certPath,
		Key:    keyPath,
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseHostIP(t *testing.T) {
	tests := []struct {
		host    string
		wantIP  string
		wantErr bool
	}{
		{"tcp://192.168.1.100:2376", "192.168.1.100", false},
		{"tcp://docker.example.com:2376", "docker.example.com", false},
		{"tcp://10.0.0.1:2376", "10.0.0.1", false},
		{"://invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		ip, err := parseHostIP(tt.host)
		if tt.wantErr && err == nil {
			t.Errorf("parseHostIP(%q) expected error", tt.host)
			continue
		}
		if !tt.wantErr && err != nil {
			t.Errorf("parseHostIP(%q) unexpected error: %v", tt.host, err)
			continue
		}
		if ip != tt.wantIP {
			t.Errorf("parseHostIP(%q) = %q, want %q", tt.host, ip, tt.wantIP)
		}
	}
}

func TestIsRemote(t *testing.T) {
	local := &DockerRuntime{isRemote: false}
	remote := &DockerRuntime{isRemote: true}

	if local.IsRemote() {
		t.Error("local runtime should not be remote")
	}
	if !remote.IsRemote() {
		t.Error("remote runtime should be remote")
	}
}
