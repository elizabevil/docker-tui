package docker

import (
	"testing"

	podmanapi "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

func TestDetectHostKeepsExplicitEndpointAndDriver(t *testing.T) {
	host, runtimeType := detectHost(ClientConfig{Host: "tcp://podman.example:2376", Runtime: "podman"})
	if host != "tcp://podman.example:2376" || runtimeType != RuntimePodman {
		t.Fatalf("host=%q runtime=%q", host, runtimeType)
	}
}

func TestNewClientRejectsUnverifiedTLSBeforeConnecting(t *testing.T) {
	_, err := NewClient(ClientConfig{
		Host:    "tcp://docker.example:2376",
		Runtime: "docker",
		TLS:     TLSConfig{Enabled: true, Verify: false},
	})
	if err == nil {
		t.Fatal("expected unverified TLS to be rejected")
	}
}

func TestTLSConfigAllowsExplicitInsecureSkipVerify(t *testing.T) {
	config, err := tlsConfig(TLSConfig{Enabled: true, InsecureSkipVerify: true}, "tcp://runtime.example:2376")
	if err != nil {
		t.Fatalf("tlsConfig() error = %v", err)
	}
	if !config.InsecureSkipVerify {
		t.Fatal("expected certificate verification to be explicitly disabled")
	}
}

func TestConnectionPoolPreservesRuntimeAndTLS(t *testing.T) {
	pool := NewPool()
	pool.AddHost(HostEntry{
		Name:    "remote",
		Host:    "tcp://example:2376",
		Runtime: "podman",
		TLS:     TLSConfig{Enabled: true, Verify: true, CAFile: "/ca.pem"},
	})
	entry := pool.Get("remote")
	if entry == nil || entry.Runtime != "podman" || entry.TLS.CAFile != "/ca.pem" {
		t.Fatalf("entry=%#v", entry)
	}
}

func TestPodmanTransportSelectionPreservesConnectionSemantics(t *testing.T) {
	rest, err := podmanapi.NewRESTClient(podmanapi.RESTConfig{Endpoint: "http://podman.test", APIVersion: "5.0.0"})
	if err != nil {
		t.Fatalf("NewRESTClient() error = %v", err)
	}

	tests := []struct {
		name   string
		client Client
		want   bool
	}{
		{name: "local bindings", client: Client{podmanREST: rest}},
		{name: "TLS REST", client: Client{podmanREST: rest, TLS: TLSConfig{Enabled: true}}, want: true},
		{name: "versioned REST", client: Client{podmanREST: rest, APIVersion: "5.0.0"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.client.usePodmanRESTTransport(); got != tt.want {
				t.Fatalf("usePodmanRESTTransport() = %v, want %v", got, tt.want)
			}
		})
	}
}
