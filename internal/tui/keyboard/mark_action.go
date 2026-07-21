package keyboard

import (
	"fmt"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func doToggleMark(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	var id string
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		if ctr := m.Resources.Containers.Selected(); ctr != nil {
			id = ctr.ID
		}
	case state.PanelImages:
		if img := m.Resources.Images.Selected(); img != nil {
			id = img.ID
		}
	case state.PanelVolumes:
		if vol := m.Resources.Volumes.Selected(); vol != nil {
			id = vol.Name
		}
	case state.PanelNetworks:
		if net := m.Resources.Networks.Selected(); net != nil {
			id = net.ID
		}
	}
	if id == "" {
		return m, nil
	}
	m.Selection.Toggle(id)
	return m, nil
}

func doBulkDelete(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Docker == nil || len(m.Selection.MarkedIDs) == 0 {
		return m, nil
	}
	m.Confirm.ConfirmAction = "bulk-delete"
	m.Confirm.ConfirmTarget = fmt.Sprintf("%d items", len(m.Selection.MarkedIDs))
	m.Confirm.ConfirmMessage = fmt.Sprintf("Delete %d items?", len(m.Selection.MarkedIDs))
	m.Confirm.ConfirmAudit = beginAudit(m, "resource."+bulkResourceName(m.Navigation.ActivePanel)+".delete", bulkTarget(m), m.Confirm.ConfirmMessage)
	m.Navigation.Mode = state.ModeConfirm
	return m, nil
}

func doConfirmYes(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	action := m.Confirm.ConfirmAction
	target := m.Confirm.ConfirmTarget
	trace := m.Confirm.ConfirmAudit
	m.Navigation.Mode = state.ModeNormal
	m.Confirm.ConfirmAction = ""
	m.Confirm.ConfirmTarget = ""
	m.Confirm.ConfirmMessage = ""
	m.Confirm.ConfirmAudit = audit.Trace{}

	switch {
	case strings.HasPrefix(action, "batch-"):
		return executeBatchAction(m, strings.TrimPrefix(action, "batch-"), trace)
	case action == "bulk-delete":
		return executeBulkDelete(m, trace)
	case action == "container-stop":
		return m, withContainerAudit(containerStopCmd(m.Connection.Docker, target), trace)
	case action == "container-kill":
		return m, withContainerAudit(containerKillCmd(m.Connection.Docker, target), trace)
	case action == "container-restart":
		return m, withContainerAudit(containerRestartCmd(m.Connection.Docker, target), trace)
	case action == "container-remove":
		return m, withContainerAudit(containerRemoveCmd(m.Connection.Docker, target, true), trace)
	case action == "image-remove":
		return m, withImageAudit(imageRemoveCmd(m.Connection.Docker, target, true), trace)
	case action == "volume-remove":
		return m, withGenericAudit(volumeRemoveCmd(m.Connection.Docker, target, true), trace)
	case action == "network-remove":
		return m, withGenericAudit(networkRemoveCmd(m.Connection.Docker, target), trace)
	}
	return m, nil
}

func executeBulkDelete(m *state.AppModel, trace audit.Trace) (*state.AppModel, tea.Cmd) {
	if len(m.Selection.MarkedIDs) == 0 {
		return m, nil
	}
	ids := make([]string, 0, len(m.Selection.MarkedIDs))
	for id := range m.Selection.MarkedIDs {
		ids = append(ids, id)
	}
	m.Selection.MarkedIDs = make(map[string]bool)

	var cmds []tea.Cmd
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		for _, id := range ids {
			cmds = append(cmds, withContainerAudit(containerRemoveCmd(m.Connection.Docker, id, true), trace))
		}
	case state.PanelImages:
		for _, id := range ids {
			cmds = append(cmds, withImageAudit(imageRemoveCmd(m.Connection.Docker, id, true), trace))
		}
	case state.PanelVolumes:
		for _, id := range ids {
			cmds = append(cmds, withGenericAudit(volumeRemoveCmd(m.Connection.Docker, id, true), trace))
		}
	case state.PanelNetworks:
		for _, id := range ids {
			cmds = append(cmds, withGenericAudit(networkRemoveCmd(m.Connection.Docker, id), trace))
		}
	}

	ShowToastNow(m, fmt.Sprintf("✓ Deleted %d items", len(ids)))
	return m, tea.Batch(cmds...)
}
