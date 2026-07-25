package keyboard

import "github.com/elizabevil/docker-tui/internal/tui/keys"

// mapKeyToTerm converts bubbletea key strings to terminal byte sequences
// for Docker exec TTY passthrough. Supports Linux shell basic operations:
//   - Single printable characters (a-z, A-Z, 0-9, punctuation)
//   - Control keys: Enter, Backspace, Tab, Esc, Space
//   - Arrow keys (CSI sequences)
//   - All Ctrl+letter combinations (0x01-0x1A)
//   - Common navigation: Home, End, PgUp, PgDn, Delete
func mapKeyToTerm(key string) []byte {
	switch key {
	// ── Control characters ──
	case keys.KeyEnter:
		return []byte{'\r'} // 0x0D
	case keys.KeyBackspace, keys.KeyCtrlH:
		return []byte{0x08} // BS
	case keys.KeyTab, keys.KeyCtrlI:
		return []byte{'\t'} // 0x09
	case keys.KeyEsc, keys.KeyCtrlOpenBracket:
		return []byte{'\x1b'} // Escape
	case keys.KeySpaceName, keys.KeySpace:
		return []byte{' '}

	// ── Ctrl+letter (0x01-0x1A) — shell essential ──
	case keys.KeyCtrlA:
		return []byte{0x01} // Home / beginning of line
	case keys.KeyCtrlB:
		return []byte{0x02} // Backward char
	case keys.KeyCtrlC:
		return []byte{0x03} // SIGINT / interrupt
	case keys.KeyCtrlD:
		return []byte{0x04} // EOF / exit
	case keys.KeyCtrlE:
		return []byte{0x05} // End of line
	case keys.KeyCtrlF:
		return []byte{0x06} // Forward char
	case keys.KeyCtrlG:
		return []byte{0x07} // Bell / abort
	case keys.KeyCtrlJ:
		return []byte{0x0A} // Line feed
	case keys.KeyCtrlK:
		return []byte{0x0B} // Kill to end of line
	case keys.KeyCtrlL:
		return []byte{0x0C} // Clear screen (form feed)
	case keys.KeyCtrlM:
		return []byte{'\r'} // 0x0D — same as Enter
	case keys.KeyCtrlN:
		return []byte{0x0E} // Next history
	case keys.KeyCtrlO:
		return []byte{0x0F} // Operate / newline
	case keys.KeyCtrlP:
		return []byte{0x10} // Previous history
	case keys.KeyCtrlQ:
		return []byte{0x11} // XON / resume flow
	case keys.KeyCtrlR:
		return []byte{0x12} // Reverse search
	case keys.KeyCtrlS:
		return []byte{0x13} // XOFF / stop flow
	case keys.KeyCtrlT:
		return []byte{0x14} // Transpose chars
	case keys.KeyCtrlU:
		return []byte{0x15} // Kill to beginning of line
	case keys.KeyCtrlV:
		return []byte{0x16} // Literal next / paste
	case keys.KeyCtrlW:
		return []byte{0x17} // Kill word backward
	case keys.KeyCtrlX:
		return []byte{0x18} // Sequence prefix
	case keys.KeyCtrlY:
		return []byte{0x19} // Yank
	case keys.KeyCtrlZ:
		return []byte{0x1A} // Suspend (SIGTSTP)
	case keys.KeyCtrlBackslash:
		return []byte{0x1C} // SIGQUIT
	case keys.KeyCtrlCloseBracket:
		return []byte{0x1D} // Quit / search terminator
	case keys.KeyCtrlCaret:
		return []byte{0x1E} // Record separator
	case keys.KeyCtrlUnderscore:
		return []byte{0x1F} // Undo

	// ── Arrow keys (CSI sequences) ──
	case keys.KeyUp:
		return []byte{'\x1b', '[', 'A'}
	case keys.KeyDown:
		return []byte{'\x1b', '[', 'B'}
	case keys.KeyRight:
		return []byte{'\x1b', '[', 'C'}
	case keys.KeyLeft:
		return []byte{'\x1b', '[', 'D'}

	// ── Navigation / editing keys ──
	case keys.KeyHome:
		return []byte{'\x1b', '[', 'H'}
	case keys.KeyEnd:
		return []byte{'\x1b', '[', 'F'}
	case keys.KeyPgUp:
		return []byte{'\x1b', '[', '5', '~'}
	case keys.KeyPgDn:
		return []byte{'\x1b', '[', '6', '~'}
	case keys.KeyDelete:
		return []byte{'\x1b', '[', '3', '~'}

	// ── Function keys F1-F12 ──
	case keys.KeyF1:
		return []byte{'\x1b', '[', '1', '1', '~'}
	case keys.KeyF2:
		return []byte{'\x1b', '[', '1', '2', '~'}
	case keys.KeyF3:
		return []byte{'\x1b', '[', '1', '3', '~'}
	case keys.KeyF4:
		return []byte{'\x1b', '[', '1', '4', '~'}
	case keys.KeyF5:
		return []byte{'\x1b', '[', '1', '5', '~'}
	case keys.KeyF6:
		return []byte{'\x1b', '[', '1', '7', '~'}
	case keys.KeyF7:
		return []byte{'\x1b', '[', '1', '8', '~'}
	case keys.KeyF8:
		return []byte{'\x1b', '[', '1', '9', '~'}
	case keys.KeyF9:
		return []byte{'\x1b', '[', '2', '0', '~'}
	case keys.KeyF10:
		return []byte{'\x1b', '[', '2', '1', '~'}
	case keys.KeyF11:
		return []byte{'\x1b', '[', '2', '3', '~'}
	case keys.KeyF12:
		return []byte{'\x1b', '[', '2', '4', '~'}

	// ── Default: single printable character ──
	default:
		if len(key) == 1 {
			return []byte{key[0]}
		}
		return nil
	}
}
