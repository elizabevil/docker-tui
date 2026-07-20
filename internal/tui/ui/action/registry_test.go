package action

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestContextPrefersModeActions(t *testing.T) {
	app := &state.AppModel{Mode: state.ModeDetail, ActivePanel: state.PanelImages}
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
