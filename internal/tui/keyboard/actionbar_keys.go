package keyboard

import (
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/elizabevil/docker-tui/internal/tui/actionbar"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func handleActionBarKeys(rawKey string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Navigation.Mode != state.ModeActionBar {
		return m, nil
	}
	key := keys.Normalize(rawKey)
	items := actionbar.VisibleItems(m)
	clampActionBarSelection(m, items)

	if m.Navigation.ActionBar.Filtering {
		switch key {
		case keys.KeyEsc:
			m.Navigation.ActionBar.ExitFilter()
		case keys.KeyEnter:
			return executeActionBarSelection(m, items)
		case keys.KeyBackspace:
			filter := []rune(m.Navigation.ActionBar.Filter)
			if len(filter) > 0 {
				m.Navigation.ActionBar.SetFilter(string(filter[:len(filter)-1]))
			}
		default:
			if utf8.RuneCountInString(rawKey) == 1 {
				m.Navigation.ActionBar.SetFilter(m.Navigation.ActionBar.Filter + rawKey)
			}
		}
		return m, nil
	}

	switch key {
	case keys.KeyEsc:
		closeActionBar(m)
	case keys.KeyEnter:
		return executeActionBarSelection(m, items)
	case keys.KeyJ, keys.KeyDown:
		moveActionBarSelection(m, items, 1)
	case keys.KeyK, keys.KeyUp:
		moveActionBarSelection(m, items, -1)
	case keys.KeySlash:
		m.Navigation.ActionBar.EnterFilter()
	default:
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
			index := int(key[0] - '1')
			if index < len(items) && !items[index].Disabled {
				m.Navigation.ActionBar.Selected = index
			}
		}
	}
	return m, nil
}

func executeActionBarSelection(m *state.AppModel, items []actionbar.ActionItem) (*state.AppModel, tea.Cmd) {
	clampActionBarSelection(m, items)
	if len(items) == 0 || items[m.Navigation.ActionBar.Selected].Disabled {
		return m, nil
	}
	action := items[m.Navigation.ActionBar.Selected].Action
	closeActionBar(m)
	return handleAction(action, m, nil)
}

func closeActionBar(m *state.AppModel) {
	m.Navigation.ActionBar.Close()
	m.Navigation.Mode = state.ModeNormal
}

func clampActionBarSelection(m *state.AppModel, items []actionbar.ActionItem) {
	if len(items) == 0 {
		m.Navigation.ActionBar.Selected = 0
		return
	}
	if m.Navigation.ActionBar.Selected < 0 || m.Navigation.ActionBar.Selected >= len(items) {
		m.Navigation.ActionBar.Selected = 0
	}
	if items[m.Navigation.ActionBar.Selected].Disabled {
		moveActionBarSelection(m, items, 1)
	}
}

func moveActionBarSelection(m *state.AppModel, items []actionbar.ActionItem, delta int) {
	if len(items) == 0 {
		m.Navigation.ActionBar.Selected = 0
		return
	}
	for range items {
		m.Navigation.ActionBar.MoveSelection(delta, len(items))
		if !items[m.Navigation.ActionBar.Selected].Disabled {
			return
		}
	}
}
