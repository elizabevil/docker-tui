package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func handleLogKeys(key string, m *state.AppModel) bool {
	if m.Mode != state.ModeLogView {
		return false
	}
	switch key {
	case keys.KeyEsc:
		m.Mode = state.ModeNormal
		m.LogViewOffset = 0
		m.LogSearchText = ""
	case keys.KeyJ, keys.KeyDown:
		m.LogViewOffset++
	case keys.KeyK, keys.KeyUp:
		if m.LogViewOffset > 0 {
			m.LogViewOffset--
		}
	case keys.KeyPgDn:
		m.LogViewOffset += 20
	case keys.KeyPgUp:
		m.LogViewOffset -= 20
		if m.LogViewOffset < 0 {
			m.LogViewOffset = 0
		}
	case keys.KeyG:
		m.LogViewOffset = 0
	case keys.KeyCtrlG:
		m.LogViewOffset = len(m.LogContent)
	case keys.KeyN:
		if m.LogSearchText != "" {
			m.LogSearchMatch++
			scrollToMatch(m)
		}
	case keys.KeyCtrlN:
		if m.LogSearchText != "" {
			m.LogSearchMatch--
			if m.LogSearchMatch < 0 {
				m.LogSearchMatch = 0
			}
			scrollToMatch(m)
		}
	case keys.KeyW:
		m.LogWrapEnabled = !m.LogWrapEnabled
	}
	return true
}

// scrollToMatch scrolls the log view to the current search match line.
func scrollToMatch(m *state.AppModel) {
	count := 0
	for i, line := range m.LogContent {
		if contains(line, m.LogSearchText) {
			if count == m.LogSearchMatch {
				m.LogViewOffset = i
				return
			}
			count++
		}
	}
}

func contains(s, substr string) bool {
	if substr == "" {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
