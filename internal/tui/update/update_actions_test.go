package update

import (
	"errors"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestContainerActionResultCompletesAuditProjection(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
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

func TestContainerDetailLoadedWritesRawJSON(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
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

func TestVolumeDetailLoadedShowsErrorPlaceholder(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
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
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
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
