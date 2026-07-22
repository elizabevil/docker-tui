package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// handleAuditDetailKey handles key presses in audit detail view.
func handleAuditDetailKey(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch key {
	case keys.KeyEsc, keys.KeyEnter:
		m.Navigation.Mode = state.ModeNormal
		return m, nil
	}
	return m, nil
}

// handleAuditPanelKey handles key presses on the audit panel in normal mode.
func handleAuditPanelKey(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch key {
	case keys.KeyEnter:
		records := m.Audit.AuditListModel()
		if m.Audit.Cursor >= 0 && m.Audit.Cursor < len(records) {
			rec := records[m.Audit.Cursor]
			m.Audit.DetailRecord = &rec
			m.Audit.SelectedID = rec.TraceID
			m.Navigation.Mode = state.ModeAuditDetail
		}
		return m, nil
	case keys.KeyE:
		// Cycle filter level: All -> Errors -> Warnings -> All
		switch m.Audit.FilterLevel {
		case state.AuditFilterAll:
			m.Audit.FilterLevel = state.AuditFilterErrors
		case state.AuditFilterErrors:
			m.Audit.FilterLevel = state.AuditFilterWarnings
		case state.AuditFilterWarnings:
			m.Audit.FilterLevel = state.AuditFilterAll
		}
		m.Audit.Cursor = 0
		m.Audit.ViewOffset = 0
		return m, nil
	}
	return m, nil
}
