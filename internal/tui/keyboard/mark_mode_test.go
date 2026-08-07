package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestEscExitsMarkModeAndClearsMarks(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.Mode = state.ModeMark
	m.Selection.Toggle(state.PanelContainers, "one")
	m.Selection.Toggle(state.PanelImages, "img-1")

	updated, cmd := HandleKeyPress(keyMessage(keys.KeyEsc), m)

	if cmd != nil {
		t.Fatalf("Esc from mark mode returned unexpected cmd: %v", cmd)
	}
	if updated.Navigation.Mode != state.ModeNormal {
		t.Fatalf("mode = %v, want %v", updated.Navigation.Mode, state.ModeNormal)
	}
	if updated.Selection.MarkedCount(state.PanelContainers) != 0 {
		t.Fatalf("containers marks = %d, want 0", updated.Selection.MarkedCount(state.PanelContainers))
	}
	if updated.Selection.MarkedCount(state.PanelImages) != 0 {
		t.Fatalf("images marks = %d, want 0", updated.Selection.MarkedCount(state.PanelImages))
	}
}

func TestEscFromMarkModeClearsUnmarkedPanelToo(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.Mode = state.ModeMark
	m.Selection.Toggle(state.PanelVolumes, "vol-1")

	updated, _ := HandleKeyPress(keyMessage(keys.KeyEsc), m)

	if updated.Selection.MarkedCount(state.PanelVolumes) != 0 {
		t.Fatalf("volumes marks = %d, want 0", updated.Selection.MarkedCount(state.PanelVolumes))
	}
}
