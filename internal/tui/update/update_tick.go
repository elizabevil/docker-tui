package update

import (
	"context"
	"fmt"
	"time"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleStatsReceived(m *state.AppModel, msg state.StatsReceived) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		return m, nil
	}
	m.Resources.Containers.Stats[msg.ContainerID] = state.ContainerStats{
		CPU:      msg.CPU,
		MemUsage: msg.MemUsage,
		MemLimit: msg.MemLimit,
		MemPerc:  msg.MemPerc,
		NetRx:    msg.NetRx,
		NetTx:    msg.NetTx,
	}
	return m, nil
}

func handleStatsTick(m *state.AppModel, _ state.StatsTick) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		m.Metrics.StatsActive = false
		return m, nil
	}
	// Auto-enable stats when containers panel is active
	if !m.Metrics.StatsActive && m.Navigation.ActivePanel == state.PanelContainers {
		m.Metrics.StatsActive = true
	}
	if !m.Metrics.StatsActive {
		return m, nil
	}

	// Fetch stats for all currently visible (filtered) containers
	items := m.Resources.Containers.SortedItems()
	cmds := make([]tea.Cmd, 0, len(items))
	for _, c := range items {
		cmds = append(cmds, keyboard.FetchStats(m.Connection.Engine.Containers(), c.ID))
	}

	// Schedule next tick
	pollSec := 3
	if m.Dependencies.Config != nil && m.Dependencies.Config.Docker.StatsPollSec > 0 {
		pollSec = m.Dependencies.Config.Docker.StatsPollSec
	}
	cmds = append(cmds, tea.Tick(time.Duration(pollSec)*time.Second, func(t time.Time) tea.Msg {
		return state.StatsTick{}
	}))

	return m, tea.Batch(cmds...)
}

func handleDockerConnected(m *state.AppModel, msg state.DockerConnected) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		failureMessage := i18n.ConnectionFailureMessage(string(runtime.ClassifyConnectionError(msg.Error).Kind))
		if m.Navigation.Mode == state.ModeRuntimeSelect {
			m.Connection.SelectionFailed(msg.Name, msg.Error)
			m.Feedback.RecordError(failureMessage)
			keyboard.ShowToastWarn(m, fmt.Sprintf("Connection failed (%s): %s", msg.Name, failureMessage))
			return m, nil
		}
		m.Events.Stop()
		m.Connection.Failed(msg.Name, msg.Error)
		m.Feedback.RecordError(failureMessage)
		keyboard.ShowToastWarn(m, fmt.Sprintf("Connection failed (%s): %s", msg.Name, failureMessage))
		return m, nil
	}
	m.Connection.ConnectedTo(msg.Name, msg.Engine)
	m.Navigation.Mode = state.ModeNormal
	m.Feedback.ClearError()
	if msg.Name != "" {
		m.Connection.RuntimeType = msg.Name
	}
	if msg.Notice != "" {
		keyboard.ShowToastWarn(m, msg.Notice)
	}
	m.Resources.Containers.Loading = true
	commands := keyboard.FetchAll(m.Connection.Engine)
	commands = append(commands, startEventStream(m))
	return m, tea.Batch(commands...)
}

func handleToastTick(m *state.AppModel, msg state.ToastTick) (*state.AppModel, tea.Cmd) {
	if msg.Generation != m.Feedback.ToastGeneration || m.Feedback.ToastTimer <= 0 {
		return m, nil
	}
	m.Feedback.TickToast()
	if m.Feedback.ToastTimer <= 0 {
		return m, nil
	}
	return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return state.ToastTick{Generation: msg.Generation}
	})
}

func handleHostStatsTick(m *state.AppModel, _ state.HostStatsTick) (*state.AppModel, tea.Cmd) {
	stats := runtime.ReadHostStats()
	m.Metrics.ApplyHost(stats)
	return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return state.HostStatsTick{}
	})
}

func handleRuntimeHealthTick(m *state.AppModel, _ state.RuntimeHealthTick) (*state.AppModel, tea.Cmd) {
	health := config.DefaultConfig().Runtime.Health
	if m.Dependencies.Config != nil {
		health = m.Dependencies.Config.Runtime.Health
	}
	cmds := []tea.Cmd{tea.Tick(health.Interval(), func(time.Time) tea.Msg {
		return state.RuntimeHealthTick{}
	})}
	if m.Connection.Engine != nil {
		engine, name := m.Connection.Engine, m.Connection.ConnectionTarget
		cmds = append(cmds, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), health.Timeout())
			defer cancel()
			return state.RuntimeHealthResult{Name: name, Error: engine.PingContext(ctx)}
		})
	}
	return m, tea.Batch(cmds...)
}

func handleRuntimeHealthResult(m *state.AppModel, msg state.RuntimeHealthResult) (*state.AppModel, tea.Cmd) {
	threshold := config.DefaultConfig().Runtime.Health.FailureThreshold
	if m.Dependencies.Config != nil {
		threshold = m.Dependencies.Config.Runtime.Health.FailureThreshold
	}
	transition := m.Connection.ApplyHealthResult(msg.Name, msg.Error, threshold)
	switch transition {
	case state.HealthDisconnected:
		m.Feedback.RecordError(fmt.Sprintf("runtime health check failed: %v", msg.Error))
		keyboard.ShowToastWarn(m, fmt.Sprintf("Runtime %s disconnected: %v", m.Connection.ConnectionTarget, msg.Error))
	case state.HealthRecovered:
		keyboard.ShowToastNow(m, fmt.Sprintf("Runtime %s recovered", m.Connection.ConnectionTarget))
	}
	return m, nil
}

func handleRuntimeProbeResult(m *state.AppModel, msg state.RuntimeProbeResult) (*state.AppModel, tea.Cmd) {
	if m.Navigation.Mode == state.ModeRuntimeSelect {
		m.Connection.SetProbeResult(msg.Name, msg.Error)
		return m, nil
	}
	// Show toast for probe results outside the selector.
	if msg.Error != nil {
		keyboard.ShowToastWarn(m, fmt.Sprintf("%s: %v", msg.Name, msg.Error))
	}
	return m, nil
}

func handleConnectionRefreshTick(m *state.AppModel, _ state.ConnectionRefreshTick) (*state.AppModel, tea.Cmd) {
	if m.Connection.Pool == nil {
		return m, nil
	}
	results := m.Connection.Pool.RefreshAll(2 * time.Second)
	if len(results) == 0 {
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return state.ConnectionRefreshTick{}
		})
	}
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
	return m, tea.Batch(cmds...)
}

func handleEscTimeout(m *state.AppModel, _ state.EscTimeout) (*state.AppModel, tea.Cmd) {
	m.Navigation.EscPending = false
	m.Feedback.InfoMessage = ""
	return m, nil
}

func handleKeyHintTick(m *state.AppModel, _ state.KeyHintTick) (*state.AppModel, tea.Cmd) {
	if m.Feedback.KeyHintTimer > 0 {
		m.Feedback.KeyHintTimer--
		if m.Feedback.KeyHintTimer <= 0 {
			m.Feedback.KeyHint = ""
		} else {
			return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
				return state.KeyHintTick{}
			})
		}
	}
	return m, nil
}

func handleKeyStrokeTick(m *state.AppModel, _ state.KeyStrokeTick) (*state.AppModel, tea.Cmd) {
	// 倒计时：按键显示 3s 后销毁，不响应过期 Tick
	if m.Feedback.KeyStrokeTimer <= 0 {
		return m, nil
	}
	m.Feedback.KeyStrokeTimer--
	if m.Feedback.KeyStrokeTimer <= 0 {
		m.Feedback.KeyStrokeBuffer = nil
		m.Feedback.LastKeyStroke = nil
		return m, nil
	}
	return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return state.KeyStrokeTick{}
	})
}

func handleFilterExitTimeout(m *state.AppModel, msg state.FilterExitTimeout) (*state.AppModel, tea.Cmd) {
	if msg.Token == m.Navigation.FilterExitToken {
		m.Navigation.ClearFilterExit()
	}
	return m, nil
}
