package keyboard

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func volumeRemoveCmd(client *docker.Client, name string, force bool) tea.Cmd {
	return func() tea.Msg {
		err := client.RemoveVolume(name, force)
		return state.GenericActioned{Action: state.ActionRemoved, ID: name, Success: err == nil, Error: err}
	}
}

func doVolumeRemove(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil || m.ActivePanel != state.PanelVolumes {
		return m, nil
	}
	vol := m.Volumes.Selected()
	if vol == nil {
		return m, nil
	}
	confirmAction(m, "volume-remove", vol.Name, fmt.Sprintf("Remove volume %s?", vol.Name))
	m.ConfirmAudit = beginAudit(m, "resource.volume.delete", audit.VolumeTarget{Name: vol.Name, Meta: audit.VolumeMeta{Driver: vol.Driver}}, "Remove volume "+vol.Name)
	return m, nil
}
