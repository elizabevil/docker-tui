package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestBeginAuditUsesStableViewAndConnectionName(t *testing.T) {
	i18n.Init("zh")
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Audit = audit.NewService(nil)
	app.ActivePanel = state.PanelContainers
	app.RuntimeType = "docker"
	app.Pool = dockerclient.NewPool()
	app.Pool.AddHost(dockerclient.HostEntry{Name: "staging", Host: "unix:///tmp/docker.sock"})
	app.Pool.SetActive("staging")

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
	app.Audit = audit.NewService(nil)
	app.Mode = state.ModeImagePull
	app.FilterText = "nginx:alpine"
	app.FilterCursor = len(app.FilterText)

	updated := handleImagePullInput(keys.KeyEnter, app)
	if updated.PendingImagePull != "nginx:alpine" || !updated.PendingImagePullAudit.Valid() {
		t.Fatalf("pending pull=%q audit=%#v", updated.PendingImagePull, updated.PendingImagePullAudit)
	}
	records := updated.Audit.RecentOperations()
	if len(records) != 2 || records[0].TraceID != records[1].TraceID || records[0].Action != "resource.image.pull" {
		t.Fatalf("pull records=%#v", records)
	}
}

func TestConfirmCancelCompletesAuditTrace(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Audit = audit.NewService(nil)
	app.Mode = state.ModeConfirm
	app.ConfirmAction = "container-stop"
	app.ConfirmAudit = beginAudit(app, "resource.container.stop", audit.ContainerTarget{ID: "one", Name: "api"}, "Stop api")

	updated, _ := handleConfirmKeys(keys.KeyEsc, app)
	records := updated.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultCancelled || records[2].TraceID != records[0].TraceID {
		t.Fatalf("cancel records=%#v", records)
	}
	if updated.ConfirmAudit.Valid() {
		t.Fatal("confirm audit trace was not cleared")
	}
}

func TestUIMessageUsesNotificationWithoutAuditHistory(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Audit = audit.NewService(nil)
	ShowToastWarn(app, "check input")
	if app.ToastMessage != "check input" {
		t.Fatalf("toast=%q", app.ToastMessage)
	}
	if records := app.Audit.RecentOperations(); len(records) != 0 {
		t.Fatalf("UI message leaked into audit history: %#v", records)
	}
}
