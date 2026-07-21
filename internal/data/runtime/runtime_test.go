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

func TestFilterSetCloneDoesNotShareValues(t *testing.T) {
	filters := FilterSet{"label": {"app=api"}}
	clone := filters.Clone()
	clone.Add("label", "tier=backend")

	if got := len(filters.Values("label")); got != 1 {
		t.Fatalf("clone mutated source values: got %d", got)
	}
	if got := len(clone.Values("label")); got != 2 {
		t.Fatalf("expected two cloned values, got %d", got)
	}
}
