package keyboard

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime/docker"
	"github.com/elizabevil/docker-tui/internal/tui/actionbar"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestQIsUnassignedAndCtrlCStillQuits(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	updated, cmd := HandleKeyPress(keyMessage("q"), m)
	if cmd != nil || updated.Feedback.ToastMessage != "" || updated.Navigation.Mode != state.ModeNormal {
		t.Fatalf("q should be a silent no-op: toast=%q mode=%v cmd=%v", updated.Feedback.ToastMessage, updated.Navigation.Mode, cmd)
	}
	updated, cmd = HandleKeyPress(keyMessage("ctrl+c"), updated)
	if cmd == nil {
		t.Fatal("ctrl+c no longer returns quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("ctrl+c command returned %T, want tea.QuitMsg", cmd())
	}
}

func TestActionBarOpenFilterAndClose(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	updated, _ := HandleKeyPress(keyMessage(";"), m)
	if updated.Navigation.Mode != state.ModeActionBar || !updated.Navigation.ActionBar.Visible {
		t.Fatalf("action bar did not open: mode=%v state=%#v", updated.Navigation.Mode, updated.Navigation.ActionBar)
	}
	updated, _ = HandleKeyPress(keyMessage("/"), updated)
	updated, _ = HandleKeyPress(keyMessage("镜"), updated)
	updated, _ = HandleKeyPress(keyMessage("backspace"), updated)
	if updated.Navigation.ActionBar.Filter != "" {
		t.Fatalf("unicode backspace left filter %q", updated.Navigation.ActionBar.Filter)
	}
	updated, _ = HandleKeyPress(keyMessage("esc"), updated)
	if !updated.Navigation.ActionBar.Visible || updated.Navigation.ActionBar.Filtering {
		t.Fatalf("first Esc should leave filter mode: %#v", updated.Navigation.ActionBar)
	}
	updated, _ = HandleKeyPress(keyMessage("esc"), updated)
	if updated.Navigation.Mode != state.ModeNormal || updated.Navigation.ActionBar.Visible {
		t.Fatalf("second Esc should close Action Bar: mode=%v state=%#v", updated.Navigation.Mode, updated.Navigation.ActionBar)
	}
}

func TestActionBarDoesNotExecuteDisabledItems(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelContainers
	doActionBar(m)
	updated, cmd := handleActionBarKeys("enter", m)
	if cmd != nil || updated.Navigation.Mode != state.ModeActionBar {
		t.Fatalf("disabled action executed: mode=%v cmd=%v", updated.Navigation.Mode, cmd)
	}
}

func TestActionBarRenameOpensExistingDialog(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), &dockerclient.Client{}, "test")
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{ID: "one", Name: "api", State: state.ContainerStateRunning}}
	doActionBar(m)
	items := actionbar.VisibleItems(m)
	if len(items) == 0 {
		t.Fatal("rename action missing")
	}
	updated, cmd := handleActionBarKeys("enter", m)
	if cmd != nil || updated.Navigation.Mode != state.ModeRename || updated.Dialog.Input.Text != "api" {
		t.Fatalf("rename workflow not opened: mode=%v dialog=%#v cmd=%v", updated.Navigation.Mode, updated.Dialog, cmd)
	}
}
