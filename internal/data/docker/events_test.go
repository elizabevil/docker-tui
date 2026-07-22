package docker

import (
	"testing"

	"github.com/docker/docker/api/types/events"
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
