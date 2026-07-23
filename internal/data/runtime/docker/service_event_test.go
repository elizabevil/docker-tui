package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/events"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func TestMapDockerEventPreservesRuntimeNeutralFields(t *testing.T) {
	event := mapDockerEvent(events.Message{
		Type: events.ContainerEventType, Action: events.ActionStart,
		Actor: events.Actor{ID: "container-id", Attributes: map[string]string{"name": "api"}},
		Scope: "local", Time: 10, TimeNano: 20,
	})
	if event.ResourceType != "container" || event.Action != "start" || event.ActorID != "container-id" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if event.Attributes["name"] != "api" || event.TimeNano != 20 {
		t.Fatalf("event metadata was not preserved: %#v", event)
	}
}

func TestSendEventItemStopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if runtimeapi.SendEventItem(ctx, make(chan runtimeapi.EventItem), runtimeapi.EventItem{}) {
		t.Fatal("sendEventItem reported delivery after cancellation")
	}
}
