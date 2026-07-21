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

func TestHTTPErrorClassification(t *testing.T) {
	tests := map[int]ErrorKind{
		400: ErrorInvalid,
		401: ErrorAuthentication,
		403: ErrorPermission,
		404: ErrorNotFound,
		409: ErrorConflict,
		429: ErrorRateLimited,
		503: ErrorUnavailable,
		500: ErrorInternal,
	}
	for status, want := range tests {
		if got := ClassifyHTTPStatus(status); got != want {
			t.Errorf("status %d: got %q, want %q", status, got, want)
		}
	}
}

func TestRetryableErrorKinds(t *testing.T) {
	for _, kind := range []ErrorKind{ErrorConnection, ErrorTimeout, ErrorRateLimited, ErrorUnavailable} {
		if !IsRetryableKind(kind) {
			t.Errorf("expected %q to be retryable", kind)
		}
	}
	for _, kind := range []ErrorKind{ErrorCanceled, ErrorPermission, ErrorInvalid, ErrorConflict, ErrorUnsupported} {
		if IsRetryableKind(kind) {
			t.Errorf("expected %q not to be retryable", kind)
		}
	}
}

func TestImageListOptionsRejectsMultipleValuesForANDSemantics(t *testing.T) {
	options := ImageListOptions{Filters: FilterSet{ImageFilterLabel: {"app=api", "tier=backend"}}}
	_, err := options.NativeFilters()
	if !IsErrorKind(err, ErrorUnsupported) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestImageListOptionsCopiesNativeFilters(t *testing.T) {
	options := ImageListOptions{Filters: FilterSet{ImageFilterReference: {"example/*"}}}
	filters, err := options.NativeFilters()
	if err != nil {
		t.Fatalf("NativeFilters() error = %v", err)
	}
	filters[ImageFilterReference][0] = "changed"
	if options.Filters.Values(ImageFilterReference)[0] != "example/*" {
		t.Fatal("native filters mutated input options")
	}
}
