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
	switch key {
	case keys.KeyEnter, keys.KeyEsc:
		BackFromDetail(m)
	case keys.KeyJ, keys.KeyDown:
		m.DetailOffset++
	case keys.KeyK, keys.KeyUp:
		if m.DetailOffset > 0 {
			m.DetailOffset--
		}
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
