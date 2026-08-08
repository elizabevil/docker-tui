package update

import (
	"fmt"
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

// handleComposeLogBatchReceived appends aggregated multi-source lines
// (R08-06). Per-source stream errors are reported but do not abort
// the view; the user keeps reading whatever succeeded.
func handleComposeLogBatchReceived(m *state.AppModel, msg state.ComposeLogBatchReceived) (*state.AppModel, tea.Cmd) {
	m.Log.Append(msg.Lines)
	for _, e := range msg.StreamErrs {
		m.Feedback.RecordError(fmt.Sprintf("compose log stream %s: %v", e.ContainerID, e.Error))
	}
	return m, nil
}

// handleComposeServiceTopLoaded / PortLoaded / StatsLoaded stash the
// per-service query results on the Compose state so a UI subview can
// render them. R08-07 leaves the actual subviews to a UI overhaul.
func handleComposeServiceTopLoaded(m *state.AppModel, msg state.ComposeServiceTopLoaded) (*state.AppModel, tea.Cmd) {
	m.Compose.ComposeServiceTop = msg.Items
	if msg.Error != nil {
		m.Feedback.RecordError(msg.Error.Error())
	}
	return m, nil
}

func handleComposeServicePortLoaded(m *state.AppModel, msg state.ComposeServicePortLoaded) (*state.AppModel, tea.Cmd) {
	m.Compose.ComposeServicePort = msg.Items
	if msg.Error != nil {
		m.Feedback.RecordError(msg.Error.Error())
	}
	return m, nil
}

func handleComposeServiceStatsLoaded(m *state.AppModel, msg state.ComposeServiceStatsLoaded) (*state.AppModel, tea.Cmd) {
	m.Compose.ComposeServiceStats = msg.Items
	if msg.Error != nil {
		m.Feedback.RecordError(msg.Error.Error())
	}
	return m, nil
}

// handleComposeServiceRunCompleted surfaces a `docker compose run`
// failure (R08-08 F2). Success is silent — the engine either holds
// the exec session (attached) or leaves the oneoff container running
// (detached); both surfaces the running container in the panel.
func handleComposeServiceRunCompleted(m *state.AppModel, msg state.ComposeServiceRunCompleted) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		m.Feedback.RecordError(fmt.Sprintf("compose run %s/%s: %v", msg.Project, msg.Service, msg.Error))
	}
	return m, nil
}

// handleComposeServicePushCompleted surfaces the aggregated push
// result (R08-05): a toast reports the success count, errors land in
// the feedback panel.
func handleComposeServicePushCompleted(m *state.AppModel, msg state.ComposeServicePushCompleted) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		m.Feedback.RecordError(fmt.Sprintf("compose push %s: %v", msg.Project, msg.Error))
		return m, nil
	}
	if msg.Success > 0 {
		keyboard.ShowToastNow(m, fmt.Sprintf("✓ compose push %s: %d images", msg.Project, msg.Success))
	}
	return m, nil
}

// handleComposeServicePullCompleted surfaces the aggregated pull
// result (R08-05): a toast reports the success count, errors land in
// the feedback panel.
func handleComposeServicePullCompleted(m *state.AppModel, msg state.ComposeServicePullCompleted) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		m.Feedback.RecordError(fmt.Sprintf("compose pull %s: %v", msg.Project, msg.Error))
		return m, nil
	}
	if msg.Success > 0 {
		keyboard.ShowToastNow(m, fmt.Sprintf("✓ compose pull %s: %d images", msg.Project, msg.Success))
	}
	return m, nil
}

func handleLogTick(m *state.AppModel, _ state.LogTick) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine != nil && m.Log.LogContainerID != "" {
		// Fetch recent logs on subsequent ticks (use "10s" since to get new lines)
		return m, keyboard.FetchLogBatch(m.Connection.Engine, m.Log.LogContainerID, "10s", "200", true)
	}
	return m, nil
}
