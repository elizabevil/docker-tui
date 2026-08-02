package keyboard

import (
	"os"
	"path/filepath"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/docker"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestImageSaveInputStartsCancellableWorkflow(t *testing.T) {
	model := &state.AppModel{
		Connection: state.ConnectionState{Engine: &docker.Client{}},
		Navigation: state.NavigationState{ActivePanel: state.PanelImages},
		Resources:  state.ResourceState{Images: state.NewImageListModel()},
	}
	model.Resources.Images.Items = []runtimeapi.ImageSummary{{ID: "sha256:image", RepoTags: []string{"example/app:v1"}}}
	if _, command := openImageWorkflow(model, runtimeapi.ImageTransferSave); command != nil {
		t.Fatal("opening a workflow unexpectedly started a command")
	}
	if model.Navigation.Mode != state.ModeContainerForm || model.Form.Kind != state.FormImageSave {
		t.Fatalf("save form = %#v, mode=%v", model.Form, model.Navigation.Mode)
	}
	dir := t.TempDir()
	model.Form.Get(fieldImagePath).Input.Set("archive.tar")
	model.Form.CWD = dir
	model.Form.OnConfirm = true
	updated, command := handleContainerFormKey(keys.KeyEnter, model)
	if command == nil || updated.Navigation.Mode != state.ModeImageTransfer {
		t.Fatalf("workflow did not start: mode=%v command=%v", updated.Navigation.Mode, command)
	}
	if updated.ImageTransfer.Request.Source != "example/app:v1" || updated.ImageTransfer.Request.Path != filepath.Join(dir, "archive.tar") {
		t.Fatalf("transfer request = %#v", updated.ImageTransfer.Request)
	}
}

func TestImageLoadWorkflowDoesNotRequireSelection(t *testing.T) {
	model := &state.AppModel{
		Connection: state.ConnectionState{Engine: &docker.Client{}},
		Navigation: state.NavigationState{ActivePanel: state.PanelImages},
		Resources:  state.ResourceState{Images: state.NewImageListModel()},
	}
	updated, _ := openImageWorkflow(model, runtimeapi.ImageTransferLoad)
	if updated.Navigation.Mode != state.ModeContainerForm || updated.Form.Kind != state.FormImageLoad {
		t.Fatalf("load form = %#v, mode=%v", updated.Form, updated.Navigation.Mode)
	}
}

func TestImageLoadFormStartsTransferWithExistingFile(t *testing.T) {
	model := &state.AppModel{
		Connection: state.ConnectionState{Engine: &docker.Client{}},
		Navigation: state.NavigationState{ActivePanel: state.PanelImages},
		Resources:  state.ResourceState{Images: state.NewImageListModel()},
	}
	dir := t.TempDir()
	archive := filepath.Join(dir, "image.tar")
	if err := os.WriteFile(archive, []byte("tar"), 0o644); err != nil {
		t.Fatal(err)
	}
	openImageWorkflow(model, runtimeapi.ImageTransferLoad)
	model.Form.Get(fieldImagePath).Input.Set("image.tar")
	model.Form.CWD = dir
	model.Form.OnConfirm = true
	updated, command := submitContainerForm(model)
	if command == nil || updated.Navigation.Mode != state.ModeImageTransfer {
		t.Fatalf("load did not start: mode=%v command=%v", updated.Navigation.Mode, command)
	}
	if updated.ImageTransfer.Request.Operation != runtimeapi.ImageTransferLoad || updated.ImageTransfer.Request.Path != archive {
		t.Fatalf("transfer request = %#v", updated.ImageTransfer.Request)
	}
}

func TestImageSaveOverwriteRequiresForce(t *testing.T) {
	model := &state.AppModel{
		Connection: state.ConnectionState{Engine: &docker.Client{}},
		Navigation: state.NavigationState{ActivePanel: state.PanelImages},
		Resources:  state.ResourceState{Images: state.NewImageListModel()},
	}
	model.Resources.Images.Items = []runtimeapi.ImageSummary{{ID: "sha256:image", RepoTags: []string{"example/app:v1"}}}
	openImageWorkflow(model, runtimeapi.ImageTransferSave)
	target := filepath.Join(t.TempDir(), "existing.tar")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	model.Form.Get(fieldImagePath).Input.Set(target)
	model.Form.OnConfirm = true
	updated, command := submitContainerForm(model)
	if command != nil || updated.Navigation.Mode != state.ModeConfirm || updated.Confirm.Focus != 0 {
		t.Fatalf("overwrite confirmation = %#v, command=%v", updated.Confirm, command)
	}
	updated, _ = handleConfirmKeys(keys.KeyTab, updated)
	updated, command = handleConfirmKeys(keys.KeyEnter, updated)
	if command == nil || updated.Navigation.Mode != state.ModeImageTransfer {
		t.Fatalf("force save did not start: mode=%v command=%v", updated.Navigation.Mode, command)
	}
	if updated.ImageTransfer.Request.Path != target || updated.ImageTransfer.Request.Source != "example/app:v1" {
		t.Fatalf("transfer request = %#v", updated.ImageTransfer.Request)
	}
}
