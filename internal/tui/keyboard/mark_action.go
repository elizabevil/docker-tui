package keyboard

import (
	"fmt"
	"strings"

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
	m.Mode = state.ModeConfirm
	return m, nil
}

func doConfirmYes(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	action := m.ConfirmAction
	target := m.ConfirmTarget
	m.Mode = state.ModeNormal
	m.ConfirmAction = ""
	m.ConfirmTarget = ""
	m.ConfirmMessage = ""

	switch {
	case strings.HasPrefix(action, "batch-"):
		return executeBatchAction(m, strings.TrimPrefix(action, "batch-"))
	case action == "bulk-delete":
		return executeBulkDelete(m)
	case action == "container-stop":
		return m, containerStopCmd(m.Docker, target)
	case action == "container-kill":
		return m, containerKillCmd(m.Docker, target)
	case action == "container-restart":
		return m, containerRestartCmd(m.Docker, target)
	case action == "container-remove":
		return m, containerRemoveCmd(m.Docker, target, true)
	case action == "image-remove":
		return m, imageRemoveCmd(m.Docker, target, true)
	case action == "volume-remove":
		return m, volumeRemoveCmd(m.Docker, target, true)
	case action == "network-remove":
		return m, networkRemoveCmd(m.Docker, target)
	}
	return m, nil
}

func executeBulkDelete(m *state.AppModel) (*state.AppModel, tea.Cmd) {
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
			cmds = append(cmds, containerRemoveCmd(m.Docker, id, true))
		}
	case state.PanelImages:
		for _, id := range ids {
			cmds = append(cmds, imageRemoveCmd(m.Docker, id, true))
		}
	case state.PanelVolumes:
		for _, id := range ids {
			cmds = append(cmds, volumeRemoveCmd(m.Docker, id, true))
		}
	case state.PanelNetworks:
		for _, id := range ids {
			cmds = append(cmds, networkRemoveCmd(m.Docker, id))
		}
	}

	ShowToastNow(m, fmt.Sprintf("✓ Deleted %d items", len(ids)))
	return m, tea.Batch(cmds...)
}
