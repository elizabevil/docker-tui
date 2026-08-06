package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func enterMarkMode(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	m.Navigation.Mode = state.ModeMark
	return doToggleMark(m)
}

func exitMarkMode(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	m.Navigation.Mode = state.ModeNormal
	return m, nil
}

func handleMarkMode(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	action, known := resolveAction(key, m)
	if known {
		switch action {
		case keys.ActionBack:
			return exitMarkMode(m)
		case keys.ActionContainerStart, keys.ActionContainerStop, keys.ActionContainerRestart, keys.ActionContainerKill, keys.ActionContainerPause:
			if m.Selection.MarkedCount(state.PanelContainers) > 0 {
				return handleAction(action, m, nil)
			}
			return m, nil
		case keys.ActionDelete, keys.ActionContainerRemove, keys.ActionImageRemove, keys.ActionVolumeRemove, keys.ActionNetworkRemove:
			return doBulkDelete(m)
		case keys.ActionEnter:
			return doToggleMark(m)
		case keys.ActionUp:
			moveCursor(m, -1)
			return m, nil
		case keys.ActionDown:
			moveCursor(m, 1)
			return m, nil
		case keys.ActionTabNext:
			switchPanel(m, 1)
			return m, nil
		case keys.ActionTabPrev:
			switchPanel(m, -1)
			return m, nil
		}
	}

	if keys.IsSpace(key) {
		return doToggleMark(m)
	}

	if key == keys.KeyG {
		moveCursor(m, -999)
	}
	return m, nil
}
