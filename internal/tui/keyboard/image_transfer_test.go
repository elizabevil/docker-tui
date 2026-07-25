package keyboard

import (
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
	model.Dialog.Input.Set("archive.tar")
	updated, command := handleImageWorkflowInput(keys.KeyEnter, model)
	if command == nil || updated.Navigation.Mode != state.ModeImageTransfer {
		t.Fatalf("workflow did not start: mode=%v command=%v", updated.Navigation.Mode, command)
	}
	if updated.ImageTransfer.Request.Source != "example/app:v1" || updated.ImageTransfer.Request.Path != "archive.tar" {
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
	if updated.Navigation.Mode != state.ModeImageWorkflow || updated.Dialog.Kind != state.DialogImageLoad {
		t.Fatalf("load dialog = %#v", updated.Dialog)
	}
}
