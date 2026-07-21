package docker

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

func TestFromRuntimeConnPreservesTLSFields(t *testing.T) {
	spec := FromRuntimeConn(config.RuntimeConn{
		Name:     "remote",
		Driver:   "podman",
		Endpoint: "tcp://example:2376",
		TLS: config.RuntimeTLSConfig{
			Enabled:    true,
			Verify:     true,
			CAFile:     "/ca.pem",
			CertFile:   "/cert.pem",
			KeyFile:    "/key.pem",
			ServerName: "example",
		},
	})

	if spec.Name != "remote" || spec.Runtime != RuntimePodman {
		t.Fatalf("spec=%#v", spec)
	}
	if spec.TLS.ServerName != "example" || spec.TLS.CAFile != "/ca.pem" {
		t.Fatalf("tls=%#v", spec.TLS)
	}
}

func TestConnectionKeyNormalizesEndpoint(t *testing.T) {
	left := ConnectionKey(RuntimeDocker, "unix:///var/run/docker.sock")
	right := ConnectionKey("docker", "unix:///var/run/./docker.sock")
	if left != right {
		t.Fatalf("keys differ: %q vs %q", left, right)
	}
}

func TestNormalizeRuntimeType(t *testing.T) {
	if got := NormalizeRuntimeType("podman"); got != RuntimePodman {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeRuntimeType("unknown"); got != RuntimeDocker {
		t.Fatalf("got %q", got)
	}
}
