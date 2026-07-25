package update

import (
	"fmt"
	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
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
	if m.Connection.Engine != nil {
		return m, keyboard.FetchContainers(m.Connection.Engine, true)
	}
	return m, nil
}

// handleBatchActioned processes the aggregate batch summary emitted by
// container / image / volume / network / compose batch and bulk operations.
// TASK-010: collect per-target results into one summary so the user sees
// "succeeded/skipped/failed" instead of N independent toasts.
func handleBatchActioned(m *state.AppModel, msg state.BatchActioned) (*state.AppModel, tea.Cmd) {
	display := fmt.Sprintf("%s %d, %s %d, %s %d",
		i18n.T("batch.succeeded"), msg.Success,
		i18n.T("batch.skipped"), msg.Skipped,
		i18n.T("batch.failed"), msg.Failed,
	)
	result := audit.ResultSucceeded
	switch {
	case msg.Failed > 0 && msg.Success > 0:
		result = audit.ResultPartial
	case msg.Failed > 0:
		result = audit.ResultFailed
	case msg.Total > 0 && msg.Success == 0 && msg.Failed == 0 && msg.Skipped == 0:
		result = audit.ResultSucceeded
	}
	if msg.Audit.Valid() {
		keyboard.FinishAudit(m, msg.Audit, result, display, audit.Details{
			Error: errorText(msg.Error),
		})
	} else {
		keyboard.ShowToastNow(m, display)
	}
	if m.Connection.Engine == nil {
		return m, nil
	}
	// Refresh whichever resource lists the affected scope implies.
	switch msg.Resource {
	case state.ResourceContainer:
		return m, keyboard.FetchContainers(m.Connection.Engine, true)
	case state.ResourceImage:
		return m, keyboard.FetchImages(m.Connection.Engine)
	case state.ResourceVolume:
		return m, keyboard.FetchVolumes(m.Connection.Engine)
	case state.ResourceNetwork:
		return m, keyboard.FetchNetworks(m.Connection.Engine)
	case state.ResourceComposeProject:
		// Compose down/up/stop also touches containers; refresh those.
		return m, tea.Batch(
			keyboard.FetchContainers(m.Connection.Engine, true),
			keyboard.FetchVolumes(m.Connection.Engine),
			keyboard.FetchNetworks(m.Connection.Engine),
		)
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
	if m.Connection.Engine != nil {
		if msg.Action == state.ActionRenamed && msg.Error == nil {
			m.Resources.Containers.SelectionAnchorID = msg.ID
		}
		return m, keyboard.FetchContainers(m.Connection.Engine, true)
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
	if m.Connection.Engine != nil {
		return m, keyboard.FetchImages(m.Connection.Engine)
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
	if m.Connection.Engine != nil {
		return m, tea.Batch(keyboard.FetchAll(m.Connection.Engine)...)
	}
	return m, nil
}

func handleResourcePruned(m *state.AppModel, msg state.ResourcePruned) (*state.AppModel, tea.Cmd) {
	succeeded, failed := msg.Result.Counts()
	resourceLabel := i18n.T("panel.networks")
	if msg.ResourceType == runtimeapi.ResourceVolume {
		resourceLabel = i18n.T("panel.volumes")
	}
	display := i18n.T("resource.prune.result", resourceLabel, succeeded, failed)
	if msg.Result.SpaceReclaimed > 0 {
		display += i18n.T("resource.prune.reclaimed", msg.Result.SpaceReclaimed)
	}
	result := audit.ResultSucceeded
	if failed > 0 && succeeded > 0 {
		result = audit.ResultPartial
	} else if msg.Error != nil {
		result = audit.ResultFailed
		m.Feedback.RecordError(display + ": " + msg.Error.Error())
	} else {
		keyboard.ShowToastNow(m, display)
	}
	keyboard.FinishAudit(m, msg.Audit, result, display, audit.Details{Error: errorText(msg.Error)})
	if m.Connection.Engine != nil {
		return m, tea.Batch(keyboard.FetchAll(m.Connection.Engine)...)
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
