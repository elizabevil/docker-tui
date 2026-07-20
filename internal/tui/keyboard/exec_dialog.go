package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

const (
	execFocusShell1  = iota // 0
	execFocusShell2         // 1
	execFocusShell3         // 2
	execFocusInput          // 3
	execFocusConfirm        // 4
	execFocusCancel         // 5
	execFocusCount          // 6
)

var execShellOptions = []string{"/bin/sh", "/bin/bash", "/bin/ash"}

// handleExecDialogKeys handles keyboard input for the exec dialog (ModeExec).
// Focus positions: 0-2 shell options, 3 custom input, 4 confirm, 5 cancel.
// When input (3) is focused, supports full line editing with cursor.
func handleExecDialogKeys(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch key {
	// ── Focus navigation ──
	case keys.KeyTab:
		m.DialogFocus = (m.DialogFocus + 1) % execFocusCount
		return m, nil

	case keys.KeyShiftTab:
		m.DialogFocus = (m.DialogFocus - 1 + execFocusCount) % execFocusCount
		return m, nil

	// ── Execute / confirm ──
	case keys.KeyEnter:
		switch m.DialogFocus {
		case execFocusShell1, execFocusShell2, execFocusShell3:
			m.ExecShell = execShellOptions[m.DialogFocus]
			return doExecAction(m)
		case execFocusInput:
			m.ExecShell = m.FilterText
			if m.ExecShell == "" {
				m.ExecShell = "/bin/sh"
			}
			return doExecAction(m)
		case execFocusConfirm:
			m.ExecShell = m.FilterText
			if m.ExecShell == "" {
				m.ExecShell = "/bin/sh"
			}
			return doExecAction(m)
		case execFocusCancel:
			clearDialogState(m)
		}
		return m, nil

	case keys.KeyEsc:
		clearDialogState(m)
		return m, nil

	// ── Input editing (only when input field is focused) ──
	case keys.KeyLeft:
		if m.DialogFocus == execFocusInput && m.DialogCursor > 0 {
			m.DialogCursor--
		}
		return m, nil

	case keys.KeyRight:
		if m.DialogFocus == execFocusInput && m.DialogCursor < len(m.FilterText) {
			m.DialogCursor++
		}
		return m, nil

	case keys.KeyHome:
		if m.DialogFocus == execFocusInput {
			m.DialogCursor = 0
		}
		return m, nil

	case keys.KeyEnd:
		if m.DialogFocus == execFocusInput {
			m.DialogCursor = len(m.FilterText)
		}
		return m, nil

	case keys.KeyBackspace, keys.KeyDelete:
		if m.DialogFocus == execFocusInput && len(m.FilterText) > 0 {
			if key == keys.KeyDelete || key == "delete" {
				// Delete at cursor: remove character AFTER cursor
				if m.DialogCursor < len(m.FilterText) {
					m.FilterText = m.FilterText[:m.DialogCursor] + m.FilterText[m.DialogCursor+1:]
				}
			} else {
				// Backspace: remove character BEFORE cursor
				if m.DialogCursor > 0 {
					m.FilterText = m.FilterText[:m.DialogCursor-1] + m.FilterText[m.DialogCursor:]
					m.DialogCursor--
				}
			}
		}
		return m, nil

	// ── Control editing shortcuts ──
	case "ctrl+a":
		if m.DialogFocus == execFocusInput {
			m.DialogCursor = 0
		}
		return m, nil

	case "ctrl+e":
		if m.DialogFocus == execFocusInput {
			m.DialogCursor = len(m.FilterText)
		}
		return m, nil

	case "ctrl+u":
		if m.DialogFocus == execFocusInput {
			m.FilterText = ""
			m.DialogCursor = 0
		}
		return m, nil

	case "ctrl+w":
		if m.DialogFocus == execFocusInput && m.DialogCursor > 0 {
			// Delete word backward: find start of word before cursor
			pos := m.DialogCursor - 1
			for pos >= 0 && m.FilterText[pos] == '/' {
				pos--
			}
			for pos >= 0 && m.FilterText[pos] != '/' && m.FilterText[pos] != ' ' {
				pos--
			}
			pos++
			m.FilterText = m.FilterText[:pos] + m.FilterText[m.DialogCursor:]
			m.DialogCursor = pos
		}
		return m, nil

	// ── Printable character: insert at cursor position ──
	default:
		if m.DialogFocus == execFocusInput && len(key) == 1 && key != " " {
			// Insert character at cursor position
			before := m.FilterText[:m.DialogCursor]
			after := m.FilterText[m.DialogCursor:]
			m.FilterText = before + key + after
			m.DialogCursor++
		}
		return m, nil
	}
}
