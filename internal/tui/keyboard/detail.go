package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// handleDetailKeys handles key presses in detail view mode.
// Fold/collapse functionality removed — all sections always expanded.
// Space now acts as Page Down for scrolling.
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
			m.DetailOffset++
			return true
		case keys.ActionUp:
			if m.DetailOffset > 0 {
				m.DetailOffset--
			}
			return true
		}
	}
	switch key {
	case keys.KeySpace, keys.KeyPgDn:
		m.DetailOffset += 20
	case keys.KeyPgUp:
		m.DetailOffset -= 20
		if m.DetailOffset < 0 {
			m.DetailOffset = 0
		}
	case keys.KeyG:
		m.DetailOffset = 0
	}
	return true
}
