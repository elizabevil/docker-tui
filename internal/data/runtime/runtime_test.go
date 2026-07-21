package runtime

import (
	"errors"
	"testing"
)

func TestCapabilitySetDefaultsToUnsupported(t *testing.T) {
	capabilities := CapabilitySet{
		CapabilityContainers: {Support: Available},
		CapabilityEvents:     {Support: Degraded, Reason: "polling fallback"},
	}

	if !capabilities.Supports(CapabilityContainers) {
		t.Fatal("available capability should be supported")
	}
	if !capabilities.Supports(CapabilityEvents) {
		t.Fatal("degraded capability should be supported")
	}
	if capabilities.Supports(CapabilityExec) {
		t.Fatal("missing capability should be unsupported")
	}
}

func TestErrorPreservesKindAndCause(t *testing.T) {
	cause := errors.New("socket unavailable")
	err := NewError(ErrorConnection, "ping", "local-podman", cause)

	if !IsErrorKind(err, ErrorConnection) {
		t.Fatalf("expected connection error, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatal("runtime error should preserve its cause")
	}
	if got := err.Error(); got != "ping: connection local-podman: socket unavailable" {
		t.Fatalf("unexpected error text: %q", got)
	}
}
