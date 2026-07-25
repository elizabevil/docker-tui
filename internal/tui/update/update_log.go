package update

import (
	"time"

	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleLogBatchReceived(m *state.AppModel, msg state.LogBatchReceived) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		m.Feedback.RecordError(msg.Error.Error())
	} else {
		m.Log.Append(msg.Lines)
	}
	// Keep polling logs every 2 seconds while in log view
	if m.Navigation.Mode == state.ModeLogView && m.Log.LogContainerID == msg.ContainerID {
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return state.LogTick{ContainerID: msg.ContainerID}
		})
	}
	return m, nil
}

func handleLogStreamError(m *state.AppModel, msg state.LogStreamError) (*state.AppModel, tea.Cmd) {
	m.Feedback.RecordError(msg.Error.Error())
	return m, nil
}

func handleLogTick(m *state.AppModel, _ state.LogTick) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine != nil && m.Log.LogContainerID != "" {
		// Fetch recent logs on subsequent ticks (use "10s" since to get new lines)
		return m, keyboard.FetchLogBatch(m.Connection.Engine, m.Log.LogContainerID, "10s", "200", true)
	}
	return m, nil
}
