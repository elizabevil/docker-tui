package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func enterMarkMode(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	m.Mode = state.ModeMark
	if m.MarkedIDs == nil {
		m.MarkedIDs = make(map[string]bool)
	}
	return m, nil
}

func exitMarkMode(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	m.Mode = state.ModeNormal
	return m, nil
}

func handleMarkMode(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	// Mark-specific keys
	switch key {
	case keys.KeyEsc:
		m.MarkedIDs = make(map[string]bool)
		return exitMarkMode(m)
	case keys.KeyCtrlD:
		return doBulkDelete(m)
	case keys.KeyEnter:
		return doToggleMark(m)
	}

	if keys.IsSpace(key) {
		return doToggleMark(m)
	}

	// Navigation passthrough
	if key == keys.KeyJ || key == keys.KeyUp {
		moveCursor(m, -1)
	} else if key == keys.KeyK || key == keys.KeyDown {
		moveCursor(m, 1)
	} else if key == keys.KeyG {
		moveCursor(m, -999)
	} else if key == keys.KeyG {
		moveCursor(m, 999)
	} else if key == keys.KeyTab {
		switchPanel(m, 1)
	} else if key == keys.KeyShiftTab {
		switchPanel(m, -1)
	}
	return m, nil
}
