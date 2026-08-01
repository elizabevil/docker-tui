package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/runtime"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime/docker"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestPauseRejectsInapplicableContainer(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), &dockerclient.Client{}, "test")
	m.Resources.Containers.Items = []runtime.ContainerSummary{{ID: "one", Name: "done", State: state.ContainerStateExited}}
	_, cmd := doPauseAction(m)
	if cmd != nil || m.Feedback.ToastMessage == "" {
		t.Fatalf("inapplicable pause was not rejected: cmd=%v toast=%q", cmd != nil, m.Feedback.ToastMessage)
	}
}

func TestBatchPauseSkipsInapplicableContainers(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), &dockerclient.Client{}, "test")
	m.Resources.Containers.Items = []runtime.ContainerSummary{{ID: "one", State: state.ContainerStateExited}}
	m.Selection.Toggle("one")
	_, cmd := doPauseAction(m)
	if cmd == nil {
		t.Fatal("batch pause did not return a result command")
	}
	msg := cmd().(state.ContainerBatchActioned)
	if msg.Success != 0 || msg.Skipped != 1 || msg.Failed != 0 {
		t.Fatalf("batch result = %#v", msg)
	}
	if len(m.Selection.MarkedIDs) != 0 {
		t.Fatalf("marks were not cleared: %#v", m.Selection.MarkedIDs)
	}
}

func TestBatchContainerActionInitializesChoiceOptions(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), &dockerclient.Client{}, "test")
	m.Selection.Toggle("one")

	updated, cmd := doBatchContainerAction(m, "stop", nil)
	if cmd != nil || updated.Navigation.Mode != state.ModeConfirm {
		t.Fatalf("batch confirmation = mode %v cmd=%v", updated.Navigation.Mode, cmd != nil)
	}
	if len(updated.Confirm.Options) != 2 || updated.Confirm.Options[0].ID != "cancel" || updated.Confirm.Options[1].ID != "force" {
		t.Fatalf("batch choice options = %#v", updated.Confirm.Options)
	}
}

func TestBulkDeleteOptionsExposeForceWhenSupported(t *testing.T) {
	options := bulkDeleteOptions(state.PanelContainers)
	if len(options) != 2 || options[0].ID != "cancel" || options[1].ID != "force" {
		t.Fatalf("container bulk delete options = %#v", options)
	}

	networkOptions := bulkDeleteOptions(state.PanelNetworks)
	if len(networkOptions) != 2 || networkOptions[0].ID != "cancel" || networkOptions[1].ID != "confirm" {
		t.Fatalf("network bulk delete options = %#v", networkOptions)
	}
}

func TestBatchStopOffersCancelDefaultAndForce(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), &dockerclient.Client{}, "test")
	m.Selection.Toggle("one")
	updated, cmd := doBatchContainerAction(m, "stop", nil)
	if cmd != nil || updated.Confirm.Focus != 0 || len(updated.Confirm.Options) != 2 {
		t.Fatalf("batch stop confirmation = %#v", updated.Confirm)
	}
	if updated.Confirm.Options[0].ID != "cancel" || updated.Confirm.Options[1].ID != "force" {
		t.Fatalf("batch stop options = %#v", updated.Confirm.Options)
	}
}

func TestTopRejectsStoppedContainer(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), &dockerclient.Client{}, "test")
	m.Resources.Containers.Items = []runtime.ContainerSummary{{ID: "one", Name: "done", State: state.ContainerStateExited}}
	_, cmd := openTopView(m)
	if cmd != nil || m.Navigation.Mode == state.ModeTop || m.Feedback.ToastMessage == "" {
		t.Fatalf("stopped container entered top: mode=%v toast=%q", m.Navigation.Mode, m.Feedback.ToastMessage)
	}
}

func TestRenameDialogValidatesName(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), &dockerclient.Client{}, "test")
	m.Resources.Containers.Items = []runtime.ContainerSummary{{ID: "one", Name: "api", State: state.ContainerStateRunning}}
	openRenameDialog(m)
	if m.Navigation.Mode != state.ModeRename || m.Dialog.Input.Text != "api" {
		t.Fatalf("rename dialog = %#v", m.Dialog)
	}
	m.Dialog.Input.Set("bad name")
	_, cmd := handleRenameDialogKey("enter", m)
	if cmd != nil || m.Navigation.Mode != state.ModeRename || m.Feedback.ToastMessage == "" {
		t.Fatalf("invalid rename was accepted: mode=%v", m.Navigation.Mode)
	}
}

func TestContainerCommandsIgnoreOtherPanels(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), &dockerclient.Client{}, "test")
	m.Resources.Containers.Items = []runtime.ContainerSummary{{ID: "one", Name: "api", State: state.ContainerStateRunning}}
	m.Navigation.ActivePanel = state.PanelImages
	if _, cmd := openRenameDialog(m); cmd != nil || m.Navigation.Mode != state.ModeNormal {
		t.Fatal("rename opened outside containers panel")
	}
	if _, cmd := openTopView(m); cmd != nil || m.Navigation.Mode != state.ModeNormal {
		t.Fatal("top opened outside containers panel")
	}
	if _, cmd := openPortDetail(m); cmd != nil || m.Navigation.Mode != state.ModeNormal {
		t.Fatal("port detail opened outside containers panel")
	}
}
