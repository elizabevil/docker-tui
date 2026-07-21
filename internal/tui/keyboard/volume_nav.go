package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// handleVolumeEnter handles Enter key on the volumes panel.
// Returns nil, nil if not handled.
func handleVolumeEnter(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if key != keys.KeyEnter || m.Navigation.ActivePanel != state.PanelVolumes {
		return nil, nil
	}
	vol := m.Resources.Volumes.Selected()
	if vol != nil {
		ToVolumeDetail(m, vol.Name)
	}
	return m, nil
}
