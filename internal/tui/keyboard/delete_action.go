package keyboard

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func doDeleteAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	if len(m.Selection.MarkedIDs) > 0 {
		return doBulkDelete(m)
	}
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		ctr := m.Resources.Containers.Selected()
		if ctr != nil {
			confirmAction(m, keys.ShowContainerRemove, ctr.ID, fmt.Sprintf("Remove container %s?", ctr.Name))
		}
	case state.PanelImages:
		img := m.Resources.Images.Selected()
		if img != nil {
			tag := ""
			if len(img.RepoTags) > 0 {
				tag = img.RepoTags[0]
			}
			confirmAction(m, keys.ShowImageRemove, img.ID, fmt.Sprintf("Remove image %s?", tag))
		}
	case state.PanelVolumes:
		vol := m.Resources.Volumes.Selected()
		if vol != nil {
			confirmAction(m, keys.ShowVolumeRemove, vol.Name, fmt.Sprintf("Remove volume %s?", vol.Name))
		}
	case state.PanelNetworks:
		net := m.Resources.Networks.Selected()
		if net != nil {
			confirmAction(m, keys.ShowNetworkRemove, net.ID, fmt.Sprintf("Remove network %s?", net.Name))
		}
	}
	return m, nil
}

func doEnterAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Navigation.Mode == state.ModeDetail {
		BackFromDetail(m)
		return m, nil
	}
	if m.Navigation.Mode == state.ModeMark {
		return doToggleMark(m)
	}
	if m.Navigation.ActivePanel == state.PanelContainers && m.Connection.Engine != nil {
		return doLogAction(m)
	}
	if m.Navigation.ActivePanel == state.PanelImages {
		return doImageExpand(m)
	}
	if m.Navigation.ActivePanel == state.PanelVolumes {
		if vol := m.Resources.Volumes.Selected(); vol != nil {
			ToVolumeDetail(m, vol.Name)
		}
		return m, nil
	}
	if m.Navigation.ActivePanel == state.PanelCompose {
		return doComposeEnter(m)
	}
	return m, nil
}

func doDetailAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		return doInspectAction(m)
	case state.PanelImages:
		return doImageDetail(m)
	case state.PanelVolumes:
		return doVolumeInspect(m)
	case state.PanelNetworks:
		return doNetworkInspect(m)
	default:
		return m, nil
	}
}

func showConnectionInfo(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine != nil {
		identity := m.Connection.Engine.Identity()
		ShowToastNow(m, fmt.Sprintf("Connected: %s @ %s", identity.Type, identity.Endpoint))
	} else {
		ShowToastNow(m, "Disconnected — no container engine available")
	}
	return m, nil
}

func ShowToastNow(m *state.AppModel, msg string) {
	publishUIMessage(m, audit.LevelInfo, msg)
	if m != nil && m.Dependencies.Audit != nil {
		return
	}
	m.Feedback.ShowToast(msg, state.NotificationInfo, 30)
}

func ShowToastSuccess(m *state.AppModel, msg string) {
	publishUIMessage(m, audit.LevelInfo, msg)
	if m != nil && m.Dependencies.Audit != nil {
		return
	}
	m.Feedback.ShowToast(msg, state.NotificationSuccess, 20)
}

func ShowToastWarn(m *state.AppModel, msg string) {
	publishUIMessage(m, audit.LevelWarn, msg)
	if m != nil && m.Dependencies.Audit != nil {
		return
	}
	m.Feedback.ShowToast(msg, state.NotificationWarning, 40)
}

// ShowKeyHint sets a hint in the header gap area. Returns a cmd to start the auto-clear timer.
func ShowKeyHint(m *state.AppModel, msg string) tea.Cmd {
	sec := m.Dependencies.Config.UI.HintTimeout
	if sec <= 0 {
		sec = 3
	}
	m.Feedback.SetKeyHint(msg, sec*10)
	return func() tea.Msg { return state.KeyHintTick{} }
}
