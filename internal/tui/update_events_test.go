package tui

import (
	"errors"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
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
