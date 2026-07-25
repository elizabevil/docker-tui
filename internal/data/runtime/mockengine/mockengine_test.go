package mockengine_test

import (
	"context"
	"errors"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/mockengine"
)

// TestNewEngineHasNoFailures verifies a fresh mock engine accepts all calls.
func TestNewEngineAcceptsAllCalls(t *testing.T) {
	m := mockengine.New()
	svc := m.Actions()
	ref := runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: "a"}
	if _, err := svc.Execute(context.Background(), ref, runtimeapi.ActionStart, runtimeapi.ActionOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := m.CallCount(); got != 1 {
		t.Errorf("expected 1 call, got %d", got)
	}
	if calls := m.Calls(); len(calls) != 1 || calls[0].ID != "a" {
		t.Errorf("calls = %v", calls)
	}
}

// TestEngineFailMarksIdAndFailAllPropagates verifies per-id and global
// failure modes both surface as errors.
func TestEngineFailAndFailAll(t *testing.T) {
	m := mockengine.New()
	m.Fail("b", errors.New("boom"))
	m.FailAll(errors.New("global"))

	svc := m.Actions()
	if _, err := svc.Execute(context.Background(), runtimeapi.ResourceRef{ID: "a"}, runtimeapi.ActionStart, runtimeapi.ActionOptions{}); err == nil {
		t.Error("FailAll did not propagate")
	}
	if _, err := svc.Execute(context.Background(), runtimeapi.ResourceRef{ID: "b"}, runtimeapi.ActionStart, runtimeapi.ActionOptions{}); err == nil {
		t.Error("Fail did not propagate")
	}
	if _, err := svc.Execute(context.Background(), runtimeapi.ResourceRef{ID: "c"}, runtimeapi.ActionStart, runtimeapi.ActionOptions{}); err == nil {
		t.Error("third (unmarked) call should have been blocked by FailAll")
	}
	if got := m.CallCount(); got != 3 {
		t.Errorf("expected 3 calls recorded, got %d", got)
	}
}
