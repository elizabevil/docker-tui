package keyboard

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func openResourceCreate(m *state.AppModel, resourceType runtimeapi.ResourceType) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	m.Dialog.Open(state.DialogSpec{Kind: state.DialogResourceCreate, Title: i18n.T("resource.create.title", resourceTypeLabel(resourceType)), Body: string(resourceType)})
	m.Navigation.Mode = state.ModeResourceCreate
	return m, nil
}

func handleResourceCreateKey(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if key == keys.KeyEsc {
		clearDialogState(m)
		return m, nil
	}
	if key != keys.KeyEnter {
		editQueryInput(key, &m.Dialog.Input)
		return m, nil
	}
	name := strings.TrimSpace(m.Dialog.Input.Text)
	if !validResourceName(name) {
		ShowToastWarn(m, i18n.T("resource.create.invalid"))
		return m, nil
	}
	resourceType := runtimeapi.ResourceType(m.Dialog.Body)
	trace := beginAudit(m, "resource."+string(resourceType)+".create", auditTargetForCreate(resourceType, name), "Create "+string(resourceType)+" "+name)
	clearDialogState(m)
	return m, withGenericAudit(resourceCreateCmd(m, resourceType, name), trace)
}

func validResourceName(name string) bool {
	if name == "" {
		return false
	}
	for _, char := range name {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || strings.ContainsRune("._-", char) {
			continue
		}
		return false
	}
	return true
}

func resourceCreateCmd(m *state.AppModel, resourceType runtimeapi.ResourceType, name string) tea.Cmd {
	client := m.Connection.Engine
	return func() tea.Msg {
		var err error
		switch resourceType {
		case runtimeapi.ResourceVolume:
			_, err = client.Volumes().Create(context.Background(), runtimeapi.VolumeCreateOptions{Name: name})
		case runtimeapi.ResourceNetwork:
			_, err = client.Networks().Create(context.Background(), runtimeapi.NetworkCreateOptions{Name: name})
		default:
			err = runtimeapi.NewError(runtimeapi.ErrorInvalid, "resource.create", name, fmt.Errorf("unsupported resource type %q", resourceType))
		}
		return state.GenericActioned{Action: state.ActionCreated, ID: name, Success: err == nil, Error: err}
	}
}

func confirmResourcePrune(m *state.AppModel, resourceType runtimeapi.ResourceType) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	message := i18n.T("resource.prune.confirm", resourceTypeLabel(resourceType))
	trace := beginAudit(m, "resource."+string(resourceType)+".prune", auditTargetForCreate(resourceType, "unused"), message)
	m.Confirm.Open(string(resourceType)+"-prune", string(resourceType), message, trace)
	m.Navigation.Mode = state.ModeConfirm
	return m, nil
}

func resourceTypeLabel(resourceType runtimeapi.ResourceType) string {
	if resourceType == runtimeapi.ResourceVolume {
		return i18n.T("panel.volumes")
	}
	return i18n.T("panel.networks")
}

func resourcePruneCmd(m *state.AppModel, resourceType runtimeapi.ResourceType, trace audit.Trace) tea.Cmd {
	client := m.Connection.Engine
	return func() tea.Msg {
		var result runtimeapi.PruneResult
		var err error
		switch resourceType {
		case runtimeapi.ResourceVolume:
			result, err = client.Volumes().Prune(context.Background(), runtimeapi.PruneOptions{})
		case runtimeapi.ResourceNetwork:
			result, err = client.Networks().Prune(context.Background(), runtimeapi.PruneOptions{})
		default:
			err = runtimeapi.NewError(runtimeapi.ErrorInvalid, "resource.prune", "", fmt.Errorf("unsupported resource type %q", resourceType))
		}
		if err == nil {
			var failures []error
			for _, resource := range result.Resources {
				if resource.Error != nil {
					failures = append(failures, fmt.Errorf("%s: %w", resource.ID, resource.Error))
				}
			}
			if len(failures) > 0 {
				err = errors.Join(failures...)
			}
		}
		return state.ResourcePruned{ResourceType: resourceType, Result: result, Error: err, Audit: trace}
	}
}

func auditTargetForCreate(resourceType runtimeapi.ResourceType, name string) audit.Target {
	if resourceType == runtimeapi.ResourceVolume {
		return audit.VolumeTarget{Name: name}
	}
	return audit.NetworkTarget{ID: name, Name: name}
}
