package docker

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/utils"
)

func TestNewClientRejectsUnverifiedTLSBeforeConnecting(t *testing.T) {
	_, err := NewClient(runtime.ClientConfig{
		Host: "tcp://docker.example:2376",
		TLS:  utils.TLSConfig{Enabled: true, Verify: false},
	})
	if err == nil {
		t.Fatal("expected unverified TLS to be rejected")
	}
}

func TestTLSConfigAllowsExplicitInsecureSkipVerify(t *testing.T) {
	config, err := tlsConfig(utils.TLSConfig{Enabled: true, InsecureSkipVerify: true}, "tcp://runtime.example:2376")
	if err != nil {
		t.Fatalf("tlsConfig() error = %v", err)
	}
	if !config.InsecureSkipVerify {
		t.Fatal("expected certificate verification to be explicitly disabled")
	}
}

func TestConnectionPoolPreservesRuntimeAndTLS(t *testing.T) {
	pool := runtime.NewPool(nil)
	pool.AddHost(runtime.HostEntry{
		Name:    "remote",
		Host:    "tcp://example:2376",
		Runtime: "podman",
		TLS:     utils.TLSConfig{Enabled: true, Verify: true, CAFile: "/ca.pem"},
	})
	entry := pool.Get("remote")
	if entry == nil || entry.Runtime != "podman" || entry.TLS.CAFile != "/ca.pem" {
		t.Fatalf("entry=%#v", entry)
	}
}
