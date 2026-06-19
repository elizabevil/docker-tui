package keyboard

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
	case "enter":
		return []byte{'\r'} // 0x0D
	case "backspace", "ctrl+h":
		return []byte{0x08} // BS
	case "tab", "ctrl+i":
		return []byte{'\t'} // 0x09
	case "esc", "ctrl+[":
		return []byte{'\x1b'} // Escape
	case "space", " ":
		return []byte{' '}

	// ── Ctrl+letter (0x01-0x1A) — shell essential ──
	case "ctrl+a":
		return []byte{0x01} // Home / beginning of line
	case "ctrl+b":
		return []byte{0x02} // Backward char
	case "ctrl+c":
		return []byte{0x03} // SIGINT / interrupt
	case "ctrl+d":
		return []byte{0x04} // EOF / exit
	case "ctrl+e":
		return []byte{0x05} // End of line
	case "ctrl+f":
		return []byte{0x06} // Forward char
	case "ctrl+g":
		return []byte{0x07} // Bell / abort
	case "ctrl+j":
		return []byte{0x0A} // Line feed
	case "ctrl+k":
		return []byte{0x0B} // Kill to end of line
	case "ctrl+l":
		return []byte{0x0C} // Clear screen (form feed)
	case "ctrl+m":
		return []byte{'\r'} // 0x0D — same as Enter
	case "ctrl+n":
		return []byte{0x0E} // Next history
	case "ctrl+o":
		return []byte{0x0F} // Operate / newline
	case "ctrl+p":
		return []byte{0x10} // Previous history
	case "ctrl+q":
		return []byte{0x11} // XON / resume flow
	case "ctrl+r":
		return []byte{0x12} // Reverse search
	case "ctrl+s":
		return []byte{0x13} // XOFF / stop flow
	case "ctrl+t":
		return []byte{0x14} // Transpose chars
	case "ctrl+u":
		return []byte{0x15} // Kill to beginning of line
	case "ctrl+v":
		return []byte{0x16} // Literal next / paste
	case "ctrl+w":
		return []byte{0x17} // Kill word backward
	case "ctrl+x":
		return []byte{0x18} // Sequence prefix
	case "ctrl+y":
		return []byte{0x19} // Yank
	case "ctrl+z":
		return []byte{0x1A} // Suspend (SIGTSTP)
	case "ctrl+\\":
		return []byte{0x1C} // SIGQUIT
	case "ctrl+]":
		return []byte{0x1D} // Quit / search terminator
	case "ctrl+^":
		return []byte{0x1E} // Record separator
	case "ctrl+_":
		return []byte{0x1F} // Undo

	// ── Arrow keys (CSI sequences) ──
	case "up":
		return []byte{'\x1b', '[', 'A'}
	case "down":
		return []byte{'\x1b', '[', 'B'}
	case "right":
		return []byte{'\x1b', '[', 'C'}
	case "left":
		return []byte{'\x1b', '[', 'D'}

	// ── Navigation / editing keys ──
	case "home":
		return []byte{'\x1b', '[', 'H'}
	case "end":
		return []byte{'\x1b', '[', 'F'}
	case "pgup":
		return []byte{'\x1b', '[', '5', '~'}
	case "pgdown":
		return []byte{'\x1b', '[', '6', '~'}
	case "delete":
		return []byte{'\x1b', '[', '3', '~'}

	// ── Function keys F1-F12 ──
	case "f1":
		return []byte{'\x1b', '[', '1', '1', '~'}
	case "f2":
		return []byte{'\x1b', '[', '1', '2', '~'}
	case "f3":
		return []byte{'\x1b', '[', '1', '3', '~'}
	case "f4":
		return []byte{'\x1b', '[', '1', '4', '~'}
	case "f5":
		return []byte{'\x1b', '[', '1', '5', '~'}
	case "f6":
		return []byte{'\x1b', '[', '1', '7', '~'}
	case "f7":
		return []byte{'\x1b', '[', '1', '8', '~'}
	case "f8":
		return []byte{'\x1b', '[', '1', '9', '~'}
	case "f9":
		return []byte{'\x1b', '[', '2', '0', '~'}
	case "f10":
		return []byte{'\x1b', '[', '2', '1', '~'}
	case "f11":
		return []byte{'\x1b', '[', '2', '3', '~'}
	case "f12":
		return []byte{'\x1b', '[', '2', '4', '~'}

	// ── Default: single printable character ──
	default:
		if len(key) == 1 {
			return []byte{key[0]}
		}
		return nil
	}
}
