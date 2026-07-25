package update

import (
	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleImageTransferReceived(m *state.AppModel, msg state.ImageTransferReceived) (*state.AppModel, tea.Cmd) {
	if !m.ImageTransfer.Current(msg.Generation) {
		return m, nil
	}
	if msg.Event.Progress != nil {
		m.ImageTransfer.Progress = *msg.Event.Progress
	}
	if !msg.Event.Done {
		return m, keyboard.ReadImageTransferCmd(msg.Generation, msg.Events)
	}

	request := m.ImageTransfer.Request
	trace := m.ImageTransfer.Audit
	m.ImageTransfer.Reset()
	m.Dialog.Close()
	m.Navigation.Mode = state.ModeNormal
	result := audit.ResultSucceeded
	display := i18n.T("image.transfer.completed", i18n.T("key."+string(request.Operation)))
	if msg.Event.Error != nil {
		result = audit.ResultFailed
		display = i18n.T("image.transfer.failed", i18n.T("key."+string(request.Operation)), msg.Event.Error)
		m.Feedback.RecordError(display)
		keyboard.ShowToastWarn(m, display)
	} else {
		keyboard.ShowToastNow(m, display)
	}
	keyboard.FinishAudit(m, trace, result, display, audit.Details{Error: errorText(msg.Event.Error)})
	if m.Connection.Engine != nil {
		return m, keyboard.FetchImages(m.Connection.Engine)
	}
	return m, nil
}
