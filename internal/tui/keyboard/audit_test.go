package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestBeginAuditUsesStableViewAndConnectionName(t *testing.T) {
	i18n.Init("zh")
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Dependencies.Audit = audit.NewService(nil)
	app.Navigation.ActivePanel = state.PanelContainers
	app.Connection.RuntimeType = "docker"
	app.Connection.Pool = dockerclient.NewPool(nil)
	app.Connection.Pool.AddHost(dockerclient.HostEntry{Name: "staging", Host: "unix:///tmp/docker.sock"})
	app.Connection.Pool.SetActive("staging")

	trace := beginAudit(app, "resource.container.start", audit.ContainerTarget{ID: "one", Name: "api"}, "Starting api")
	if trace.Runtime.Name != "staging" || trace.Runtime.Type != "docker" {
		t.Fatalf("runtime context=%#v", trace.Runtime)
	}
	if trace.UI.View != "containers" {
		t.Fatalf("localized audit view=%q", trace.UI.View)
	}
}

func TestImagePullInputCreatesAuditTrace(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Dependencies.Audit = audit.NewService(nil)
	app.Navigation.Mode = state.ModeImagePull
	app.Dialog.Input.Set("nginx:alpine")

	updated := handleImagePullInput(keys.KeyEnter, app)
	if updated.Selection.PendingImagePull != "nginx:alpine" || !updated.Selection.PendingImagePullAudit.Valid() {
		t.Fatalf("pending pull=%q audit=%#v", updated.Selection.PendingImagePull, updated.Selection.PendingImagePullAudit)
	}
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 2 || records[0].TraceID != records[1].TraceID || records[0].Action != "resource.image.pull" {
		t.Fatalf("pull records=%#v", records)
	}
}

func TestConfirmCancelCompletesAuditTrace(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Dependencies.Audit = audit.NewService(nil)
	app.Navigation.Mode = state.ModeConfirm
	app.Confirm.ConfirmAction = "container-stop"
	app.Confirm.ConfirmAudit = beginAudit(app, "resource.container.stop", audit.ContainerTarget{ID: "one", Name: "api"}, "Stop api")

	updated, _ := handleConfirmKeys(keys.KeyEsc, app)
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultCancelled || records[2].TraceID != records[0].TraceID {
		t.Fatalf("cancel records=%#v", records)
	}
	if updated.Confirm.ConfirmAudit.Valid() {
		t.Fatal("confirm audit trace was not cleared")
	}
}

func TestUIMessageUsesNotificationWithoutAuditHistory(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Dependencies.Audit = audit.NewService(nil)
	ShowToastWarn(app, "check input")
	if app.Feedback.ToastMessage != "check input" {
		t.Fatalf("toast=%q", app.Feedback.ToastMessage)
	}
	if records := app.Dependencies.Audit.RecentOperations(); len(records) != 0 {
		t.Fatalf("UI message leaked into audit history: %#v", records)
	}
}
