package keyboard

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestHandleKeyPressUsesConfiguredBinding(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Dependencies.Config.Keymap.Help = []string{"z"}

	updated, _ := HandleKeyPress(keyMessage("z"), app)
	if updated.Navigation.Mode != state.ModeHelp {
		t.Fatalf("configured help binding did not open help: mode=%v", updated.Navigation.Mode)
	}
}

func TestHandleKeyPressUsesConfiguredNavigationBinding(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Dependencies.Config.Keymap.Down = []string{"z"}
	app.Resources.Containers.Items = []dockerclient.ContainerSummary{{ID: "one"}, {ID: "two"}}

	updated, _ := HandleKeyPress(keyMessage("z"), app)
	if updated.Resources.Containers.Cursor != 1 {
		t.Fatalf("configured down binding cursor=%d", updated.Resources.Containers.Cursor)
	}
	updated.Resources.Containers.Cursor = 0
	updated, _ = HandleKeyPress(keyMessage("j"), updated)
	if updated.Resources.Containers.Cursor != 0 {
		t.Fatalf("overridden default down binding remained active")
	}
}

func TestHandleKeyPressRemovesOverriddenDefaultBinding(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Dependencies.Config.Keymap.Help = []string{"z"}

	updated, _ := HandleKeyPress(keyMessage("?"), app)
	if updated.Navigation.Mode == state.ModeHelp {
		t.Fatal("overridden default help binding remained active")
	}
}

func TestConfiguredBindingsApplyInDetailMode(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Dependencies.Config.Keymap.Down = []string{"z"}
	app.Dependencies.Config.Keymap.Back = []string{"x"}
	app.Navigation.Mode = state.ModeDetail

	updated, _ := HandleKeyPress(keyMessage("z"), app)
	if updated.Detail.DetailOffset != 1 {
		t.Fatalf("configured detail down offset=%d", updated.Detail.DetailOffset)
	}
	updated, _ = HandleKeyPress(keyMessage("x"), updated)
	if updated.Navigation.Mode != state.ModeNormal {
		t.Fatalf("configured detail back mode=%v", updated.Navigation.Mode)
	}
}

func TestConfiguredBindingsApplyInLogMode(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Dependencies.Config.Keymap.Down = []string{"z"}
	app.Dependencies.Config.Keymap.Back = []string{"x"}
	app.Navigation.Mode = state.ModeLogView

	updated, _ := HandleKeyPress(keyMessage("z"), app)
	if updated.Log.LogViewOffset != 1 {
		t.Fatalf("configured log down offset=%d", updated.Log.LogViewOffset)
	}
	updated, _ = HandleKeyPress(keyMessage("x"), updated)
	if updated.Navigation.Mode != state.ModeNormal {
		t.Fatalf("configured log back mode=%v", updated.Navigation.Mode)
	}
}

func TestConfiguredBackBindingExitsMarkMode(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Dependencies.Config.Keymap.Back = []string{"x"}
	app.Navigation.Mode = state.ModeMark
	app.Selection.MarkedIDs = make(map[string]bool)
	app.Selection.MarkedIDs["one"] = true

	updated, _ := HandleKeyPress(keyMessage("x"), app)
	if updated.Navigation.Mode != state.ModeNormal || len(updated.Selection.MarkedIDs) != 0 {
		t.Fatalf("configured mark back mode=%v marked=%v", updated.Navigation.Mode, updated.Selection.MarkedIDs)
	}
}

func TestConfiguredHelpBindingClosesHelp(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Dependencies.Config.Keymap.Help = []string{"z"}
	app.Navigation.Mode = state.ModeHelp

	updated, _ := HandleKeyPress(keyMessage("?"), app)
	if updated.Navigation.Mode != state.ModeHelp {
		t.Fatal("overridden default help binding closed help")
	}
	updated, _ = HandleKeyPress(keyMessage("z"), updated)
	if updated.Navigation.Mode != state.ModeNormal {
		t.Fatalf("configured help binding did not close help: mode=%v", updated.Navigation.Mode)
	}
}

func TestCommandInputPreservesPrintableKeyCase(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Navigation.Mode = state.ModeCommand

	updated, _ := HandleKeyPress(keyMessage("A"), app)
	if updated.Navigation.CommandInput.Text != "A" {
		t.Fatalf("command text = %q, want uppercase input", updated.Navigation.CommandInput.Text)
	}
}

func keyMessage(key string) tea.KeyPressMsg {
	runes := []rune(key)
	return tea.KeyPressMsg(tea.Key{Text: key, Code: runes[0]})
}
