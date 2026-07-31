package action

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestContextPrefersModeActions(t *testing.T) {
	app := &state.AppModel{Navigation: state.NavigationState{Mode: state.ModeDetail, ActivePanel: state.PanelImages}}
	got := Context(app)
	if len(got) == 0 || got[0].Key != "Esc/Enter" {
		t.Fatalf("Context() did not project detail mode actions: %#v", got)
	}
}

func TestSectionsUseRegistryActions(t *testing.T) {
	sections := Sections()
	if len(sections) < 5 {
		t.Fatalf("Sections() returned %d sections", len(sections))
	}
	if len(sections[0].Shortcuts) != len(Global()) {
		t.Fatalf("global help section and footer registry differ")
	}
}

func TestGlobalProjectsConfiguredBindings(t *testing.T) {
	app := &state.AppModel{Dependencies: state.Dependencies{Config: config.DefaultConfig()}}
	app.Dependencies.Config.Keymap.Help = []string{"f3"}
	shortcuts := Global(app)
	for _, shortcut := range shortcuts {
		if shortcut.Description == i18n.T("key.help") && shortcut.Key == "F3" {
			return
		}
	}
	t.Fatalf("configured help binding not found: %#v", shortcuts)
}

func TestContextUsesComposeFocus(t *testing.T) {
	app := &state.AppModel{Navigation: state.NavigationState{ActivePanel: state.PanelCompose}}

	app.Compose.ComposeFocus = 0
	left := Context(app)
	if len(left) == 0 || left[0].Key != "Enter/→" || left[0].Description != i18n.T("key.services") {
		t.Fatalf("compose project focus shortcuts = %#v", left)
	}

	app.Compose.ComposeFocus = 1
	right := Context(app)
	if len(right) == 0 || right[0].Key != "Enter" || right[0].Description != i18n.T("key.expand") {
		t.Fatalf("compose service focus shortcuts = %#v", right)
	}
}
