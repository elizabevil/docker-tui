package update

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	view "github.com/elizabevil/docker-tui/internal/tui/ui/app"

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
		case state.ModeEvents:
			m.EventPanel.MoveCursor(-3, len(m.EventPanel.FilteredEvents()), max(1, m.Viewport.Height-18))
		}
	case tea.MouseWheelDown:
		switch m.Navigation.Mode {
		case state.ModeDetail:
			m.Detail.Scroll(3)
		case state.ModeLogView:
			m.Log.Scroll(3)
		case state.ModeEvents:
			m.EventPanel.MoveCursor(3, len(m.EventPanel.FilteredEvents()), max(1, m.Viewport.Height-18))
		}
	}
	return m
}

// handleMouseClick translates a mouse click into a cursor move or a
// scroll step. It delegates to the view package so the layout math
// stays in one place; keyboard input remains the primary control
// surface and mouse is purely additive.
func handleMouseClick(m *state.AppModel, msg tea.MouseClickMsg) (*state.AppModel, tea.Cmd) {
	view.ApplyMouseClick(m, msg)
	return m, nil
}

func Update(msg tea.Msg, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Viewport.Resize(msg.Width, msg.Height)
		if m.Navigation.Mode == state.ModeExecPassthrough && m.Exec.ExecConn != nil {
			go func() {
				_ = m.Exec.ExecConn.Resize(context.Background(), runtimeapi.TerminalSize{Height: uint(msg.Height), Width: uint(msg.Width)})
			}() //nolint:errcheck // fire-and-forget resize on window change.
		}
		return m, nil

	case tea.MouseWheelMsg:
		return handleMouseWheel(m, msg), nil

	case tea.MouseClickMsg:
		return handleMouseClick(m, msg)

	case tea.KeyPressMsg:
		updatedModel, cmd := keyboard.HandleKeyPress(msg, m)
		if updatedModel.Selection.PendingImagePull != "" && updatedModel.Navigation.Mode == state.ModeNormal && updatedModel.Connection.Engine != nil {
			pullRef, trace := updatedModel.Selection.TakeImagePull()
			if cmd != nil {
				return updatedModel, tea.Batch(cmd, keyboard.ImagePullCmdWithAudit(updatedModel.Connection.Engine, pullRef, trace))
			}
			return updatedModel, keyboard.ImagePullCmdWithAudit(updatedModel.Connection.Engine, pullRef, trace)
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
	case state.BatchActioned:
		return handleBatchActioned(m, msg)
	case state.ContainerProcessesLoaded:
		m.Processes.Apply(msg.ContainerID, msg.Processes.Titles, msg.Processes.Processes, msg.Error)
		return m, nil
	case keyboard.ContainerDiffDone:
		return handleContainerDiffDone(m, msg)
	case keyboard.ContainerWaitDone:
		return handleContainerWaitDone(m, msg)

	case state.ImageActioned:
		return handleImageActioned(m, msg)

	case state.ImageDetailLoaded:
		return handleImageDetailLoaded(m, msg)
	case state.ContainerDetailLoaded:
		return handleContainerDetailLoaded(m, msg)
	case state.VolumeDetailLoaded:
		return handleVolumeDetailLoaded(m, msg)
	case state.NetworkDetailLoaded:
		return handleNetworkDetailLoaded(m, msg)
	case state.HistoryLoadedMsg:
		return handleHistoryLoaded(m, msg)
	case state.ImageTransferReceived:
		return handleImageTransferReceived(m, msg)

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

	case state.ConnectionRefreshTick:
		return handleConnectionRefreshTick(m, msg)

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

func handleContainerDiffDone(m *state.AppModel, msg keyboard.ContainerDiffDone) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		keyboard.FinishAudit(m, msg.Audit, audit.ResultFailed, "Container diff failed", audit.Details{Error: msg.Error.Error()})
		m.Feedback.RecordError("resource.container.diff: " + msg.Error.Error())
		return m, nil
	}
	var content strings.Builder
	content.WriteString("KIND  PATH\n")
	for _, change := range msg.Changes {
		content.WriteString(fmt.Sprintf("%-5s %s\n", containerDiffKind(change.Kind), change.Path))
	}
	if len(msg.Changes) == 0 {
		content.WriteString("No filesystem changes.\n")
	}
	keyboard.FinishAudit(m, msg.Audit, audit.ResultSucceeded, "Container diff loaded", audit.Details{})
	keyboard.ToDetail(m, "Container Diff: "+shortMessageID(msg.ContainerID), content.String())
	return m, nil
}

func handleContainerWaitDone(m *state.AppModel, msg keyboard.ContainerWaitDone) (*state.AppModel, tea.Cmd) {
	// A response from an older wait must not affect a newer operation. A
	// response for the current generation is still terminal after Stop(), and
	// must close its audit record instead of being discarded as stale.
	if msg.Generation == 0 || msg.Generation != m.ContainerWait.Generation {
		return m, nil
	}
	m.ContainerWait.Stop()
	if errors.Is(msg.Error, context.Canceled) {
		keyboard.FinishAudit(m, msg.Audit, audit.ResultCancelled, "Container wait cancelled", audit.Details{})
		return m, nil
	}
	if msg.Error != nil {
		keyboard.FinishAudit(m, msg.Audit, audit.ResultFailed, "Container wait failed", audit.Details{Error: msg.Error.Error()})
		m.Feedback.RecordError("resource.container.wait: " + msg.Error.Error())
		return m, nil
	}
	message := fmt.Sprintf("Container exited with status %d", msg.Result.StatusCode)
	keyboard.FinishAudit(m, msg.Audit, audit.ResultSucceeded, message, audit.Details{})
	m.Feedback.ShowToast(message, state.NotificationSuccess, 30)
	return m, nil
}

func containerDiffKind(kind runtimeapi.ChangeKind) string {
	switch kind {
	case runtimeapi.ChangeAdded:
		return "ADD"
	case runtimeapi.ChangeDeleted:
		return "DEL"
	default:
		return "MOD"
	}
}

func shortMessageID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
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
