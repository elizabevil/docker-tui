package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/tui/keys"

	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/filter"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// handleAction executes a bound key action from the default key mapping.
// cmds is a pre-allocated slice for batched commands (used by keys.ActionRefresh).
func handleAction(action keys.KeyAction, m *state.AppModel, cmds []tea.Cmd) (*state.AppModel, tea.Cmd) {
	switch action {
	case keys.ActionQuit:
		m.ContainerWait.Stop()
		return m, tea.Quit

	case keys.ActionHelp:
		ToHelp(m)
		return m, nil

	case keys.ActionFilter:
		return m, ToFilter(m)

	case keys.ActionRefresh:
		if m.Connection.Engine != nil {
			m.Resources.Containers.Loading = true
			cmds = FetchAll(m.Connection.Engine)
		}
		cmds = append(cmds, probeAllConnections(m)...)
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
		return doContainerAction(m, string(runtimeapi.ActionStart), containerStartCmd)
	case keys.ActionContainerStop:
		return doContainerAction(m, string(runtimeapi.ActionStop), func(c runtimeapi.Engine, id string) tea.Cmd { return containerStopCmd(c, id) })
	case keys.ActionContainerRestart:
		return doContainerAction(m, string(runtimeapi.ActionRestart), func(c runtimeapi.Engine, id string) tea.Cmd { return containerRestartCmd(c, id) })
	case keys.ActionContainerKill:
		return doContainerAction(m, string(runtimeapi.ActionKill), func(c runtimeapi.Engine, id string) tea.Cmd { return containerKillCmd(c, id) })
	case keys.ActionContainerRemove, keys.ActionDelete:
		return doDeleteAction(m)

	case keys.ActionDetail:
		return doDetailAction(m)

	case keys.ActionContainerLogs:
		return doLogAction(m)

	case keys.ActionContainerStats:
		return doStatsAction(m)
	case keys.ActionContainerPause:
		return doPauseAction(m)

	case keys.ActionContainerExec:
		if m.Navigation.ActivePanel != state.PanelContainers {
			return m, nil
		}
		return doAutoExecAction(m)

	case keys.ActionContainerInspect:
		return doInspectAction(m)

	case keys.ActionImagePull:
		return doImagePull(m)
	case keys.ActionImagePrune:
		return doImagePrune(m)
	case keys.ActionImageRemove:
		return doImageRemove(m)
	case keys.ActionImageTag:
		return openImageWorkflow(m, runtimeapi.ImageTransferTag)
	case keys.ActionImagePush:
		return openImageWorkflow(m, runtimeapi.ImageTransferPush)
	case keys.ActionImageSave:
		return openImageWorkflow(m, runtimeapi.ImageTransferSave)
	case keys.ActionImageLoad:
		return openImageWorkflow(m, runtimeapi.ImageTransferLoad)

	case keys.ActionVolumeRemove:
		return doVolumeRemove(m)
	case keys.ActionVolumeCreate:
		return openResourceCreate(m, runtimeapi.ResourceVolume)
	case keys.ActionVolumePrune:
		return confirmResourcePrune(m, runtimeapi.ResourceVolume)
	case keys.ActionNetworkRemove:
		return doNetworkRemove(m)
	case keys.ActionNetworkCreate:
		return openResourceCreate(m, runtimeapi.ResourceNetwork)
	case keys.ActionNetworkPrune:
		return confirmResourcePrune(m, runtimeapi.ResourceNetwork)
	case keys.ActionSwitchRuntime:
		return openRuntimeSelector(m)

	case keys.ActionRefreshConnections:
		return refreshAllConnections(m)

	case keys.ActionCommand:
		ToCommand(m)
		return m, nil

	case keys.ActionImageHistory:
		return openHistoryPage(m)
	case keys.ActionEvents:
		return openEventsPage(m)
	case keys.ActionContainerDiff:
		return doContainerDiff(m)
	case keys.ActionContainerWait:
		return doContainerWait(m)

	case keys.ActionActionBar:
		return doActionBar(m)
	case keys.ActionContainerRename:
		return openRenameDialog(m)
	case keys.ActionContainerTop:
		return openTopView(m)
	case keys.ActionContainerPort:
		return openPortDetail(m)
	case keys.ActionContainerCopy:
		return openContainerCopyForm(m)
	case keys.ActionContainerUpdate:
		return openContainerUpdateForm(m)
	case keys.ActionContainerExport:
		return openContainerExportForm(m)
	case keys.ActionContainerCommit:
		return openContainerCommitForm(m)
	}

	return m, tea.Batch(cmds...)
}

// handleBackAction handles the Esc/Back action with layered context-aware behavior.
func handleBackAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Navigation.Mode == state.ModeDetail {
		BackFromDetail(m)
		return m, nil
	}
	if m.Navigation.Mode == state.ModeLogView {
		BackFromLogView(m)
		return m, nil
	}
	if m.Navigation.Mode == state.ModeFilter {
		BackFromFilter(m)
		return m, nil
	}
	if m.Navigation.ActivePanel == state.PanelImages && m.Resources.Images.ContainersViewID != "" {
		BackFromImageContainers(m)
		return m, nil
	}
	if m.Navigation.ActivePanel == state.PanelVolumes && m.Resources.Volumes.DetailName != "" {
		BackFromVolumeDetail(m)
		return m, nil
	}
	// If a filter is active in the current panel, single Esc clears it
	// instead of starting the double-Esc-to-exit-app flow.
	ctrl := filter.New(m)
	if ctrl.HasActive() {
		ctrl.Clear()
		return m, nil
	}
	m.Navigation.Mode = state.ModeNormal
	m.Metrics.StatsActive = false
	if m.Navigation.EscPending {
		m.Navigation.EscPending = false
		m.ContainerWait.Stop()
		return m, tea.Quit
	}
	m.Navigation.EscPending = true
	m.Feedback.InfoMessage = "Press Esc again to quit"
	return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return state.EscTimeout{}
	})
}

// probeAllConnections issues a concurrent probe for every known host. The
// returned command batch also includes a 5-second ticker so the runtime
// selector keeps latency/state up to date without manual refresh.
func probeAllConnections(m *state.AppModel) []tea.Cmd {
	if m.Connection.Pool == nil {
		ShowToastWarn(m, "no connection pool")
		return nil
	}
	results := m.Connection.Pool.RefreshAll(2 * time.Second)
	if len(results) == 0 {
		ShowToastWarn(m, "no runtime connections")
		return nil
	}
	m.Connection.Connecting = true
	cmds := make([]tea.Cmd, 0, len(results)+1)
	for i := range results {
		probe := &results[i]
		cmds = append(cmds, func() tea.Msg {
			return state.RuntimeProbeResult{Name: probe.Name, Error: probe.Error}
		})
	}
	cmds = append(cmds, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return state.ConnectionRefreshTick{}
	}))
	return cmds
}

// refreshAllConnections probes all pool connections concurrently.
func refreshAllConnections(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	return m, tea.Batch(probeAllConnections(m)...)
}

func doActionBar(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	m.Navigation.ActionBar.Open()
	m.Navigation.Mode = state.ModeActionBar
	return m, nil
}
