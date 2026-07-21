package tui

import (
	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleContainerActioned(m *state.AppModel, msg state.ContainerActioned) (*state.AppModel, tea.Cmd) {
	display := "✕ failed: " + msg.Action
	if msg.Error == nil {
		display = "✓ " + i18n.T("toast."+msg.Action, shortAuditID(msg.ID))
	}
	if msg.Audit.Valid() {
		result := audit.ResultSucceeded
		if msg.Error != nil {
			result = audit.ResultFailed
		}
		keyboard.FinishAudit(m, msg.Audit, result, display, audit.Details{Error: errorText(msg.Error)})
	} else {
		keyboard.ShowToastNow(m, display)
	}
	if m.Docker != nil {
		return m, keyboard.FetchContainers(m.Docker, true)
	}
	return m, nil
}

func handleImageActioned(m *state.AppModel, msg state.ImageActioned) (*state.AppModel, tea.Cmd) {
	display := "✕ failed: " + msg.Action
	if msg.Error == nil {
		display = "✓ " + i18n.T("toast."+msg.Action, msg.Ref)
	}
	if msg.Audit.Valid() {
		result := audit.ResultSucceeded
		if msg.Error != nil {
			result = audit.ResultFailed
		}
		keyboard.FinishAudit(m, msg.Audit, result, display, audit.Details{Error: errorText(msg.Error)})
	} else {
		keyboard.ShowToastNow(m, display)
	}
	if m.Docker != nil {
		return m, keyboard.FetchImages(m.Docker)
	}
	return m, nil
}

func handleImageDetailLoaded(m *state.AppModel, msg state.ImageDetailLoaded) (*state.AppModel, tea.Cmd) {
	if msg.ImageID != m.ImageDetailID {
		return m, nil
	}
	if msg.Error != nil {
		m.FeedbackState.RecordError(msg.Error.Error())
	} else {
		m.ImageDetailData = msg.Detail
		if m.Mode == state.ModeDetail {
			m.DetailOffset = 0
		}
	}
	return m, nil
}

func handleGenericActioned(m *state.AppModel, msg state.GenericActioned) (*state.AppModel, tea.Cmd) {
	display := "✕ failed: " + msg.Action
	if msg.Error == nil {
		display = "✓ " + i18n.T("toast."+msg.Action, msg.ID)
	}
	if msg.Audit.Valid() {
		result := audit.ResultSucceeded
		if msg.Error != nil {
			result = audit.ResultFailed
		}
		keyboard.FinishAudit(m, msg.Audit, result, display, audit.Details{Error: errorText(msg.Error)})
	} else {
		keyboard.ShowToastNow(m, display)
	}
	if m.Docker != nil {
		return m, tea.Batch(keyboard.FetchAll(m.Docker)...)
	}
	return m, nil
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func shortAuditID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
