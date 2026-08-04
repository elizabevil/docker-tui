package runtime

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/utils"
)

func TestFromRuntimeConnectionPreservesTLSFields(t *testing.T) {
	spec := FromRuntimeConnection(config.RuntimeConnection{
		Name:       "remote",
		Driver:     "podman",
		Endpoint:   "tcp://example:2376",
		APIVersion: "5.0.0",
		TLS: config.RuntimeTLSConfig{
			Enabled:    true,
			Verify:     true,
			CAFile:     "/ca.pem",
			CertFile:   "/cert.pem",
			KeyFile:    "/key.pem",
			ServerName: "example",
		},
	})

	if spec.Name != "remote" || spec.Runtime != RuntimePodman || spec.APIVersion != "5.0.0" {
		t.Fatalf("spec=%#v", spec)
	}
	if spec.TLS.ServerName != "example" || spec.TLS.CAFile != "/ca.pem" {
		t.Fatalf("tls=%#v", spec.TLS)
	}
}

func TestConnectionSpecKeyIncludesTransportIdentity(t *testing.T) {
	verified := ConnectionSpec{Runtime: RuntimeDocker, Host: "tcp://example:2376", TLS: utils.TLSConfig{Enabled: true, Verify: true, CAFile: "/ca.pem"}}
	insecure := verified
	insecure.TLS.Verify = false
	insecure.TLS.InsecureSkipVerify = true
	versioned := verified
	versioned.APIVersion = "1.48"

	if verified.Key() == insecure.Key() {
		t.Fatal("verified and insecure connections must have different keys")
	}
	if verified.Key() == versioned.Key() {
		t.Fatal("API version override must be part of the connection key")
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
