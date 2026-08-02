package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleConfirmKeys(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Navigation.Mode != state.ModeConfirm {
		return nil, nil
	}
	switch key {
	case keys.KeyTab:
		m.Confirm.MoveFocus(1)
	case keys.KeyShiftTab:
		m.Confirm.MoveFocus(-1)
	case keys.KeyEnter:
		if m.Confirm.Focus >= 0 && m.Confirm.Focus < len(m.Confirm.Options) {
			switch m.Confirm.Options[m.Confirm.Focus].ID {
			case "cancel":
				return cancelConfirm(m)
			case "confirm":
				return doConfirmYes(m)
			case "force":
				if m.Confirm.ConfirmAction == "bulk-delete" {
					m.Confirm.ConfirmAction = "bulk-delete-force"
				} else if m.Confirm.ConfirmAction == "batch-stop" {
					m.Confirm.ConfirmAction = "batch-kill"
				}
				return doConfirmYes(m)
			}
		}
		if m.Confirm.Focus == 0 || m.Confirm.Focus < 0 || m.Confirm.Focus >= len(m.Confirm.Options) {
			return doConfirmYes(m)
		}
		return cancelConfirm(m)
	case keys.KeyY, keys.KeyY_upper:
		return doConfirmYes(m)
	case keys.KeyN, keys.KeyNUpper, keys.KeyEsc:
		FinishAudit(m, m.Confirm.ConfirmAudit, audit.ResultCancelled, "Operation cancelled", audit.Details{})
		m.Navigation.Mode = confirmReturnMode(m)
		m.Confirm.Close()
		return m, nil
	}
	return m, nil
}

func cancelConfirm(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	FinishAudit(m, m.Confirm.ConfirmAudit, audit.ResultCancelled, "Operation cancelled", audit.Details{})
	m.Navigation.Mode = confirmReturnMode(m)
	m.Confirm.Close()
	return m, nil
}

func confirmReturnMode(m *state.AppModel) state.AppMode {
	if m.Confirm.ReturnMode != state.ModeNormal {
		return m.Confirm.ReturnMode
	}
	return state.ModeNormal
}
