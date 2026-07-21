package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// handleDetailKeys handles key presses in detail view mode.
// Fold/collapse functionality removed — all sections always expanded.
// Space now acts as Page Down for scrolling.
// 's' cycles between section/yaml/json source views.
func handleDetailKeys(key string, m *state.AppModel) bool {
	if m.Mode != state.ModeDetail {
		return false
	}
	action, known := resolveAction(key, m)
	if known {
		switch action {
		case keys.ActionEnter, keys.ActionBack:
			BackFromDetail(m)
			return true
		case keys.ActionDown:
			m.DetailState.Scroll(1)
			return true
		case keys.ActionUp:
			m.DetailState.Scroll(-1)
			return true
		}
	}
	switch key {
	case keys.KeySpace, keys.KeyPgDn:
		m.DetailState.Scroll(20)
	case keys.KeyPgUp:
		m.DetailState.Scroll(-20)
	case keys.KeyG:
		m.DetailState.Scroll(-m.DetailOffset)
	case "s":
		m.DetailState.CycleSource()
	}
	return true
}
