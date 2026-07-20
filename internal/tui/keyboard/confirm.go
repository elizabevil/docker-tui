package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleConfirmKeys(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Mode != state.ModeConfirm {
		return nil, nil
	}
	switch key {
	case keys.KeyY, keys.KeyY_upper:
		return doConfirmYes(m)
	case keys.KeyN, "N", "esc":
		m.Mode = state.ModeNormal
		m.ConfirmAction = ""
		m.ConfirmTarget = ""
		m.ConfirmMessage = ""
		return m, nil
	}
	return m, nil
}
