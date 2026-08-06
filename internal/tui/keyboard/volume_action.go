package keyboard

import (
	"context"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// volumeRemoveCmd returns a tea.Cmd that removes a volume.
func volumeRemoveCmd(client runtimeapi.Engine, name string, force bool) tea.Cmd {
	return func() tea.Msg {
		err := client.Volumes().Remove(context.Background(), name, force)
		return state.GenericActioned{Action: state.ActionRemoved, ID: name, Success: err == nil, Error: err}
	}
}

// doVolumeInspect opens the detail view for the selected volume.
func doVolumeInspect(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelVolumes {
		return m, nil
	}
	vol := m.Resources.Volumes.Selected()
	if vol == nil {
		return m, nil
	}
	title := i18n.T("detail.title.volume", vol.Name)
	ToDetail(m, title, "")
	return m, volumeInspectCmd(m.Connection.Engine, vol.Name, title)
}

func volumeInspectCmd(client runtimeapi.Engine, name, title string) tea.Cmd {
	return func() tea.Msg {
		detail, err := client.Volumes().Inspect(context.Background(), name)
		return state.VolumeDetailLoaded{
			VolumeID: name,
			Title:    title,
			Detail:   detail,
			Error:    err,
		}
	}
}

// doVolumeRemove opens the parameter form for removing the selected volume.
// Dangerous=true forces the Cancel focus on open so a stray Enter cannot
// drop the volume before the user opts in.
func doVolumeRemove(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelVolumes {
		return m, nil
	}
	vol := m.Resources.Volumes.Selected()
	if vol == nil {
		return m, nil
	}
	return openVolumeRemoveForm(m, vol)
}
