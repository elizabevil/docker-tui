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
	app.Config.Keymap.Help = []string{"z"}

	updated, _ := HandleKeyPress(keyMessage("z"), app)
	if updated.Mode != state.ModeHelp {
		t.Fatalf("configured help binding did not open help: mode=%v", updated.Mode)
	}
}

func TestHandleKeyPressUsesConfiguredNavigationBinding(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Config.Keymap.Down = []string{"z"}
	app.Containers.Items = []dockerclient.ContainerSummary{{ID: "one"}, {ID: "two"}}

	updated, _ := HandleKeyPress(keyMessage("z"), app)
	if updated.Containers.Cursor != 1 {
		t.Fatalf("configured down binding cursor=%d", updated.Containers.Cursor)
	}
	updated.Containers.Cursor = 0
	updated, _ = HandleKeyPress(keyMessage("j"), updated)
	if updated.Containers.Cursor != 0 {
		t.Fatalf("overridden default down binding remained active")
	}
}

func TestHandleKeyPressRemovesOverriddenDefaultBinding(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Config.Keymap.Help = []string{"z"}

	updated, _ := HandleKeyPress(keyMessage("?"), app)
	if updated.Mode == state.ModeHelp {
		t.Fatal("overridden default help binding remained active")
	}
}

func TestConfiguredBindingsApplyInDetailMode(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Config.Keymap.Down = []string{"z"}
	app.Config.Keymap.Back = []string{"x"}
	app.Mode = state.ModeDetail

	updated, _ := HandleKeyPress(keyMessage("z"), app)
	if updated.DetailOffset != 1 {
		t.Fatalf("configured detail down offset=%d", updated.DetailOffset)
	}
	updated, _ = HandleKeyPress(keyMessage("x"), updated)
	if updated.Mode != state.ModeNormal {
		t.Fatalf("configured detail back mode=%v", updated.Mode)
	}
}

func TestConfiguredBindingsApplyInLogMode(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Config.Keymap.Down = []string{"z"}
	app.Config.Keymap.Back = []string{"x"}
	app.Mode = state.ModeLogView

	updated, _ := HandleKeyPress(keyMessage("z"), app)
	if updated.LogViewOffset != 1 {
		t.Fatalf("configured log down offset=%d", updated.LogViewOffset)
	}
	updated, _ = HandleKeyPress(keyMessage("x"), updated)
	if updated.Mode != state.ModeNormal {
		t.Fatalf("configured log back mode=%v", updated.Mode)
	}
}

func TestConfiguredBackBindingExitsMarkMode(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Config.Keymap.Back = []string{"x"}
	app.Mode = state.ModeMark
	app.MarkedIDs = make(map[string]bool)
	app.MarkedIDs["one"] = true

	updated, _ := HandleKeyPress(keyMessage("x"), app)
	if updated.Mode != state.ModeNormal || len(updated.MarkedIDs) != 0 {
		t.Fatalf("configured mark back mode=%v marked=%v", updated.Mode, updated.MarkedIDs)
	}
}

func TestConfiguredHelpBindingClosesHelp(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Config.Keymap.Help = []string{"z"}
	app.Mode = state.ModeHelp

	updated, _ := HandleKeyPress(keyMessage("?"), app)
	if updated.Mode != state.ModeHelp {
		t.Fatal("overridden default help binding closed help")
	}
	updated, _ = HandleKeyPress(keyMessage("z"), updated)
	if updated.Mode != state.ModeNormal {
		t.Fatalf("configured help binding did not close help: mode=%v", updated.Mode)
	}
}

func TestCommandInputPreservesPrintableKeyCase(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Mode = state.ModeCommand

	updated, _ := HandleKeyPress(keyMessage("A"), app)
	if updated.FilterText != "A" {
		t.Fatalf("command text = %q, want uppercase input", updated.FilterText)
	}
}

func keyMessage(key string) tea.KeyPressMsg {
	runes := []rune(key)
	return tea.KeyPressMsg(tea.Key{Text: key, Code: runes[0]})
}
