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
		m.Dialog.MoveFocus(1, state.ExecFocusCount)
		return m, nil
	case keys.KeyShiftTab:
		m.Dialog.MoveFocus(-1, state.ExecFocusCount)
		return m, nil
	case keys.KeyEnter:
		switch m.Dialog.Focus {
		case state.ExecFocusShell1, state.ExecFocusShell2, state.ExecFocusShell3:
			m.Exec.SetShell(execShellOptions[m.Dialog.Focus])
			return doExecAction(m)
		case state.ExecFocusInput, state.ExecFocusConfirm:
			m.Exec.SetShell(m.Dialog.Input.Text)
			return doExecAction(m)
		case state.ExecFocusCancel:
			clearDialogState(m)
		}
		return m, nil
	case keys.KeyEsc:
		clearDialogState(m)
		return m, nil
	default:
		if m.Dialog.Focus != state.ExecFocusInput || key == " " {
			return m, nil
		}
		if key == "ctrl+w" {
			m.Dialog.Input.DeleteDelimitedBackward(func(r rune) bool {
				return unicode.IsSpace(r) || r == '/'
			})
			return m, nil
		}
		if key == "ctrl+u" {
			m.Dialog.Input.Reset()
			return m, nil
		}
		editQueryInput(key, &m.Dialog.Input)
		return m, nil
	}
}
