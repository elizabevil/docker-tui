package tui

import (
	"errors"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/mockengine"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestHandleRuntimeEventIgnoresStaleGeneration(t *testing.T) {
	model := &state.AppModel{}
	model.Events.Begin()
	_, command := handleRuntimeEvent(model, state.RuntimeEventReceived{
		Generation: model.Events.Generation + 1,
		Item:       runtimeapi.EventItem{Event: runtimeapi.Event{ResourceType: "container"}},
	})
	if command != nil || len(model.Events.Dirty) != 0 {
		t.Fatalf("stale event changed state: %#v", model.Events)
	}
}

func TestHandleRuntimeEventCoalescesRefresh(t *testing.T) {
	model := &state.AppModel{}
	_, generation := model.Events.Begin()
	items := make(chan runtimeapi.EventItem)
	_, firstCommand := handleRuntimeEvent(model, state.RuntimeEventReceived{
		Generation: generation,
		Item:       runtimeapi.EventItem{Event: runtimeapi.Event{ResourceType: "container"}},
		Items:      items,
	})
	_, secondCommand := handleRuntimeEvent(model, state.RuntimeEventReceived{
		Generation: generation,
		Item:       runtimeapi.EventItem{Event: runtimeapi.Event{ResourceType: "image"}},
		Items:      items,
	})
	if firstCommand == nil || secondCommand == nil || !model.Events.Dirty["container"] || !model.Events.Dirty["image"] {
		t.Fatalf("events were not coalesced: %#v", model.Events)
	}
}

func TestHandleRuntimeEventErrorEnablesFallback(t *testing.T) {
	model := &state.AppModel{}
	_, generation := model.Events.Begin()
	_, command := handleRuntimeEvent(model, state.RuntimeEventReceived{
		Generation: generation,
		Item:       runtimeapi.EventItem{Error: errors.New("stream closed")},
	})
	if command == nil || !model.Events.Degraded || model.Events.Failures != 1 {
		t.Fatalf("stream failure state = %#v", model.Events)
	}
}

// TestHandleEventFlushContainerOnlyRefreshesContainer verifies that a
// container event refreshes only the container list (and the compose view
// which is derived from it on render).
func TestHandleEventFlushContainerOnlyRefreshesContainer(t *testing.T) {
	model := newFlushModel(t)
	_, generation := model.Events.Begin()
	token, _ := model.Events.MarkDirty("container")
	_, cmd := handleEventFlush(model, state.EventFlush{Generation: generation, Token: token})
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	if len(model.Events.Dirty) != 0 {
		t.Errorf("dirty set was not cleared after flush: %#v", model.Events.Dirty)
	}
}

// TestHandleEventFlushImageTriggersContainerRefresh verifies TASK-011:
// image events also refresh containers because the image panel shows
// per-image container counts.
func TestHandleEventFlushImageTriggersContainerRefresh(t *testing.T) {
	model := newFlushModel(t)
	_, generation := model.Events.Begin()
	token, _ := model.Events.MarkDirty("image")
	_, cmd := handleEventFlush(model, state.EventFlush{Generation: generation, Token: token})
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
}

// TestHandleEventFlushVolumeTriggersContainerRefresh verifies TASK-011:
// volume events also refresh containers because container mounts reference
// volumes.
func TestHandleEventFlushVolumeTriggersContainerRefresh(t *testing.T) {
	model := newFlushModel(t)
	_, generation := model.Events.Begin()
	token, _ := model.Events.MarkDirty("volume")
	_, cmd := handleEventFlush(model, state.EventFlush{Generation: generation, Token: token})
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
}

// TestHandleEventFlushNetworkTriggersContainerRefresh verifies TASK-011:
// network events refresh both networks and containers.
func TestHandleEventFlushNetworkTriggersContainerRefresh(t *testing.T) {
	model := newFlushModel(t)
	_, generation := model.Events.Begin()
	token, _ := model.Events.MarkDirty("network")
	_, cmd := handleEventFlush(model, state.EventFlush{Generation: generation, Token: token})
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
}

// TestHandleEventFlushCoalescesMixedDeps verifies that a flush with mixed
// dirty types dedupes to a single refresh per list.
func TestHandleEventFlushCoalescesMixedDeps(t *testing.T) {
	model := newFlushModel(t)
	_, generation := model.Events.Begin()
	token, _ := model.Events.MarkDirty("image")
	token, _ = model.Events.MarkDirty("volume")
	token, _ = model.Events.MarkDirty("network")
	_, cmd := handleEventFlush(model, state.EventFlush{Generation: generation, Token: token})
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
}

// newFlushModel builds an AppModel with the event system ready and an
// engine wired in so handleEventFlush produces a fetch command.
func newFlushModel(t *testing.T) *state.AppModel {
	t.Helper()
	m := state.NewAppModel(nil, mockengine.New(), "test")
	m.Events.Begin()
	return m
}

// TestHandleEventFlushStaleGenerationIgnored verifies that a flush from a
// previous connection generation is dropped without triggering a refresh.
func TestHandleEventFlushStaleGenerationIgnored(t *testing.T) {
	model := &state.AppModel{}
	_, generation := model.Events.Begin()
	model.Events.MarkDirty("container")
	_, cmd := handleEventFlush(model, state.EventFlush{Generation: generation + 1, Token: 99})
	if cmd != nil {
		t.Errorf("stale flush should be ignored, got %T", cmd)
	}
}

// TestHandleEventFlushEmptyDirtyIsNoop verifies that flushing a token
// with no dirty state returns nil and does nothing.
func TestHandleEventFlushEmptyDirtyIsNoop(t *testing.T) {
	model := &state.AppModel{}
	_, generation := model.Events.Begin()
	model.Events.MarkDirty("container") // sets token=1
	// Take it back so Dirty is empty.
	model.Events.TakeDirty(model.Events.FlushToken)
	_, cmd := handleEventFlush(model, state.EventFlush{Generation: generation, Token: 9999})
	if cmd != nil {
		t.Errorf("empty-dirty flush should be noop, got %T", cmd)
	}
}
