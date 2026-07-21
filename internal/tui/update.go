package tui

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/term"

	tea "charm.land/bubbletea/v2"
)

// handleMouseWheel processes mouse wheel events for detail and log views.
func handleMouseWheel(m *state.AppModel, msg tea.MouseWheelMsg) *state.AppModel {
	ev := msg.Mouse()
	switch ev.Button {
	case tea.MouseWheelUp:
		switch m.Mode {
		case state.ModeDetail:
			if m.DetailOffset > 0 {
				m.DetailOffset -= 3
				if m.DetailOffset < 0 {
					m.DetailOffset = 0
				}
			}
		case state.ModeLogView:
			if m.LogViewOffset > 0 {
				m.LogViewOffset -= 3
				if m.LogViewOffset < 0 {
					m.LogViewOffset = 0
				}
			}
		}
	case tea.MouseWheelDown:
		switch m.Mode {
		case state.ModeDetail:
			m.DetailOffset += 3
		case state.ModeLogView:
			m.LogViewOffset += 3
		}
	}
	return m
}

func Update(msg tea.Msg, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		if m.Mode == state.ModeExecPassthrough && m.ExecID != "" && m.Docker != nil {
			cli := m.Docker.Raw()
			go cli.ContainerExecResize(context.Background(), m.ExecID, container.ResizeOptions{
				Height: uint(msg.Height),
				Width:  uint(msg.Width),
			})
		}
		return m, nil

	case tea.MouseWheelMsg:
		return handleMouseWheel(m, msg), nil

	case tea.KeyPressMsg:
		updatedModel, cmd := keyboard.HandleKeyPress(msg, m)
		if updatedModel.PendingImagePull != "" && updatedModel.Mode == state.ModeNormal && updatedModel.Docker != nil {
			pullRef := updatedModel.PendingImagePull
			updatedModel.PendingImagePull = ""
			if cmd != nil {
				return updatedModel, tea.Batch(cmd, keyboard.ImagePullCmd(updatedModel.Docker, pullRef))
			}
			return updatedModel, keyboard.ImagePullCmd(updatedModel.Docker, pullRef)
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

	case state.ImageActioned:
		return handleImageActioned(m, msg)

	case state.ImageDetailLoaded:
		return handleImageDetailLoaded(m, msg)

	case state.GenericActioned:
		return handleGenericActioned(m, msg)

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

	case state.ContainerEvent:
		return handleContainerEvent(m, msg)

	case state.ToastTick:
		return handleToastTick(m, msg)

	case state.HostStatsTick:
		return handleHostStatsTick(m, msg)

	case state.EscTimeout:
		return handleEscTimeout(m, msg)

	case state.FilterExitTimeout:
		return handleFilterExitTimeout(m, msg)

	case state.KeyHintTick:
		return handleKeyHintTick(m, msg)

	case state.KeyStrokeTick:
		return handleKeyStrokeTick(m, msg)
	case state.SpinnerTick:
		if m.Spinner != nil && m.Spinner.Active() {
			m.Spinner.Tick()
		}
		return m, nil

	case state.ExecOutput:
		return handleExecOutput(m, msg)
	case state.ExecDone:
		return handleExecDone(m)
	}

	return m, nil
}

func handleExecOutput(m *state.AppModel, msg state.ExecOutput) (*state.AppModel, tea.Cmd) {
	if m.ExecBuf == nil {
		m.ExecBuf = term.NewBuffer(2000)
	}
	m.ExecBuf.Write(msg.Data)
	if m.ExecCh == nil {
		return m, nil
	}
	return m, func() tea.Msg {
		data, ok := <-m.ExecCh
		if !ok {
			return state.ExecDone{}
		}
		return state.ExecOutput{Data: data}
	}
}

func handleExecDone(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	m.ExecConn = nil
	m.ExecID = ""
	m.ExecCh = nil
	m.ExecDone = nil
	if m.ExecBuf != nil {
		m.ExecBuf.Reset()
	}
	m.ExecScroll = 0
	m.Mode = state.ModeNormal
	return m, nil
}
