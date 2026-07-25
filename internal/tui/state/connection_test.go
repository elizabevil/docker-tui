package state

import (
	"crypto/x509"
	"errors"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/runtime"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime/docker"
)

func TestConnectionStateOwnsConnectionTransitions(t *testing.T) {
	client := &dockerclient.Client{EngineVersion: "5.0"}
	var connection ConnectionState
	connection.Begin()
	connection.ConnectedTo("docker", client)
	if connection.Connecting || !connection.Connected || connection.Engine != client {
		t.Fatalf("connected state = %#v", connection)
	}
	if connection.RuntimeType != "docker" || connection.EngineVersion != "5.0" {
		t.Fatalf("runtime metadata = %q, %q", connection.RuntimeType, connection.EngineVersion)
	}

	connection.Failed("docker", errors.New("socket unavailable"))
	if connection.Connected || connection.Engine != nil || connection.ConnectionTarget != "docker" {
		t.Fatalf("failed state = %#v", connection)
	}
}

func TestConnectionStateKeepsActiveClientOnSelectionFailure(t *testing.T) {
	client := &dockerclient.Client{}
	connection := NewConnectionState(client)
	connection.SelectionFailed("remote", x509.UnknownAuthorityError{})
	if connection.Engine != client || !connection.Connected {
		t.Fatal("selection failure discarded the active connection")
	}
	if connection.RuntimeSelectorError["remote"].Kind != runtime.ConnectionErrorCA {
		t.Fatalf("selector errors = %#v", connection.RuntimeSelectorError)
	}
}
