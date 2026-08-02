package keyboard

import (
	"unicode/utf8"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func openEventsPage(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	previous := m.Navigation.Mode
	if previous == state.ModeEvents {
		previous = state.ModeNormal
	}
	m.EventPanel.Open(previous)
	m.Navigation.Mode = state.ModeEvents
	return m, nil
}

func handleEventPanelKeys(rawKey string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	key := keys.Normalize(rawKey)
	panel := &m.EventPanel
	if panel.Filtering {
		switch key {
		case keys.KeyEsc, keys.KeyEnter:
			panel.Filtering = false
		case keys.KeyBackspace:
			filter := []rune(panel.Filter)
			if len(filter) > 0 {
				panel.SetFilter(string(filter[:len(filter)-1]))
				panel.Filtering = true
			}
		default:
			if utf8.RuneCountInString(rawKey) == 1 {
				panel.SetFilter(panel.Filter + rawKey)
				panel.Filtering = true
			}
		}
		return m, nil
	}

	items := panel.FilteredEvents()
	visible := max(1, m.Viewport.Height-18)
	switch key {
	case keys.KeyEsc:
		previous := panel.PreviousMode
		if previous == state.ModeEvents {
			previous = state.ModeNormal
		}
		m.Navigation.Mode = previous
		return m, nil
	case keys.KeyJ, keys.KeyDown:
		panel.MoveCursor(1, len(items), visible)
	case keys.KeyK, keys.KeyUp:
		panel.MoveCursor(-1, len(items), visible)
	case keys.KeyPgDn:
		panel.MoveCursor(visible, len(items), visible)
	case keys.KeyPgUp:
		panel.MoveCursor(-visible, len(items), visible)
	case keys.KeySpace:
		panel.TogglePause()
	case keys.KeySlash:
		panel.Filtering = true
	case keys.KeyCtrlD:
		m.Confirm.Open("events-clear", "events", "Clear all retained runtime events?", audit.Trace{})
		m.Confirm.ReturnMode = state.ModeEvents
		m.Navigation.Mode = state.ModeConfirm
	}
	return m, nil
}
