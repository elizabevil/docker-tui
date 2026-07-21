package state

import (
	"errors"
	"testing"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
)

func TestConnectionStateOwnsConnectionTransitions(t *testing.T) {
	client := &dockerclient.Client{RuntimeType: dockerclient.RuntimePodman, EngineVersion: "5.0"}
	var connection ConnectionState
	connection.Begin()
	connection.ConnectedTo("podman", client)
	if connection.Connecting || !connection.Connected || connection.Docker != client {
		t.Fatalf("connected state = %#v", connection)
	}
	if connection.RuntimeType != "podman" || connection.EngineVersion != "5.0" {
		t.Fatalf("runtime metadata = %q, %q", connection.RuntimeType, connection.EngineVersion)
	}

	connection.Failed("docker", errors.New("socket unavailable"))
	if connection.Connected || connection.Docker != nil || connection.ConnectionTarget != "docker" {
		t.Fatalf("failed state = %#v", connection)
	}
}

func TestConnectionStateKeepsActiveClientOnSelectionFailure(t *testing.T) {
	client := &dockerclient.Client{}
	connection := NewConnectionState(client)
	connection.SelectionFailed("remote", errors.New("certificate expired"))
	if connection.Docker != client || !connection.Connected {
		t.Fatal("selection failure discarded the active connection")
	}
	if connection.RuntimeSelectorError["remote"] != "certificate expired" {
		t.Fatalf("selector errors = %#v", connection.RuntimeSelectorError)
	}
}
