package tui

import (
	"fmt"
	"time"

	"github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleStatsReceived(m *state.AppModel, msg state.StatsReceived) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		return m, nil
	}
	m.Containers.Stats[msg.ContainerID] = state.ContainerStats{
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
	if m.Docker == nil {
		m.StatsActive = false
		return m, nil
	}
	// Auto-enable stats when containers panel is active
	if !m.StatsActive && m.ActivePanel == state.PanelContainers {
		m.StatsActive = true
	}
	if !m.StatsActive {
		return m, nil
	}

	// Fetch stats for all currently visible (filtered) containers
	items := m.Containers.SortedItems()
	cmds := make([]tea.Cmd, 0, len(items))
	for _, c := range items {
		cmds = append(cmds, keyboard.FetchStats(m.Docker, c.ID))
	}

	// Schedule next tick
	pollSec := 3
	if m.Config != nil && m.Config.Docker.StatsPollSec > 0 {
		pollSec = m.Config.Docker.StatsPollSec
	}
	cmds = append(cmds, tea.Tick(time.Duration(pollSec)*time.Second, func(t time.Time) tea.Msg {
		return state.StatsTick{}
	}))

	return m, tea.Batch(cmds...)
}

func handleDockerConnected(m *state.AppModel, msg state.DockerConnected) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		if m.Mode == state.ModeRuntimeSelect {
			m.Connecting = false
			m.ErrorMessage = msg.Error.Error()
			m.ErrorCount++
			if m.RuntimeSelectorError == nil {
				m.RuntimeSelectorError = make(map[string]string)
			}
			m.RuntimeSelectorError[msg.Name] = msg.Error.Error()
			keyboard.ShowToastWarn(m, fmt.Sprintf("Connection failed (%s): %s", msg.Name, msg.Error))
			return m, nil
		}
		m.Connecting = false
		m.Connected = false
		m.Docker = nil
		m.ConnectionTarget = msg.Name
		m.ConnectionError = msg.Error.Error()
		m.RuntimeType = ""
		m.EngineVersion = ""
		m.ErrorMessage = msg.Error.Error()
		m.ErrorCount++
		m.HealthFailures = 0
		m.HealthDegraded = false
		keyboard.ShowToastWarn(m, fmt.Sprintf("Connection failed (%s): %s", msg.Name, msg.Error))
		return m, nil
	}
	m.Docker = msg.Client
	m.Connecting = false
	m.Connected = true
	m.ConnectionTarget = msg.Name
	m.ConnectionError = ""
	m.HealthFailures = 0
	m.HealthDegraded = false
	m.Mode = state.ModeNormal
	if m.RuntimeSelectorError != nil {
		delete(m.RuntimeSelectorError, msg.Name)
	}
	m.ErrorMessage = ""
	m.RuntimeType = string(msg.Client.RuntimeType)
	m.EngineVersion = msg.Client.EngineVersion
	if msg.Name != "" {
		m.RuntimeType = msg.Name
	}
	if msg.Notice != "" {
		keyboard.ShowToastWarn(m, msg.Notice)
	}
	m.Containers.Loading = true
	return m, tea.Batch(keyboard.FetchAll(m.Docker)...)
}

func handleContainerEvent(m *state.AppModel, msg state.ContainerEvent) (*state.AppModel, tea.Cmd) {
	if m.Docker != nil {
		switch msg.Action {
		case "start", "stop", "die", "kill", "destroy", "create":
			return m, keyboard.FetchContainers(m.Docker, true)
		}
	}
	return m, nil
}

func handleToastTick(m *state.AppModel, _ state.ToastTick) (*state.AppModel, tea.Cmd) {
	if m.ToastTimer > 0 {
		m.ToastTimer--
		if m.ToastTimer <= 0 {
			m.ToastMessage = ""
			m.ToastLevel = 0
		}
	}
	return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return state.ToastTick{}
	})
}

func handleHostStatsTick(m *state.AppModel, _ state.HostStatsTick) (*state.AppModel, tea.Cmd) {
	stats := docker.ReadHostStats()
	m.HostCPU = stats.CPUPercent
	m.HostMem = stats.MemPercent
	m.HostDisk = stats.DiskStr
	m.HostCPUCores = stats.CPUCores
	m.HostMemUsed = stats.MemUsed
	m.HostMemTotal = stats.MemTotal
	return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return state.HostStatsTick{}
	})
}

func handleRuntimeHealthTick(m *state.AppModel, _ state.RuntimeHealthTick) (*state.AppModel, tea.Cmd) {
	interval, timeout := 3, 2
	if m.Config != nil {
		interval = m.Config.Runtime.Health.IntervalSec
		timeout = m.Config.Runtime.Health.TimeoutSec
	}
	if interval <= 0 {
		interval = 3
	}
	if timeout <= 0 {
		timeout = 2
	}
	cmds := []tea.Cmd{tea.Tick(time.Duration(interval)*time.Second, func(time.Time) tea.Msg {
		return state.RuntimeHealthTick{}
	})}
	if m.Docker != nil {
		client, name := m.Docker, m.ConnectionTarget
		cmds = append(cmds, func() tea.Msg {
			return state.RuntimeHealthResult{Name: name, Error: client.PingTimeout(time.Duration(timeout) * time.Second)}
		})
	}
	return m, tea.Batch(cmds...)
}

func handleRuntimeHealthResult(m *state.AppModel, msg state.RuntimeHealthResult) (*state.AppModel, tea.Cmd) {
	if msg.Name != m.ConnectionTarget || m.Docker == nil {
		return m, nil
	}
	threshold := 2
	if m.Config != nil && m.Config.Runtime.Health.FailureThreshold > 0 {
		threshold = m.Config.Runtime.Health.FailureThreshold
	}
	if msg.Error != nil {
		m.HealthFailures++
		if m.HealthFailures >= threshold && !m.HealthDegraded {
			m.HealthDegraded = true
			m.Connected = false
			m.ConnectionError = msg.Error.Error()
			m.ErrorMessage = fmt.Sprintf("runtime health check failed: %v", msg.Error)
			m.ErrorCount++
			keyboard.ShowToastWarn(m, fmt.Sprintf("Runtime %s disconnected: %v", m.ConnectionTarget, msg.Error))
		}
		return m, nil
	}
	wasDegraded := m.HealthDegraded
	m.HealthFailures = 0
	m.HealthDegraded = false
	if wasDegraded {
		m.Connected = true
		m.ConnectionError = ""
		keyboard.ShowToastNow(m, fmt.Sprintf("Runtime %s recovered", m.ConnectionTarget))
	}
	return m, nil
}

func handleRuntimeProbeResult(m *state.AppModel, msg state.RuntimeProbeResult) (*state.AppModel, tea.Cmd) {
	if m.Mode != state.ModeRuntimeSelect {
		return m, nil
	}
	if m.RuntimeSelectorError == nil {
		m.RuntimeSelectorError = make(map[string]string)
	}
	if msg.Error != nil {
		m.RuntimeSelectorError[msg.Name] = msg.Error.Error()
	} else {
		delete(m.RuntimeSelectorError, msg.Name)
	}
	return m, nil
}

func handleEscTimeout(m *state.AppModel, _ state.EscTimeout) (*state.AppModel, tea.Cmd) {
	m.EscPending = false
	m.InfoMessage = ""
	return m, nil
}

func handleKeyHintTick(m *state.AppModel, _ state.KeyHintTick) (*state.AppModel, tea.Cmd) {
	if m.KeyHintTimer > 0 {
		m.KeyHintTimer--
		if m.KeyHintTimer <= 0 {
			m.KeyHint = ""
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
	if m.KeyStrokeTimer <= 0 {
		return m, nil
	}
	m.KeyStrokeTimer--
	if m.KeyStrokeTimer <= 0 {
		m.KeyStrokeBuffer = nil
		m.LastKeyStroke = nil
		return m, nil
	}
	return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return state.KeyStrokeTick{}
	})
}

func handleFilterExitTimeout(m *state.AppModel, msg state.FilterExitTimeout) (*state.AppModel, tea.Cmd) {
	if msg.Token == m.FilterExitToken {
		m.FilterExitPending = false
	}
	return m, nil
}
