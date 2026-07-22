package tui

import (
	"context"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// handleMouseWheel processes mouse wheel events for detail and log views.
func handleMouseWheel(m *state.AppModel, msg tea.MouseWheelMsg) *state.AppModel {
	ev := msg.Mouse()
	switch ev.Button {
	case tea.MouseWheelUp:
		switch m.Navigation.Mode {
		case state.ModeDetail:
			m.Detail.Scroll(-3)
		case state.ModeLogView:
			m.Log.Scroll(-3)
		}
	case tea.MouseWheelDown:
		switch m.Navigation.Mode {
		case state.ModeDetail:
			m.Detail.Scroll(3)
		case state.ModeLogView:
			m.Log.Scroll(3)
		}
	}
	return m
}

func Update(msg tea.Msg, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Viewport.Resize(msg.Width, msg.Height)
		if m.Navigation.Mode == state.ModeExecPassthrough && m.Exec.ExecConn != nil {
			go m.Exec.ExecConn.Resize(context.Background(), runtimeapi.TerminalSize{Height: uint(msg.Height), Width: uint(msg.Width)})
		}
		return m, nil

	case tea.MouseWheelMsg:
		return handleMouseWheel(m, msg), nil

	case tea.KeyPressMsg:
		updatedModel, cmd := keyboard.HandleKeyPress(msg, m)
		if updatedModel.Selection.PendingImagePull != "" && updatedModel.Navigation.Mode == state.ModeNormal && updatedModel.Connection.Docker != nil {
			pullRef, trace := updatedModel.Selection.TakeImagePull()
			if cmd != nil {
				return updatedModel, tea.Batch(cmd, keyboard.ImagePullCmdWithAudit(updatedModel.Connection.Docker, pullRef, trace))
			}
			return updatedModel, keyboard.ImagePullCmdWithAudit(updatedModel.Connection.Docker, pullRef, trace)
		}
		return updatedModel, cmd

	case state.ContainersLoaded:
		return handleContainersLoaded(m, msg)

	case state.ImagesLoaded:
		return handleImagesLoaded(m, msg)

	case state.VolumesLoaded:
		return handleVolumesLoaded(m, msg)

	case state.NetworksLoaded:
		return handleNetworksLoaded(m, msg)

	case state.ContainerActioned:
		return handleContainerActioned(m, msg)
	case state.ContainerBatchActioned:
		return handleContainerBatchActioned(m, msg)
	case state.ContainerProcessesLoaded:
		m.Processes.Apply(msg.ContainerID, msg.Processes.Titles, msg.Processes.Processes, msg.Error)
		return m, nil

	case state.ImageActioned:
		return handleImageActioned(m, msg)

	case state.ImageDetailLoaded:
		return handleImageDetailLoaded(m, msg)

	case state.GenericActioned:
		return handleGenericActioned(m, msg)

	case state.ResourcePruned:
		return handleResourcePruned(m, msg)

	case state.LogBatchReceived:
		return handleLogBatchReceived(m, msg)

	case state.LogStreamError:
		return handleLogStreamError(m, msg)

	case state.LogTick:
		return handleLogTick(m, msg)

	case state.StatsReceived:
		return handleStatsReceived(m, msg)

	case state.StatsTick:
		return handleStatsTick(m, msg)

	case state.DockerConnected:
		return handleDockerConnected(m, msg)

	case state.EventStreamReady:
		return handleEventStreamReady(m, msg)

	case state.EventStreamFailed:
		return handleEventStreamFailed(m, msg.Generation, msg.Error)

	case state.RuntimeEventReceived:
		return handleRuntimeEvent(m, msg)

	case state.EventStreamClosed:
		return handleEventStreamFailed(m, msg.Generation, nil)

	case state.EventFlush:
		return handleEventFlush(m, msg)

	case state.EventReconnect:
		return handleEventReconnect(m, msg)

	case state.EventFallbackTick:
		return handleEventFallbackTick(m, msg)

	case state.ToastTick:
		return handleToastTick(m, msg)

	case state.HostStatsTick:
		return handleHostStatsTick(m, msg)

	case state.RuntimeHealthTick:
		return handleRuntimeHealthTick(m, msg)

	case state.RuntimeHealthResult:
		return handleRuntimeHealthResult(m, msg)

	case state.RuntimeProbeResult:
		return handleRuntimeProbeResult(m, msg)

	case state.EscTimeout:
		return handleEscTimeout(m, msg)

	case state.FilterExitTimeout:
		return handleFilterExitTimeout(m, msg)

	case state.KeyHintTick:
		return handleKeyHintTick(m, msg)

	case state.KeyStrokeTick:
		return handleKeyStrokeTick(m, msg)
	case state.ExecOutput:
		return handleExecOutput(m, msg)
	case state.ExecDone:
		return handleExecDone(m)
	}

	return m, nil
}

func handleExecOutput(m *state.AppModel, msg state.ExecOutput) (*state.AppModel, tea.Cmd) {
	m.Exec.Append(msg.Data)
	if m.Exec.ExecCh == nil {
		return m, nil
	}
	return m, func() tea.Msg {
		data, ok := <-m.Exec.ExecCh
		if !ok {
			return state.ExecDone{}
		}
		return state.ExecOutput{Data: data}
	}
}

func handleExecDone(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	keyboard.FinishAudit(m, m.Exec.ExecAudit, audit.ResultSucceeded, "Exec session finished", audit.Details{Shell: m.Exec.ExecShell})
	m.Exec.Reset()
	m.Navigation.Mode = state.ModeNormal
	return m, nil
}
