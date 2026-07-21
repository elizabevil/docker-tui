package keyboard

import (
	"unicode"

	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

var execShellOptions = []string{"/bin/sh", "/bin/bash", "/bin/ash"}

// handleExecDialogKeys routes keys while DialogState owns focus and input.
func handleExecDialogKeys(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch key {
	case keys.KeyTab:
		m.DialogFocus = (m.DialogFocus + 1) % state.ExecFocusCount
		return m, nil
	case keys.KeyShiftTab:
		m.DialogFocus = (m.DialogFocus - 1 + state.ExecFocusCount) % state.ExecFocusCount
		return m, nil
	case keys.KeyEnter:
		switch m.DialogFocus {
		case state.ExecFocusShell1, state.ExecFocusShell2, state.ExecFocusShell3:
			m.ExecShell = execShellOptions[m.DialogFocus]
			return doExecAction(m)
		case state.ExecFocusInput, state.ExecFocusConfirm:
			m.ExecShell = m.DialogState.Input.Text
			if m.ExecShell == "" {
				m.ExecShell = "/bin/sh"
			}
			return doExecAction(m)
		case state.ExecFocusCancel:
			clearDialogState(m)
		}
		return m, nil
	case keys.KeyEsc:
		clearDialogState(m)
		return m, nil
	default:
		if m.DialogFocus != state.ExecFocusInput || key == " " {
			return m, nil
		}
		if key == "ctrl+w" {
			m.DialogState.Input.DeleteDelimitedBackward(func(r rune) bool {
				return unicode.IsSpace(r) || r == '/'
			})
			return m, nil
		}
		if key == "ctrl+u" {
			m.DialogState.Input.Reset()
			return m, nil
		}
		editQueryInput(key, &m.DialogState.Input)
		return m, nil
	}
}
