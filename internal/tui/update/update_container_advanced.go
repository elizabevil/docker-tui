package update

import (
	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/utils"

	tea "charm.land/bubbletea/v2"
)

// handleContainerExportDone projects the Export result: a toast with the tar
// path and byte count on success, an error toast + failed audit on failure.
func handleContainerExportDone(m *state.AppModel, msg keyboard.ContainerExportDone) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		display := i18n.T("container.export.result.failed", msg.Error.Error())
		m.Feedback.RecordError(display)
		keyboard.ShowToastWarn(m, display)
		keyboard.FinishAudit(m, msg.Audit, audit.ResultFailed, "Container export failed", audit.Details{Error: msg.Error.Error()})
		return m, nil
	}
	display := i18n.T("container.export.result.success", msg.Destination, utils.FormatBytes(float64(msg.Bytes)))
	keyboard.FinishAudit(m, msg.Audit, audit.ResultSucceeded, "Container export completed: "+msg.Destination, audit.Details{})
	keyboard.ShowToastNow(m, display)
	return m, nil
}

// handleContainerCopyDone concludes the Copy result: the destination tar path
// and byte count on success, an error toast + failed audit on failure.
func handleContainerCopyDone(m *state.AppModel, msg keyboard.ContainerCopyDone) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		display := i18n.T("container.copy.result.failed", msg.Error.Error())
		m.Feedback.RecordError(display)
		keyboard.ShowToastWarn(m, display)
		keyboard.FinishAudit(m, msg.Audit, audit.ResultFailed, "Container copy failed", audit.Details{Error: msg.Error.Error()})
		return m, nil
	}
	display := i18n.T("container.copy.result.success", msg.Destination, utils.FormatBytes(float64(msg.Bytes)))
	keyboard.FinishAudit(m, msg.Audit, audit.ResultSucceeded, "Container copy completed: "+msg.Destination, audit.Details{})
	keyboard.ShowToastNow(m, display)
	return m, nil
}

// handleContainerUpdateDone concludes the Update result: a completion toast on
// success, or a failure toast + failed audit when the engine rejected it.
func handleContainerUpdateDone(m *state.AppModel, msg keyboard.ContainerUpdateDone) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		display := i18n.T("container.update.result.failed", msg.Error.Error())
		m.Feedback.RecordError(display)
		keyboard.ShowToastWarn(m, display)
		keyboard.FinishAudit(m, msg.Audit, audit.ResultFailed, "Container update failed", audit.Details{Error: msg.Error.Error()})
		return m, nil
	}
	display := i18n.T("container.update.result.success", shortMessageID(msg.ContainerID))
	keyboard.FinishAudit(m, msg.Audit, audit.ResultSucceeded, "Container update completed", audit.Details{})
	keyboard.ShowToastNow(m, display)
	return m, nil
}

// handleContainerCommitDone concludes the Commit result: it surfaces the new
// image ID on success and refreshes the image list so the committed image
// shows up immediately.
func handleContainerCommitDone(m *state.AppModel, msg keyboard.ContainerCommitDone) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		display := i18n.T("container.commit.result.failed", msg.Error.Error())
		m.Feedback.RecordError(display)
		keyboard.ShowToastWarn(m, display)
		keyboard.FinishAudit(m, msg.Audit, audit.ResultFailed, "Container commit failed", audit.Details{Error: msg.Error.Error()})
		return m, nil
	}
	display := i18n.T("container.commit.result.success", shortMessageID(msg.ImageID))
	keyboard.FinishAudit(m, msg.Audit, audit.ResultSucceeded, "Container commit completed", audit.Details{})
	keyboard.ShowToastNow(m, display)
	if m.Connection.Engine != nil {
		return m, keyboard.FetchImages(m.Connection.Engine)
	}
	return m, nil
}
