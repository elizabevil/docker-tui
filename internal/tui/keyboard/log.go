package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func handleLogKeys(key string, m *state.AppModel) bool {
	if m.Navigation.Mode != state.ModeLogView {
		return false
	}
	action, known := resolveAction(key, m)
	if known {
		switch action {
		case keys.ActionFilter:
			ToSearch(m)
			return true
		case keys.ActionBack:
			m.Navigation.Mode = state.ModeNormal
			m.Log.Close()
			return true
		case keys.ActionDown:
			m.Log.Scroll(1)
			return true
		case keys.ActionUp:
			m.Log.Scroll(-1)
			return true
		}
	}
	switch key {
	case keys.KeyPgDn:
		m.Log.Scroll(20)
	case keys.KeyPgUp:
		m.Log.Scroll(-20)
	case keys.KeyG:
		m.Log.ToTop()
	case keys.KeyCtrlG:
		m.Log.ToBottom()
	case keys.KeyN:
		m.Log.MoveMatch(1)
	case keys.KeyCtrlN:
		m.Log.MoveMatch(-1)
	case keys.KeyW:
		m.Log.ToggleWrap()
	}
	return true
}
