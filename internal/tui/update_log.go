package tui

import (
	"time"

	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleLogBatchReceived(m *state.AppModel, msg state.LogBatchReceived) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		m.ErrorMessage = msg.Error.Error()
		m.ErrorCount++
	} else {
		m.LogContent = append(m.LogContent, msg.Lines...)
		if len(m.LogContent) > 2000 {
			m.LogContent = m.LogContent[len(m.LogContent)-2000:]
		}
	}
	// Keep polling logs every 2 seconds while in log view
	if m.Mode == state.ModeLogView && m.LogContainerID == msg.ContainerID {
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return state.LogTick{ContainerID: msg.ContainerID}
		})
	}
	return m, nil
}

func handleLogStreamError(m *state.AppModel, msg state.LogStreamError) (*state.AppModel, tea.Cmd) {
	m.ErrorMessage = msg.Error.Error()
	m.ErrorCount++
	return m, nil
}

func handleLogTick(m *state.AppModel, _ state.LogTick) (*state.AppModel, tea.Cmd) {
	if m.Docker != nil && m.LogContainerID != "" {
		// Fetch recent logs on subsequent ticks (use "10s" since to get new lines)
		return m, keyboard.FetchLogBatch(m.Docker, m.LogContainerID, "10s", "200", true)
	}
	return m, nil
}
