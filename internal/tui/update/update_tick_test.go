package update

import (
	"errors"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/runtime"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime/docker"
	"github.com/elizabevil/docker-tui/internal/data/runtime/mockengine"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestCursorBlinkTickTogglesAndReschedules(t *testing.T) {
	app := &state.AppModel{}
	if _, cmd := handleCursorBlinkTick(app); cmd == nil || !app.CursorBlinkHidden {
		t.Fatalf("first cursor tick: hidden=%v cmd=%v", app.CursorBlinkHidden, cmd)
	}
	if _, cmd := handleCursorBlinkTick(app); cmd == nil || app.CursorBlinkHidden {
		t.Fatalf("second cursor tick: hidden=%v cmd=%v", app.CursorBlinkHidden, cmd)
	}
}

func TestStatsTickKeepsSchedulerAliveOutsideContainers(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), mockengine.New(), "test")
	app.Navigation.ActivePanel = state.PanelImages
	if _, cmd := handleStatsTick(app, state.StatsTick{}); cmd == nil {
		t.Fatal("stats scheduler stopped after leaving containers")
	}
	if app.Metrics.StatsActive {
		t.Fatal("stats fetch must be inactive outside containers")
	}
}

func TestTopRefreshSchedulesAndRejectsStaleTick(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), mockengine.New(), "test")
	app.Navigation.Mode = state.ModeTop
	app.Processes.Open("c1", "api")
	app.Processes.Loading = false
	msg := state.ContainerProcessesLoaded{ContainerID: "c1", Processes: runtime.ContainerProcesses{}}
	if _, cmd := handleContainerProcessesLoaded(app, msg); cmd == nil {
		t.Fatal("top result did not schedule refresh")
	}
	stale := state.ContainerProcessesTick{ContainerID: "c1", Generation: app.Processes.Generation + 1}
	if _, cmd := handleContainerProcessesTick(app, stale); cmd != nil {
		t.Fatal("stale top tick scheduled a request")
	}
}

func TestToastTickStopsWhenIdleAndIgnoresStaleGeneration(t *testing.T) {
	app := &state.AppModel{}
	app.Feedback.ShowToast("current", state.NotificationInfo, 2)

	if _, cmd := handleToastTick(app, state.ToastTick{Generation: app.Feedback.ToastGeneration - 1}); cmd != nil {
		t.Fatal("stale toast tick scheduled another tick")
	}
	if app.Feedback.ToastTimer != 2 {
		t.Fatalf("stale toast tick changed timer: %d", app.Feedback.ToastTimer)
	}

	generation := app.Feedback.ToastGeneration
	if _, cmd := handleToastTick(app, state.ToastTick{Generation: generation}); cmd == nil {
		t.Fatal("active toast did not schedule its remaining tick")
	}
	if _, cmd := handleToastTick(app, state.ToastTick{Generation: generation}); cmd != nil {
		t.Fatal("expired toast kept the timer alive")
	}
	if app.Feedback.ToastTimer != 0 || app.Feedback.ToastMessage != "" {
		t.Fatalf("expired toast state = %#v", app.Feedback)
	}
}

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
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
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
	if updated.Feedback.InfoMessage != "" {
		t.Fatalf("unexpected info message=%q", updated.Feedback.InfoMessage)
	}
}

func TestRuntimeHealthTransitionsAtThresholdAndRecovers(t *testing.T) {
	cfg := config.DefaultAppConfig()
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
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Connection.Engine = &dockerclient.Client{}
	app.Connection.Connected = true
	app.Connection.ConnectionTarget = "podman"
	updated, _ := handleRuntimeHealthResult(app, state.RuntimeHealthResult{Name: "docker", Error: errors.New("late timeout")})
	if updated.Connection.HealthFailures != 0 || !updated.Connection.Connected {
		t.Fatalf("stale result changed active connection: %#v", updated)
	}
}
