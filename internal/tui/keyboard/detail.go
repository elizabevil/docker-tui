package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// handleDetailKeys handles key presses in detail view mode.
// Fold/collapse functionality removed — all sections always expanded.
// Space now acts as Page Down for scrolling.
// 's' cycles between section/yaml/json source views.
func handleDetailKeys(key string, m *state.AppModel) (bool, tea.Cmd) {
	if m.Navigation.Mode != state.ModeDetail {
		return false, nil
	}
	action, known := resolveAction(key, m)
	if known {
		switch action {
		case keys.ActionEnter, keys.ActionBack:
			BackFromDetail(m)
			return true, nil
		case keys.ActionDown:
			m.Detail.Scroll(1)
			return true, nil
		case keys.ActionUp:
			m.Detail.Scroll(-1)
			return true, nil
		}
	}
	switch key {
	case keys.KeySpace, keys.KeyPgDn:
		m.Detail.Scroll(20)
	case keys.KeyPgUp:
		m.Detail.Scroll(-20)
	case keys.KeyG:
		m.Detail.Scroll(-m.Detail.DetailOffset)
	case "s":
		m.Detail.CycleSource()
	case keys.KeyCtrlA:
		if m.Detail.DetailSourceType != state.DetailSourceSection && m.Detail.HasRawSource() {
			m.Detail.SourceSelected = true
		}
	case keys.KeyCtrlC:
		if m.Detail.SourceSelected && m.Detail.DetailSourceType != state.DetailSourceSection {
			if source := m.Detail.SourceText(); source != "" {
				ShowToastNow(m, "✓ Source copied")
				return true, clipboardCmd(source)
			}
		}
	}
	return true, nil
}
