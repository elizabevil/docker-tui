package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/footer"
)

func TestHandleFilterExitTimeoutKeepsFilterActive(t *testing.T) {
	app := &state.AppModel{Mode: state.ModeFilter, FilterExitPending: true, FilterExitToken: 2, FilterInput: state.QueryInputState{Text: "api", Cursor: 3}}
	updated, _ := handleFilterExitTimeout(app, state.FilterExitTimeout{Token: 2})
	if updated.FilterExitPending || updated.Mode != state.ModeFilter || updated.FilterInput.Text != "api" {
		t.Fatalf("updated=%#v", updated)
	}
}

func TestHandleFilterExitTimeoutIgnoresStaleWindow(t *testing.T) {
	app := &state.AppModel{Mode: state.ModeFilter, FilterExitPending: true, FilterExitToken: 3}
	updated, _ := handleFilterExitTimeout(app, state.FilterExitTimeout{Token: 2})
	if !updated.FilterExitPending {
		t.Fatal("stale timeout cleared current filter exit window")
	}
}

func TestHandleDockerConnectedErrorProjectsTargetAndMessage(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	updated, cmd := handleDockerConnected(app, state.DockerConnected{
		Name:  "local-docker",
		Error: errors.New("dial unix /var/run/docker.sock: no such file"),
	})
	if cmd != nil {
		t.Fatal("connection failure should not schedule resource fetch")
	}
	if updated.Connected || updated.Connecting || updated.Docker != nil {
		t.Fatalf("connection state=%#v", updated)
	}
	if updated.ConnectionTarget != "local-docker" || !strings.Contains(updated.ConnectionError, "no such file") {
		t.Fatalf("connection error=%q target=%q", updated.ConnectionError, updated.ConnectionTarget)
	}
	if updated.ToastMessage == "" || !strings.Contains(updated.ToastMessage, "local-docker") {
		t.Fatalf("toast=%q", updated.ToastMessage)
	}
	status := footer.StatusBar(updated)
	if !strings.Contains(status, "local-docker") || strings.Contains(status, "docker disconnected") && !strings.Contains(status, "local-docker") {
		t.Fatalf("status=%q", status)
	}
}
