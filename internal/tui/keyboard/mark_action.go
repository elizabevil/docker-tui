package keyboard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
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
	if m.Connection.Engine == nil || len(m.Selection.MarkedIDs) == 0 {
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
		return m, withContainerAudit(containerStopCmd(m.Connection.Engine, target), trace)
	case action == "container-kill":
		return m, withContainerAudit(containerKillCmd(m.Connection.Engine, target), trace)
	case action == "container-restart":
		return m, withContainerAudit(containerRestartCmd(m.Connection.Engine, target), trace)
	case action == "container-remove":
		return m, withContainerAudit(containerRemoveCmd(m.Connection.Engine, target, true), trace)
	case action == "image-remove":
		return m, withImageAudit(imageRemoveCmd(m.Connection.Engine, target, true), trace)
	case action == "volume-remove":
		return m, withGenericAudit(volumeRemoveCmd(m.Connection.Engine, target, true), trace)
	case action == "network-remove":
		return m, withGenericAudit(networkRemoveCmd(m.Connection.Engine, target), trace)
	case action == "volume-prune":
		return m, resourcePruneCmd(m, runtimeapi.ResourceVolume, trace)
	case action == "network-prune":
		return m, resourcePruneCmd(m, runtimeapi.ResourceNetwork, trace)
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
	if len(ids) == 0 {
		return m, nil
	}

	engine := m.Connection.Engine
	panel := m.Navigation.ActivePanel
	return m, func() tea.Msg {
		result := state.BatchActioned{
			Scope:    "bulk-delete",
			Resource: bulkResourceName(panel),
			Total:    len(ids),
			Audit:    trace,
		}
		var failures []error
		for _, id := range ids {
			var msg tea.Msg
			switch panel {
			case state.PanelContainers:
				msg = containerRemoveCmd(engine, id, true)()
			case state.PanelImages:
				msg = imageRemoveCmd(engine, id, true)()
			case state.PanelVolumes:
				msg = volumeRemoveCmd(engine, id, true)()
			case state.PanelNetworks:
				msg = networkRemoveCmd(engine, id)()
			default:
				result.Failed++
				result.FailedIDs = append(result.FailedIDs, id)
				continue
			}
			switch v := msg.(type) {
			case state.ContainerActioned, state.ImageActioned, state.GenericActioned:
				var ok bool
				var err error
				switch x := v.(type) {
				case state.ContainerActioned:
					ok, err = x.Success, x.Error
				case state.ImageActioned:
					ok, err = x.Success, x.Error
				case state.GenericActioned:
					ok, err = x.Success, x.Error
				}
				if ok {
					result.Success++
				} else {
					result.Failed++
					result.FailedIDs = append(result.FailedIDs, id)
					if err != nil {
						failures = append(failures, err)
					}
				}
			default:
				result.Failed++
				result.FailedIDs = append(result.FailedIDs, id)
			}
		}
		result.Error = errors.Join(failures...)
		return result
	}
}
