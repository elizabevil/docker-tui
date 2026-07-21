package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/footer"
)

func TestHandleFilterExitTimeoutKeepsFilterActive(t *testing.T) {
	app := &state.AppModel{NavigationState: state.NavigationState{Mode: state.ModeFilter, FilterExitPending: true, FilterExitToken: 2, FilterInput: state.QueryInputState{Text: "api", Cursor: 3}}}
	updated, _ := handleFilterExitTimeout(app, state.FilterExitTimeout{Token: 2})
	if updated.FilterExitPending || updated.Mode != state.ModeFilter || updated.FilterInput.Text != "api" {
		t.Fatalf("updated=%#v", updated)
	}
}

func TestHandleFilterExitTimeoutIgnoresStaleWindow(t *testing.T) {
	app := &state.AppModel{NavigationState: state.NavigationState{Mode: state.ModeFilter, FilterExitPending: true, FilterExitToken: 3}}
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

func TestRuntimeHealthTransitionsAtThresholdAndRecovers(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Runtime.Health.FailureThreshold = 2
	app := state.NewAppModel(cfg, nil, "test")
	app.Docker = &dockerclient.Client{}
	app.Connected = true
	app.ConnectionTarget = "local-docker"

	updated, _ := handleRuntimeHealthResult(app, state.RuntimeHealthResult{Name: "local-docker", Error: errors.New("timeout")})
	if updated.HealthDegraded || !updated.Connected || updated.HealthFailures != 1 {
		t.Fatalf("first failure degraded connection: %#v", updated)
	}
	updated, _ = handleRuntimeHealthResult(updated, state.RuntimeHealthResult{Name: "local-docker", Error: errors.New("timeout")})
	if !updated.HealthDegraded || updated.Connected || updated.ErrorCount != 1 {
		t.Fatalf("threshold did not degrade connection: %#v", updated)
	}
	updated, _ = handleRuntimeHealthResult(updated, state.RuntimeHealthResult{Name: "local-docker", Error: errors.New("timeout")})
	if updated.ErrorCount != 1 {
		t.Fatalf("repeated failure emitted another transition: errors=%d", updated.ErrorCount)
	}
	updated, _ = handleRuntimeHealthResult(updated, state.RuntimeHealthResult{Name: "local-docker"})
	if updated.HealthDegraded || !updated.Connected || updated.HealthFailures != 0 || updated.ConnectionError != "" {
		t.Fatalf("successful ping did not recover connection: %#v", updated)
	}
	if !strings.Contains(updated.ToastMessage, "recovered") {
		t.Fatalf("recovery toast=%q", updated.ToastMessage)
	}
}

func TestRuntimeHealthIgnoresStaleConnectionResult(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Docker = &dockerclient.Client{}
	app.Connected = true
	app.ConnectionTarget = "podman"
	updated, _ := handleRuntimeHealthResult(app, state.RuntimeHealthResult{Name: "docker", Error: errors.New("late timeout")})
	if updated.HealthFailures != 0 || !updated.Connected {
		t.Fatalf("stale result changed active connection: %#v", updated)
	}
}
