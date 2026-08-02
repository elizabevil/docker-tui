package keyboard

import (
	"strings"
	"unicode/utf8"

	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleHistoryKeys(rawKey string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Navigation.Mode != state.ModeHistory {
		return m, nil
	}
	visible := historyVisibleRows(m)
	items := historyVisibleItems(m)
	key := keys.Normalize(rawKey)
	if m.History.Filtering {
		switch key {
		case keys.KeyEsc, keys.KeyEnter:
			m.History.Filtering = false
		case keys.KeyBackspace:
			filter := []rune(m.History.Filter)
			if len(filter) > 0 {
				m.History.SetFilter(string(filter[:len(filter)-1]))
			}
		default:
			if utf8.RuneCountInString(rawKey) == 1 {
				m.History.SetFilter(m.History.Filter + rawKey)
			}
		}
		return m, nil
	}
	if rawKey == keys.KeyGUpper {
		m.History.MoveCursor(len(items), len(items), visible)
		return m, nil
	}
	switch key {
	case keys.KeyEsc:
		m.History.Close()
		m.Navigation.Mode = state.ModeNormal
		return m, nil
	case keys.KeyJ, keys.KeyDown:
		m.History.MoveCursor(+1, len(items), visible)
		return m, nil
	case keys.KeyK, keys.KeyUp:
		m.History.MoveCursor(-1, len(items), visible)
		return m, nil
	case keys.KeyPgDn:
		m.History.MoveCursor(visible, len(items), visible)
		return m, nil
	case keys.KeyPgUp:
		m.History.MoveCursor(-visible, len(items), visible)
		return m, nil
	case keys.KeyG, keys.KeyHome:
		m.History.MoveCursor(-m.History.Cursor, len(items), visible)
		return m, nil
	case keys.KeyEnd:
		m.History.MoveCursor(len(items), len(items), visible)
		return m, nil
	case keys.KeySlash:
		m.History.Filtering = true
		return m, nil
	}
	return m, nil
}

func historyVisibleItems(m *state.AppModel) []filteredItem {
	return filterHistoryLayers(m)
}

type filteredItem struct {
	index int
}

func filterHistoryLayers(m *state.AppModel) []filteredItem {
	items := m.History.Layers
	filter := m.History.Filter
	out := make([]filteredItem, 0, len(items))
	if filter == "" {
		for i := range items {
			out = append(out, filteredItem{index: i})
		}
		return out
	}
	for i, l := range items {
		if strings.Contains(strings.ToLower(l.CreatedBy), strings.ToLower(filter)) ||
			strings.Contains(strings.ToLower(l.Comment), strings.ToLower(filter)) {
			out = append(out, filteredItem{index: i})
		}
	}
	return out
}

func historyVisibleRows(m *state.AppModel) int {
	// Standard layout reserves 11 rows for rails and 3 for panel chrome.
	return max(1, m.Viewport.Height-14-4)
}
