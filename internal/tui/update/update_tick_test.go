package update

import (
	"errors"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/runtime"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime/docker"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/footer"
)

func TestHandleFilterExitTimeoutKeepsFilterActive(t *testing.T) {
	app := &state.AppModel{Navigation: state.NavigationState{Mode: state.ModeFilter, FilterExitPending: true, FilterExitToken: 2, FilterInput: state.QueryInputState{Text: "api", Cursor: 3}}}
	updated, _ := handleFilterExitTimeout(app, state.FilterExitTimeout{Token: 2})
	if updated.Navigation.FilterExitPending || updated.Navigation.Mode != state.ModeFilter || updated.Navigation.FilterInput.Text != "api" {
		t.Fatalf("updated=%#v", updated)
	}
}

func TestHandleFilterExitTimeoutIgnoresStaleWindow(t *testing.T) {
	app := &state.AppModel{Navigation: state.NavigationState{Mode: state.ModeFilter, FilterExitPending: true, FilterExitToken: 3}}
	updated, _ := handleFilterExitTimeout(app, state.FilterExitTimeout{Token: 2})
	if !updated.Navigation.FilterExitPending {
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
	if updated.Connection.Connected || updated.Connection.Connecting || updated.Connection.Engine != nil {
		t.Fatalf("connection state=%#v", updated)
	}
	if updated.Connection.ConnectionTarget != "local-docker" || updated.Connection.ConnectionFailure.Kind != runtime.ConnectionErrorUnknown {
		t.Fatalf("connection failure=%#v target=%q", updated.Connection.ConnectionFailure, updated.Connection.ConnectionTarget)
	}
	if updated.Feedback.ToastMessage == "" || !strings.Contains(updated.Feedback.ToastMessage, "local-docker") {
		t.Fatalf("toast=%q", updated.Feedback.ToastMessage)
	}
	if strings.Contains(updated.Feedback.ToastMessage, "no such file") {
		t.Fatalf("toast exposed raw connection error: %q", updated.Feedback.ToastMessage)
	}
	status := footer.StatusBar(updated)
	if !strings.Contains(status, "local-docker") || !strings.Contains(status, "Connection failed") {
		t.Fatalf("status=%q", status)
	}
}

func TestRuntimeHealthTransitionsAtThresholdAndRecovers(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Runtime.Health.FailureThreshold = 2
	app := state.NewAppModel(cfg, nil, "test")
	app.Connection.Engine = &dockerclient.Client{}
	app.Connection.Connected = true
	app.Connection.ConnectionTarget = "local-docker"

	updated, _ := handleRuntimeHealthResult(app, state.RuntimeHealthResult{Name: "local-docker", Error: errors.New("timeout")})
	if updated.Connection.HealthDegraded || !updated.Connection.Connected || updated.Connection.HealthFailures != 1 {
		t.Fatalf("first failure degraded connection: %#v", updated)
	}
	updated, _ = handleRuntimeHealthResult(updated, state.RuntimeHealthResult{Name: "local-docker", Error: errors.New("timeout")})
	if !updated.Connection.HealthDegraded || updated.Connection.Connected || updated.Feedback.ErrorCount != 1 {
		t.Fatalf("threshold did not degrade connection: %#v", updated)
	}
	updated, _ = handleRuntimeHealthResult(updated, state.RuntimeHealthResult{Name: "local-docker", Error: errors.New("timeout")})
	if updated.Feedback.ErrorCount != 1 {
		t.Fatalf("repeated failure emitted another transition: errors=%d", updated.Feedback.ErrorCount)
	}
	updated, _ = handleRuntimeHealthResult(updated, state.RuntimeHealthResult{Name: "local-docker"})
	if updated.Connection.HealthDegraded || !updated.Connection.Connected || updated.Connection.HealthFailures != 0 || updated.Connection.ConnectionFailure.Kind != "" {
		t.Fatalf("successful ping did not recover connection: %#v", updated)
	}
	if !strings.Contains(updated.Feedback.ToastMessage, "recovered") {
		t.Fatalf("recovery toast=%q", updated.Feedback.ToastMessage)
	}
}

func TestRuntimeHealthIgnoresStaleConnectionResult(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Connection.Engine = &dockerclient.Client{}
	app.Connection.Connected = true
	app.Connection.ConnectionTarget = "podman"
	updated, _ := handleRuntimeHealthResult(app, state.RuntimeHealthResult{Name: "docker", Error: errors.New("late timeout")})
	if updated.Connection.HealthFailures != 0 || !updated.Connection.Connected {
		t.Fatalf("stale result changed active connection: %#v", updated)
	}
}
