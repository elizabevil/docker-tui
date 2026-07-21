package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"

	"time"

	"github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// handleAction executes a bound key action from the default key mapping.
// cmds is a pre-allocated slice for batched commands (used by keys.ActionRefresh).
func handleAction(action keys.KeyAction, m *state.AppModel, cmds []tea.Cmd) (*state.AppModel, tea.Cmd) {
	switch action {
	case keys.ActionQuit:
		return m, tea.Quit

	case keys.ActionHelp:
		ToHelp(m)
		return m, nil

	case keys.ActionFilter:
		return m, ToFilter(m)

	case keys.ActionRefresh:
		if m.Docker != nil {
			m.Containers.Loading = true
			cmds = FetchAll(m.Docker)
		}
		return m, tea.Batch(cmds...)

	case keys.ActionTabNext:
		switchPanel(m, 1)
		return m, nil
	case keys.ActionTabPrev:
		switchPanel(m, -1)
		return m, nil

	case keys.ActionUp:
		moveCursor(m, -1)
		return m, nil
	case keys.ActionDown:
		moveCursor(m, 1)
		return m, nil

	case keys.ActionEnter:
		return doEnterAction(m)

	case keys.ActionBack:
		return handleBackAction(m)

	case keys.ActionContainerStart:
		return doContainerAction(m, "start", containerStartCmd)
	case keys.ActionContainerStop:
		return doContainerAction(m, "stop", func(c *docker.Client, id string) tea.Cmd { return containerStopCmd(c, id) })
	case keys.ActionContainerRestart:
		return doContainerAction(m, "restart", func(c *docker.Client, id string) tea.Cmd { return containerRestartCmd(c, id) })
	case keys.ActionContainerKill:
		return doContainerAction(m, "kill", func(c *docker.Client, id string) tea.Cmd { return containerKillCmd(c, id) })
	case keys.ActionContainerRemove, keys.ActionDelete:
		return doDeleteAction(m)

	case keys.ActionDetail:
		return doDetailAction(m)

	case keys.ActionContainerLogs:
		return doLogAction(m)

	case keys.ActionContainerStats:
		return doStatsAction(m)

	case keys.ActionContainerExec:
		if m.ActivePanel != state.PanelContainers {
			return m, nil
		}
		ToExec(m)
		return m, nil

	case keys.ActionContainerInspect:
		return doInspectAction(m)

	case keys.ActionImagePull:
		return doImagePull(m)
	case keys.ActionImagePrune:
		return doImagePrune(m)
	case keys.ActionImageRemove:
		return doImageRemove(m)

	case keys.ActionVolumeRemove:
		return doVolumeRemove(m)
	case keys.ActionNetworkRemove:
		return doNetworkRemove(m)
	case keys.ActionSwitchRuntime:
		return openRuntimeSelector(m)

	case keys.ActionCommand:
		ToCommand(m)
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// handleBackAction handles the Esc/Back action with layered context-aware behavior.
func handleBackAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Mode == state.ModeDetail {
		BackFromDetail(m)
		return m, nil
	}
	if m.Mode == state.ModeLogView {
		BackFromLogView(m)
		return m, nil
	}
	if m.Mode == state.ModeFilter {
		BackFromFilter(m)
		return m, nil
	}
	if m.ActivePanel == state.PanelImages && m.Images.ContainersViewID != "" {
		BackFromImageContainers(m)
		return m, nil
	}
	if m.ActivePanel == state.PanelVolumes && m.Volumes.DetailName != "" {
		BackFromVolumeDetail(m)
		return m, nil
	}
	m.Mode = state.ModeNormal
	m.StatsActive = false
	if m.EscPending {
		m.EscPending = false
		return m, tea.Quit
	}
	m.EscPending = true
	m.InfoMessage = "Press Esc again to quit"
	return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return state.EscTimeout{}
	})
}
