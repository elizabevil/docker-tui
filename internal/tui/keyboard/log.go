package keyboard

import (
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
			m.LogState.Close()
			return true
		case keys.ActionDown:
			m.LogState.Scroll(1)
			return true
		case keys.ActionUp:
			m.LogState.Scroll(-1)
			return true
		}
	}
	switch key {
	case keys.KeyPgDn:
		m.LogState.Scroll(20)
	case keys.KeyPgUp:
		m.LogState.Scroll(-20)
	case keys.KeyG:
		m.LogState.ToTop()
	case keys.KeyCtrlG:
		m.LogState.ToBottom()
	case keys.KeyN:
		m.LogState.MoveMatch(1)
	case keys.KeyCtrlN:
		m.LogState.MoveMatch(-1)
	case keys.KeyW:
		m.LogState.ToggleWrap()
	}
	return true
}

func logSearchMatchCount(m *state.AppModel) int {
	if m == nil {
		return 0
	}
	return m.LogState.MatchCount()
}
