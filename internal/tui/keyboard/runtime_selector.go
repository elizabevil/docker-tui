package keyboard

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func openRuntimeSelector(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.RuntimeSelectorDisabled || m.Pool == nil {
		ShowToastNow(m, "runtime selection unavailable in --host mode")
		return m, nil
	}
	names := m.Pool.KnownHostNames()
	if len(names) == 0 {
		ShowToastNow(m, "no runtime connections")
		return m, nil
	}
	m.Mode = state.ModeRuntimeSelect
	m.RuntimeSelectorCursor = 0
	for i, name := range names {
		if name == m.Pool.ActiveName() {
			m.RuntimeSelectorCursor = i
			break
		}
	}
	if m.RuntimeSelectorError == nil {
		m.RuntimeSelectorError = make(map[string]string)
	}
	cmds := make([]tea.Cmd, 0, len(names))
	for _, name := range names {
		if name == m.Pool.ActiveName() {
			continue
		}
		connectionName := name
		cmds = append(cmds, func() tea.Msg {
			return state.RuntimeProbeResult{Name: connectionName, Error: m.Pool.Probe(connectionName, runtimeHealthTimeout(m))}
		})
	}
	return m, tea.Batch(cmds...)
}

func runtimeHealthTimeout(m *state.AppModel) time.Duration {
	seconds := 2
	if m.Config != nil {
		_, seconds, _ = m.Config.Runtime.Health.Effective()
	}
	return time.Duration(seconds) * time.Second
}

func handleRuntimeSelectorKey(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	names := m.Pool.KnownHostNames()
	if len(names) == 0 {
		m.Mode = state.ModeNormal
		return m, nil
	}
	switch key {
	case keys.KeyEsc:
		m.Mode = state.ModeNormal
	case keys.KeyUp, keys.KeyK:
		m.RuntimeSelectorCursor = (m.RuntimeSelectorCursor - 1 + len(names)) % len(names)
	case keys.KeyDown, keys.KeyJ:
		m.RuntimeSelectorCursor = (m.RuntimeSelectorCursor + 1) % len(names)
	case keys.KeyEnter:
		name := names[m.RuntimeSelectorCursor]
		m.Connecting = true
		return m, runtimeConnectionCmd(m, name)
	}
	return m, nil
}

func runtimeConnectionCmd(m *state.AppModel, name string) tea.Cmd {
	return func() tea.Msg {
		if err := m.Pool.Connect(name, 2*time.Second); err != nil {
			return state.DockerConnected{Name: name, Error: err}
		}
		return state.DockerConnected{Name: name, Client: m.Pool.ActiveClient()}
	}
}

func selectorError(m *state.AppModel, name string, err error) {
	if m.RuntimeSelectorError == nil {
		m.RuntimeSelectorError = make(map[string]string)
	}
	m.RuntimeSelectorError[name] = fmt.Sprint(err)
}
