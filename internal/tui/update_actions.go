package tui

import (
	"fmt"
	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleContainerBatchActioned(m *state.AppModel, msg state.ContainerBatchActioned) (*state.AppModel, tea.Cmd) {
	display := fmt.Sprintf("%s %d, %s %d, %s %d", i18n.T("batch.succeeded"), msg.Success, i18n.T("batch.skipped"), msg.Skipped, i18n.T("batch.failed"), msg.Failed)
	result := audit.ResultSucceeded
	if msg.Failed > 0 {
		result = audit.ResultFailed
	}
	if msg.Audit.Valid() {
		keyboard.FinishAudit(m, msg.Audit, result, display, audit.Details{Error: errorText(msg.Error)})
	} else {
		keyboard.ShowToastNow(m, display)
	}
	if m.Connection.Docker != nil {
		return m, keyboard.FetchContainers(m.Connection.Docker, true)
	}
	return m, nil
}

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
	if m.Connection.Docker != nil {
		if msg.Action == state.ActionRenamed && msg.Error == nil {
			m.Resources.Containers.SelectionAnchorID = msg.ID
		}
		return m, keyboard.FetchContainers(m.Connection.Docker, true)
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
	if m.Connection.Docker != nil {
		return m, keyboard.FetchImages(m.Connection.Docker)
	}
	return m, nil
}

func handleImageDetailLoaded(m *state.AppModel, msg state.ImageDetailLoaded) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		if msg.ImageID != m.Detail.ImageDetailID {
			return m, nil
		}
		m.Feedback.RecordError(msg.Error.Error())
	} else if !m.Detail.ApplyImage(msg.ImageID, msg.Detail) {
		return m, nil
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
	if m.Connection.Docker != nil {
		return m, tea.Batch(keyboard.FetchAll(m.Connection.Docker)...)
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
