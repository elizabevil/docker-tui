package keyboard

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func handleLogKeys(key string, m *state.AppModel) bool {
	if m.Mode != state.ModeLogView {
		return false
	}
	action, known := resolveAction(key, m)
	if known {
		switch action {
		case keys.ActionFilter:
			ToSearch(m)
			return true
		case keys.ActionBack:
			m.Mode = state.ModeNormal
			m.LogViewOffset = 0
			m.LogSearchText = ""
			return true
		case keys.ActionDown:
			m.LogViewOffset++
			return true
		case keys.ActionUp:
			if m.LogViewOffset > 0 {
				m.LogViewOffset--
			}
			return true
		}
	}
	switch key {
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
		if strings.Contains(strings.ToLower(line), strings.ToLower(m.LogSearchText)) {
			if count == m.LogSearchMatch {
				m.LogViewOffset = i
				return
			}
			count++
		}
	}
}

func logSearchMatchCount(m *state.AppModel) int {
	if m == nil || m.LogSearchText == "" {
		return 0
	}
	count := 0
	for _, line := range m.LogContent {
		if strings.Contains(strings.ToLower(line), strings.ToLower(m.LogSearchText)) {
			count++
		}
	}
	return count
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
