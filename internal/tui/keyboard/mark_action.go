package keyboard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
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
	target := fmt.Sprintf("%d items", len(m.Selection.MarkedIDs))
	message := fmt.Sprintf("Delete %d items?", len(m.Selection.MarkedIDs))
	trace := beginAudit(m, "resource."+bulkResourceName(m.Navigation.ActivePanel)+".delete", bulkTarget(m), message)
	m.Confirm.Open(keys.ShowBulkDelete, target, message, trace)
	m.Confirm.Options = bulkDeleteOptions(m.Navigation.ActivePanel)
	m.Navigation.Mode = state.ModeConfirm
	return m, nil
}

func bulkDeleteOptions(panel state.PanelType) []state.ChoiceOption {
	options := []state.ChoiceOption{{ID: keys.ShowOptionCancel, Label: "Cancel"}}
	if panel != state.PanelNetworks {
		options = append(options, state.ChoiceOption{ID: keys.ShowOptionForce, Label: "Force", Description: "delete even if active or in use"})
	} else {
		options = append(options, state.ChoiceOption{ID: keys.ShowOptionConfirm, Label: "Delete"})
	}
	return options
}

func doConfirmYes(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	action := m.Confirm.ConfirmAction
	target := m.Confirm.ConfirmTarget
	trace := m.Confirm.ConfirmAudit
	returnMode := confirmReturnMode(m)
	m.Navigation.Mode = returnMode
	m.Confirm.Close()

	switch {
	case strings.HasPrefix(action, "batch-"):
		return executeBatchAction(m, strings.TrimPrefix(action, "batch-"), trace)
	case action == keys.ShowBulkDelete, action == keys.ShowBulkDeleteForce:
		return executeBulkDelete(m, trace, action == keys.ShowBulkDeleteForce)
	case action == keys.ShowContainerStop:
		return m, withContainerAudit(containerStopCmd(m.Connection.Engine, target), trace)
	case action == keys.ShowContainerKill:
		return m, withContainerAudit(containerKillCmd(m.Connection.Engine, target), trace)
	case action == keys.ShowContainerRestart:
		return m, withContainerAudit(containerRestartCmd(m.Connection.Engine, target), trace)
	case action == keys.ShowContainerRemove:
		return m, withContainerAudit(containerRemoveCmd(m.Connection.Engine, target, true, false, false), trace)
	case action == keys.ShowImageRemove:
		return m, withImageAudit(imageRemoveCmd(m.Connection.Engine, target, true, false, nil), trace)
	case action == keys.ShowVolumeRemove:
		return m, withGenericAudit(volumeRemoveCmd(m.Connection.Engine, target, true), trace)
	case action == keys.ShowNetworkRemove:
		return m, withGenericAudit(networkRemoveCmd(m.Connection.Engine, target), trace)
	case action == keys.ShowContainerCopy:
		src := m.Form.Get(fieldSourcePath).Text()
		dst := m.Form.Get(fieldDestinationPath).Text()
		id := m.Form.TargetID
		clearContainerForm(m)
		return m, withAdvancedAudit(containerCopyCmd(m.Connection.Engine, id, src, dst), trace)
	case action == keys.ShowContainerExport:
		dst := m.Form.Get(fieldDestinationPath).Text()
		id := m.Form.TargetID
		clearContainerForm(m)
		return m, withAdvancedAudit(containerExportCmd(m.Connection.Engine, id, dst), trace)
	case action == keys.ShowContainerCommitExport:
		opts, archivePath, err := containerCommitFormRequest(m)
		if err != nil {
			clearContainerForm(m)
			return m, nil
		}
		return executeContainerCommitForm(m, opts, archivePath, trace)
	case action == keys.ShowImageSave:
		path := m.Form.Get(fieldImagePath)
		if path == nil {
			clearContainerForm(m)
			return m, nil
		}
		request := runtimeapi.ImageTransferRequest{Operation: runtimeapi.ImageTransferSave, Source: m.Form.TargetID, Path: path.Text()}
		clearContainerForm(m)
		return beginImageTransfer(m, request)
	case action == keys.ShowVolumePrune:
		return m, resourcePruneCmd(m, runtimeapi.ResourceVolume, trace)
	case action == keys.ShowNetworkPrune:
		return m, resourcePruneCmd(m, runtimeapi.ResourceNetwork, trace)
	case action == keys.ShowEventsClear:
		m.EventPanel.Clear()
		return m, nil
	}
	return m, nil
}

func executeBulkDelete(m *state.AppModel, trace audit.Trace, force ...bool) (*state.AppModel, tea.Cmd) {
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
	forceDelete := len(force) > 0 && force[0]
	return m, func() tea.Msg {
		result := state.BatchActioned{
			Scope:    keys.ShowBulkDelete,
			Resource: bulkResourceType(panel),
			Total:    len(ids),
			Audit:    trace,
		}
		var failures []error
		for _, id := range ids {
			var msg tea.Msg
			switch panel {
			case state.PanelContainers:
				msg = containerRemoveCmd(engine, id, forceDelete, false, false)()
			case state.PanelImages:
				msg = imageRemoveCmd(engine, id, forceDelete, false, nil)()
			case state.PanelVolumes:
				msg = volumeRemoveCmd(engine, id, forceDelete)()
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
