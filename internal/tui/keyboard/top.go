package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleTopKey(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch key {
	case keys.KeyEsc:
		m.Processes.Close()
		m.Navigation.Mode = state.ModeNormal
	case keys.KeyUp, keys.KeyK:
		m.Processes.Move(-1)
	case keys.KeyDown, keys.KeyJ:
		m.Processes.Move(1)
	case keys.KeyR:
		if m.Connection.Docker != nil && m.Processes.ContainerID != "" {
			m.Processes.Loading = true
			return m, fetchContainerProcesses(m.Connection.Docker.Containers(), m.Processes.ContainerID)
		}
	}
	return m, nil
}
