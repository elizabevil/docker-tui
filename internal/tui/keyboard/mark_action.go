package keyboard

import (
	"fmt"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func doToggleMark(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.MarkedIDs == nil {
		m.MarkedIDs = make(map[string]bool)
	}
	var id string
	switch m.ActivePanel {
	case state.PanelContainers:
		if ctr := m.Containers.Selected(); ctr != nil {
			id = ctr.ID
		}
	case state.PanelImages:
		if img := m.Images.Selected(); img != nil {
			id = img.ID
		}
	case state.PanelVolumes:
		if vol := m.Volumes.Selected(); vol != nil {
			id = vol.Name
		}
	case state.PanelNetworks:
		if net := m.Networks.Selected(); net != nil {
			id = net.ID
		}
	}
	if id == "" {
		return m, nil
	}
	if m.MarkedIDs[id] {
		delete(m.MarkedIDs, id)
	} else {
		m.MarkedIDs[id] = true
	}
	return m, nil
}

func doBulkDelete(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil || len(m.MarkedIDs) == 0 {
		return m, nil
	}
	m.ConfirmAction = "bulk-delete"
	m.ConfirmTarget = fmt.Sprintf("%d items", len(m.MarkedIDs))
	m.ConfirmMessage = fmt.Sprintf("Delete %d items?", len(m.MarkedIDs))
	m.ConfirmAudit = beginAudit(m, "resource."+bulkResourceName(m.ActivePanel)+".delete", bulkTarget(m), m.ConfirmMessage)
	m.Mode = state.ModeConfirm
	return m, nil
}

func doConfirmYes(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	action := m.ConfirmAction
	target := m.ConfirmTarget
	trace := m.ConfirmAudit
	m.Mode = state.ModeNormal
	m.ConfirmAction = ""
	m.ConfirmTarget = ""
	m.ConfirmMessage = ""
	m.ConfirmAudit = audit.Trace{}

	switch {
	case strings.HasPrefix(action, "batch-"):
		return executeBatchAction(m, strings.TrimPrefix(action, "batch-"), trace)
	case action == "bulk-delete":
		return executeBulkDelete(m, trace)
	case action == "container-stop":
		return m, withContainerAudit(containerStopCmd(m.Docker, target), trace)
	case action == "container-kill":
		return m, withContainerAudit(containerKillCmd(m.Docker, target), trace)
	case action == "container-restart":
		return m, withContainerAudit(containerRestartCmd(m.Docker, target), trace)
	case action == "container-remove":
		return m, withContainerAudit(containerRemoveCmd(m.Docker, target, true), trace)
	case action == "image-remove":
		return m, withImageAudit(imageRemoveCmd(m.Docker, target, true), trace)
	case action == "volume-remove":
		return m, withGenericAudit(volumeRemoveCmd(m.Docker, target, true), trace)
	case action == "network-remove":
		return m, withGenericAudit(networkRemoveCmd(m.Docker, target), trace)
	}
	return m, nil
}

func executeBulkDelete(m *state.AppModel, trace audit.Trace) (*state.AppModel, tea.Cmd) {
	if len(m.MarkedIDs) == 0 {
		return m, nil
	}
	ids := make([]string, 0, len(m.MarkedIDs))
	for id := range m.MarkedIDs {
		ids = append(ids, id)
	}
	m.MarkedIDs = make(map[string]bool)

	var cmds []tea.Cmd
	switch m.ActivePanel {
	case state.PanelContainers:
		for _, id := range ids {
			cmds = append(cmds, withContainerAudit(containerRemoveCmd(m.Docker, id, true), trace))
		}
	case state.PanelImages:
		for _, id := range ids {
			cmds = append(cmds, withImageAudit(imageRemoveCmd(m.Docker, id, true), trace))
		}
	case state.PanelVolumes:
		for _, id := range ids {
			cmds = append(cmds, withGenericAudit(volumeRemoveCmd(m.Docker, id, true), trace))
		}
	case state.PanelNetworks:
		for _, id := range ids {
			cmds = append(cmds, withGenericAudit(networkRemoveCmd(m.Docker, id), trace))
		}
	}

	ShowToastNow(m, fmt.Sprintf("✓ Deleted %d items", len(ids)))
	return m, tea.Batch(cmds...)
}
