package update

import (
	"errors"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/mockengine"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// advancedApp returns an AppModel with an audit service wired so handlers can
// begin/finish traces, mirroring the pattern in update_actions_test.go.
func advancedApp() *state.AppModel {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Dependencies.Audit = audit.NewService(nil)
	return app
}

// beginTrace opens a container-targeted audit trace on the app.
func beginTrace(app *state.AppModel, action, id string) audit.Trace {
	return app.Dependencies.Audit.Begin(action, audit.ContainerTarget{ID: id, Name: id}, audit.RuntimeContext{}, audit.UIContext{}, "test action")
}

func TestHandleContainerExportDoneSuccess(t *testing.T) {
	app := advancedApp()
	trace := beginTrace(app, "resource.container.export", "short")
	updated, cmd := handleContainerExportDone(app, keyboard.ContainerExportDone{
		ContainerID: "short",
		Destination: "/tmp/out.tar",
		Bytes:       4096,
		Audit:       trace,
	})
	if cmd != nil {
		t.Fatalf("success returned a cmd %T, want nil", cmd)
	}
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultSucceeded || records[2].TraceID != trace.ID {
		t.Fatalf("export records=%#v", records)
	}
	if updated.Feedback.ToastMessage == "" {
		t.Fatalf("success export must surface a toast, got %q", updated.Feedback.ToastMessage)
	}
}

func TestHandleContainerExportDoneFailure(t *testing.T) {
	app := advancedApp()
	trace := beginTrace(app, "resource.container.export", "short")
	want := errors.New("export refused")
	updated, cmd := handleContainerExportDone(app, keyboard.ContainerExportDone{
		ContainerID: "short",
		Error:       want,
		Audit:       trace,
	})
	if cmd != nil {
		t.Fatalf("failure returned a cmd %T, want nil", cmd)
	}
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultFailed || records[2].TraceID != trace.ID {
		t.Fatalf("export failure records=%#v", records)
	}
	if updated.Feedback.ErrorMessage == "" {
		t.Fatalf("export failure must set the persistent error rail, got %q", updated.Feedback.ErrorMessage)
	}
	if !strings.Contains(updated.Feedback.ErrorMessage, want.Error()) {
		t.Fatalf("export failure omitted cause: %q", updated.Feedback.ErrorMessage)
	}
}

func TestHandleContainerCopyDoneSuccess(t *testing.T) {
	app := advancedApp()
	trace := beginTrace(app, "resource.container.copy", "short")
	updated, cmd := handleContainerCopyDone(app, keyboard.ContainerCopyDone{
		ContainerID: "short",
		SourcePath:  "/etc/app.conf",
		Destination: "/tmp/app.tar",
		Bytes:       512,
		Audit:       trace,
	})
	if cmd != nil {
		t.Fatalf("success returned a cmd %T, want nil", cmd)
	}
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultSucceeded || records[2].TraceID != trace.ID {
		t.Fatalf("copy records=%#v", records)
	}
	if updated.Feedback.ToastMessage == "" {
		t.Fatalf("success copy must surface a toast, got %q", updated.Feedback.ToastMessage)
	}
}

func TestHandleContainerCopyDoneFailure(t *testing.T) {
	app := advancedApp()
	trace := beginTrace(app, "resource.container.copy", "short")
	want := errors.New("copy refused")
	updated, cmd := handleContainerCopyDone(app, keyboard.ContainerCopyDone{
		ContainerID: "short",
		Error:       want,
		Audit:       trace,
	})
	if cmd != nil {
		t.Fatalf("failure returned a cmd %T, want nil", cmd)
	}
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultFailed || records[2].TraceID != trace.ID {
		t.Fatalf("copy failure records=%#v", records)
	}
	if updated.Feedback.ErrorMessage == "" {
		t.Fatalf("copy failure must set the persistent error rail, got %q", updated.Feedback.ErrorMessage)
	}
	if !strings.Contains(updated.Feedback.ErrorMessage, want.Error()) {
		t.Fatalf("copy failure omitted cause: %q", updated.Feedback.ErrorMessage)
	}
}

func TestHandleContainerUpdateDoneSuccess(t *testing.T) {
	app := advancedApp()
	trace := beginTrace(app, "resource.container.update", "short")
	updated, cmd := handleContainerUpdateDone(app, keyboard.ContainerUpdateDone{
		ContainerID: "short",
		Audit:       trace,
	})
	if cmd != nil {
		t.Fatalf("success returned a cmd %T, want nil", cmd)
	}
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultSucceeded || records[2].TraceID != trace.ID {
		t.Fatalf("update records=%#v", records)
	}
	if updated.Feedback.ToastMessage == "" {
		t.Fatal("success update must surface a toast")
	}
}

func TestHandleContainerUpdateDoneFailure(t *testing.T) {
	app := advancedApp()
	trace := beginTrace(app, "resource.container.update", "short")
	want := errors.New("cgroup permission denied")
	updated, cmd := handleContainerUpdateDone(app, keyboard.ContainerUpdateDone{
		ContainerID: "short",
		Error:       want,
		Audit:       trace,
	})
	if cmd != nil {
		t.Fatalf("failure returned a cmd %T, want nil", cmd)
	}
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultFailed || records[2].TraceID != trace.ID {
		t.Fatalf("update failure records=%#v", records)
	}
	if updated.Feedback.ErrorMessage == "" {
		t.Fatalf("update failure must set the persistent error rail, got %q", updated.Feedback.ErrorMessage)
	}
	if !strings.Contains(updated.Feedback.ErrorMessage, want.Error()) {
		t.Fatalf("update failure omitted cause: %q", updated.Feedback.ErrorMessage)
	}
}

func TestHandleContainerCommitDoneSuccess(t *testing.T) {
	app := advancedApp()
	app.Connection.Engine = mockengine.New()
	trace := beginTrace(app, "resource.container.commit", "short")
	updated, cmd := handleContainerCommitDone(app, keyboard.ContainerCommitDone{
		ContainerID: "short",
		ImageID:     "sha256:abcdef",
		Audit:       trace,
	})
	if cmd == nil {
		t.Fatal("successful commit must trigger an image-list refresh")
	}
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultSucceeded || records[2].TraceID != trace.ID {
		t.Fatalf("commit records=%#v", records)
	}
	if updated.Feedback.ToastMessage == "" {
		t.Fatal("success commit must surface a toast")
	}
}

func TestHandleContainerCommitDoneStartsRequestedImageExport(t *testing.T) {
	app := advancedApp()
	app.Connection.Engine = mockengine.New()
	updated, cmd := handleContainerCommitDone(app, keyboard.ContainerCommitDone{
		ContainerID: "short", ImageID: "sha256:abcdef", ExportPath: "/tmp/snapshot.tar",
	})
	if cmd == nil || updated.Navigation.Mode != state.ModeImageTransfer {
		t.Fatalf("commit export did not enter image transfer: mode=%v cmd=%v", updated.Navigation.Mode, cmd)
	}
	request := updated.ImageTransfer.Request
	if request.Operation != runtimeapi.ImageTransferSave || request.Source != "sha256:abcdef" || request.Path != "/tmp/snapshot.tar" {
		t.Fatalf("image export request = %#v", request)
	}
}

func TestHandleContainerCommitDoneSuccessNoEngine(t *testing.T) {
	app := advancedApp()
	trace := beginTrace(app, "resource.container.commit", "short")
	updated, cmd := handleContainerCommitDone(app, keyboard.ContainerCommitDone{
		ContainerID: "short",
		ImageID:     "sha256:abcdef",
		Audit:       trace,
	})
	if cmd != nil {
		t.Fatalf("commit without an engine returned a cmd %T, want nil", cmd)
	}
	if updated.Feedback.ToastMessage == "" {
		t.Fatal("success commit must surface a toast even without an engine")
	}
}

func TestHandleContainerCommitDoneFailure(t *testing.T) {
	app := advancedApp()
	trace := beginTrace(app, "resource.container.commit", "short")
	want := errors.New("image name conflict")
	updated, cmd := handleContainerCommitDone(app, keyboard.ContainerCommitDone{
		ContainerID: "short",
		Error:       want,
		Audit:       trace,
	})
	if cmd != nil {
		t.Fatalf("failure returned a cmd %T, want nil", cmd)
	}
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultFailed || records[2].TraceID != trace.ID {
		t.Fatalf("commit failure records=%#v", records)
	}
	if updated.Feedback.ErrorMessage == "" {
		t.Fatalf("commit failure must set the persistent error rail, got %q", updated.Feedback.ErrorMessage)
	}
	if !strings.Contains(updated.Feedback.ErrorMessage, want.Error()) {
		t.Fatalf("commit failure omitted cause: %q", updated.Feedback.ErrorMessage)
	}
}
