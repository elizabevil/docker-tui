package update

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestContainerActionResultCompletesAuditProjection(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Dependencies.Audit = audit.NewService(nil)
	trace := app.Dependencies.Audit.Begin("resource.container.start", audit.ContainerTarget{ID: "short", Name: "api"}, audit.RuntimeContext{}, audit.UIContext{}, "Starting api")

	updated, _ := handleContainerActioned(app, state.ContainerActioned{Action: state.ActionStarted, ID: "short", Success: true, Audit: trace})
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultSucceeded || records[2].TraceID != trace.ID {
		t.Fatalf("action records=%#v", records)
	}
	if updated.Feedback.ToastMessage == "" || updated.Feedback.AuditOperationMessage == "" {
		t.Fatalf("audit projection missing: toast=%q operation=%q", updated.Feedback.ToastMessage, updated.Feedback.AuditOperationMessage)
	}
}

func TestContainerActionFailureUsesPersistentErrorRail(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Dependencies.Audit = audit.NewService(nil)
	trace := app.Dependencies.Audit.Begin("resource.container.start", audit.ContainerTarget{ID: "short", Name: "api"}, audit.RuntimeContext{}, audit.UIContext{}, "Starting api")

	updated, _ := handleContainerActioned(app, state.ContainerActioned{
		Action: state.ActionStarted,
		ID:     "short",
		Error:  errors.New("port 8080 is already allocated"),
		Audit:  trace,
	})

	if !strings.Contains(updated.Feedback.ErrorMessage, "started") ||
		!strings.Contains(updated.Feedback.ErrorMessage, "port 8080 is already allocated") {
		t.Fatalf("persistent error missing action or cause: %q", updated.Feedback.ErrorMessage)
	}
	if updated.Feedback.ToastLevel != state.NotificationError {
		t.Fatalf("toast level = %v, want error", updated.Feedback.ToastLevel)
	}
}

func TestContainerDetailLoadedWritesRawJSON(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	title := "Container Detail: api"
	app.Detail.Open(title, "")
	app.Navigation.Mode = state.ModeDetail

	updated, _ := handleContainerDetailLoaded(app, state.ContainerDetailLoaded{
		ContainerID: "abc123",
		Title:       title,
		Detail:      &runtimeapi.ContainerDetail{ID: "abc123", Name: "api"},
	})

	if updated.Detail.ContainerDetail == nil || updated.Detail.DetailResourceType != state.ResourceContainer {
		t.Fatalf("container detail not applied: %#v", updated.Detail)
	}
	if len(updated.Detail.DetailRawJSON) == 0 {
		t.Fatalf("container raw json not written: %#v", updated.Detail)
	}
}

func TestHistoryLoadedIgnoresStaleResponse(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.History.Open("current", "current:latest")

	updated, _ := handleHistoryLoaded(app, state.HistoryLoadedMsg{
		ImageID: "stale",
		Layers:  []runtimeapi.ImageHistoryLayer{{ID: "old"}},
	})
	if !updated.History.Loading || len(updated.History.Layers) != 0 {
		t.Fatalf("stale response changed history: %+v", updated.History)
	}

	updated, _ = handleHistoryLoaded(app, state.HistoryLoadedMsg{
		ImageID: "current",
		Layers:  []runtimeapi.ImageHistoryLayer{{ID: "new"}},
	})
	if updated.History.Loading || len(updated.History.Layers) != 1 || updated.History.Layers[0].ID != "new" {
		t.Fatalf("current response not applied: %+v", updated.History)
	}
}

func TestContainerWaitDoneIgnoresStaleGeneration(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	_, generation := app.ContainerWait.Begin()
	updated, _ := handleContainerWaitDone(app, keyboard.ContainerWaitDone{Generation: generation + 1})
	if !updated.ContainerWait.Current(generation) {
		t.Fatal("stale wait response stopped the current wait")
	}
}

func TestContainerWaitCancellationCompletesCurrentGeneration(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Dependencies.Audit = audit.NewService(nil)
	_, generation := app.ContainerWait.Begin()
	trace := app.Dependencies.Audit.Begin("resource.container.wait", audit.ContainerTarget{ID: "abc"}, audit.RuntimeContext{}, audit.UIContext{}, "Waiting")
	app.ContainerWait.Stop()

	updated, _ := handleContainerWaitDone(app, keyboard.ContainerWaitDone{
		Generation: generation,
		Error:      context.Canceled,
		Audit:      trace,
	})
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultCancelled {
		t.Fatalf("cancelled wait records=%#v", records)
	}
}

func TestVolumeDetailLoadedShowsErrorPlaceholder(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	title := "Volume Detail: data"
	app.Detail.Open(title, "")
	app.Navigation.Mode = state.ModeDetail

	updated, _ := handleVolumeDetailLoaded(app, state.VolumeDetailLoaded{
		VolumeID: "data",
		Title:    title,
		Error:    errors.New("boom"),
	})

	if !strings.Contains(updated.Detail.ImageDetailContent, "Inspect failed") {
		t.Fatalf("volume error placeholder missing: %#v", updated.Detail)
	}
}

func TestNetworkDetailLoadedWritesRawJSON(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	title := "Network Detail: net0"
	app.Detail.Open(title, "")
	app.Navigation.Mode = state.ModeDetail

	updated, _ := handleNetworkDetailLoaded(app, state.NetworkDetailLoaded{
		NetworkID: "net0",
		Title:     title,
		Detail:    &runtimeapi.NetworkDetail{ID: "net0", Name: "net0"},
	})

	if updated.Detail.NetworkDetail == nil || updated.Detail.DetailResourceType != state.ResourceNetwork {
		t.Fatalf("network detail not applied: %#v", updated.Detail)
	}
	if len(updated.Detail.DetailRawJSON) == 0 {
		t.Fatalf("network raw json not written: %#v", updated.Detail)
	}
}

func TestBatchProgressedTracksLiveProgress(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	updated, _ := handleBatchProgressed(app, state.BatchProgressed{
		Scope:   "container.batch.stop",
		Action:  "stop",
		Total:   5,
		Current: 2,
		Success: 2,
	})
	if updated.Feedback.BatchProgress == nil {
		t.Fatalf("BatchProgress must be set")
	}
	if updated.Feedback.BatchProgress.Current != 2 || updated.Feedback.BatchProgress.Total != 5 {
		t.Fatalf("progress = %#v, want current=2 total=5", updated.Feedback.BatchProgress)
	}
	if updated.Feedback.BatchProgress.Success != 2 {
		t.Fatalf("success count = %d, want 2", updated.Feedback.BatchProgress.Success)
	}
}

func TestBatchActionedClearsProgress(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Feedback.BatchProgress = &state.BatchProgressState{Scope: "container.batch.stop", Total: 2, Current: 2}
	updated, _ := handleBatchActioned(app, state.BatchActioned{
		Scope:    "container.batch.stop",
		Resource: state.ResourceContainer,
		Total:    2,
		Success:  2,
	})
	if updated.Feedback.BatchProgress != nil {
		t.Fatalf("BatchProgress must be cleared after completion")
	}
	if !strings.Contains(updated.Feedback.ToastMessage, "succeeded 2") {
		t.Fatalf("summary toast = %q, want succeeded 2", updated.Feedback.ToastMessage)
	}
}

func TestBatchActionedPartialShowsAllCounts(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	updated, _ := handleBatchActioned(app, state.BatchActioned{
		Scope:    "container.batch.kill",
		Resource: state.ResourceContainer,
		Total:    5,
		Success:  3,
		Skipped:  1,
		Failed:   1,
		Error:    errors.New("boom"),
	})
	if !strings.Contains(updated.Feedback.ToastMessage, "succeeded 3") ||
		!strings.Contains(updated.Feedback.ToastMessage, "skipped 1") ||
		!strings.Contains(updated.Feedback.ToastMessage, "failed 1") {
		t.Fatalf("summary toast = %q", updated.Feedback.ToastMessage)
	}
}
