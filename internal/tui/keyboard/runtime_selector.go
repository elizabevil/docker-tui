package keyboard

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func openRuntimeSelector(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.RuntimeSelectorDisabled || m.Connection.Pool == nil {
		ShowToastNow(m, "runtime selection unavailable in --host mode")
		return m, nil
	}
	names := m.Connection.Pool.KnownHostNames()
	if len(names) == 0 {
		ShowToastNow(m, "no runtime connections")
		return m, nil
	}
	m.Navigation.Mode = state.ModeRuntimeSelect
	m.Connection.RuntimeSelectorCursor = 0
	for i, name := range names {
		if name == m.Connection.Pool.ActiveName() {
			m.Connection.RuntimeSelectorCursor = i
			break
		}
	}
	if m.Connection.RuntimeSelectorError == nil {
		m.Connection.RuntimeSelectorError = make(map[string]string)
	}
	cmds := make([]tea.Cmd, 0, len(names))
	for _, name := range names {
		if name == m.Connection.Pool.ActiveName() {
			continue
		}
		connectionName := name
		cmds = append(cmds, func() tea.Msg {
			return state.RuntimeProbeResult{Name: connectionName, Error: m.Connection.Pool.Probe(connectionName, runtimeHealthTimeout(m))}
		})
	}
	return m, tea.Batch(cmds...)
}

func runtimeHealthTimeout(m *state.AppModel) time.Duration {
	if m.Dependencies.Config != nil {
		return m.Dependencies.Config.Runtime.Health.Timeout()
	}
	return config.DefaultConfig().Runtime.Health.Timeout()
}

func handleRuntimeSelectorKey(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	names := m.Connection.Pool.KnownHostNames()
	if len(names) == 0 {
		m.Navigation.Mode = state.ModeNormal
		return m, nil
	}
	switch key {
	case keys.KeyEsc:
		m.Navigation.Mode = state.ModeNormal
	case keys.KeyUp, keys.KeyK:
		m.Connection.RuntimeSelectorCursor = (m.Connection.RuntimeSelectorCursor - 1 + len(names)) % len(names)
	case keys.KeyDown, keys.KeyJ:
		m.Connection.RuntimeSelectorCursor = (m.Connection.RuntimeSelectorCursor + 1) % len(names)
	case keys.KeyEnter:
		name := names[m.Connection.RuntimeSelectorCursor]
		m.Connection.Begin()
		return m, runtimeConnectionCmd(m, name)
	}
	return m, nil
}

func runtimeConnectionCmd(m *state.AppModel, name string) tea.Cmd {
	return func() tea.Msg {
		if err := m.Connection.Pool.Connect(name, 2*time.Second); err != nil {
			return state.DockerConnected{Name: name, Error: err}
		}
		return state.DockerConnected{Name: name, Client: m.Connection.Pool.ActiveClient()}
	}
}

func selectorError(m *state.AppModel, name string, err error) {
	if m.Connection.RuntimeSelectorError == nil {
		m.Connection.RuntimeSelectorError = make(map[string]string)
	}
	m.Connection.RuntimeSelectorError[name] = fmt.Sprint(err)
}
