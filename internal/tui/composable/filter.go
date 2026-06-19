// Package composable provides reusable logic composables (Vue composables style).
package composable

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// StartSearchDebounce resets the search timer and returns a SearchTick cmd.
// Call this on every keypress in filter mode. After 1s of no input the timer
// fires and the filter is applied.
func StartSearchDebounce(m *state.AppModel) tea.Cmd {
	m.SearchTimer = 10 // 10 × 100ms = 1s
	return func() tea.Msg {
		time.Sleep(100 * time.Millisecond)
		return state.SearchTick{}
	}
}
