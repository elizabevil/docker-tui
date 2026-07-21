package keyboard

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"

	tea "charm.land/bubbletea/v2"
)

func doDeleteAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
		return m, nil
	}
	if len(m.MarkedIDs) > 0 {
		return doBulkDelete(m)
	}
	switch m.ActivePanel {
	case state.PanelContainers:
		ctr := m.Containers.Selected()
		if ctr != nil {
			confirmAction(m, "container-remove", ctr.ID, fmt.Sprintf("Remove container %s?", ctr.Name))
		}
	case state.PanelImages:
		img := m.Images.Selected()
		if img != nil {
			tag := ""
			if len(img.RepoTags) > 0 {
				tag = img.RepoTags[0]
			}
			confirmAction(m, "image-remove", img.ID, fmt.Sprintf("Remove image %s?", tag))
		}
	case state.PanelVolumes:
		vol := m.Volumes.Selected()
		if vol != nil {
			confirmAction(m, "volume-remove", vol.Name, fmt.Sprintf("Remove volume %s?", vol.Name))
		}
	case state.PanelNetworks:
		net := m.Networks.Selected()
		if net != nil {
			confirmAction(m, "network-remove", net.ID, fmt.Sprintf("Remove network %s?", net.Name))
		}
	}
	return m, nil
}

func doEnterAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Mode == state.ModeDetail {
		BackFromDetail(m)
		return m, nil
	}
	if m.Mode == state.ModeMark {
		return doToggleMark(m)
	}
	if m.ActivePanel == state.PanelContainers && m.Docker != nil {
		return doLogAction(m)
	}
	if m.ActivePanel == state.PanelImages && m.Docker != nil {
		return doImageExpand(m)
	}
	if m.ActivePanel == state.PanelVolumes {
		if vol := m.Volumes.Selected(); vol != nil {
			ToVolumeDetail(m, vol.Name)
		}
		return m, nil
	}
	if m.ActivePanel == state.PanelCompose {
		return doComposeEnter(m)
	}
	return m, nil
}

func doDetailAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch m.ActivePanel {
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
	if m.Docker != nil {
		ShowToastNow(m, fmt.Sprintf("Connected: %s @ %s", m.Docker.RuntimeType, m.Docker.Host))
	} else {
		ShowToastNow(m, "Disconnected — no container engine available")
	}
	return m, nil
}

func ShowToastNow(m *state.AppModel, msg string) {
	publishUIMessage(m, audit.LevelInfo, msg)
	if m != nil && m.Audit != nil {
		return
	}
	m.ToastMessage = msg
	m.ToastLevel = component.ToastInfo
	m.ToastTimer = 30
}

func ShowToastSuccess(m *state.AppModel, msg string) {
	publishUIMessage(m, audit.LevelInfo, msg)
	if m != nil && m.Audit != nil {
		return
	}
	m.ToastMessage = msg
	m.ToastLevel = component.ToastSuccess
	m.ToastTimer = 20
}

func ShowToastWarn(m *state.AppModel, msg string) {
	publishUIMessage(m, audit.LevelWarn, msg)
	if m != nil && m.Audit != nil {
		return
	}
	m.ToastMessage = msg
	m.ToastLevel = component.ToastWarning
	m.ToastTimer = 40
}

// ShowKeyHint sets a hint in the header gap area. Returns a cmd to start the auto-clear timer.
func ShowKeyHint(m *state.AppModel, msg string) tea.Cmd {
	m.KeyHint = msg
	sec := m.Config.UI.HintTimeout
	if sec <= 0 {
		sec = 3
	}
	m.KeyHintTimer = sec * 10 // 10 ticks per second (100ms each)
	return func() tea.Msg { return state.KeyHintTick{} }
}
