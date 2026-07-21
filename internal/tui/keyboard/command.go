package keyboard

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleCommandInput(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch key {
	case keys.KeyEnter:
		return executeCommand(m)
	case keys.KeyEsc:
		m.Mode = state.ModeNormal
		m.CommandInput.Reset()
		return m, nil
	case keys.KeyTab:
		m.CommandInput.Set(autocompleteCommand(m.CommandInput.Text))
		return m, nil
	default:
		editQueryInput(key, &m.CommandInput)
		return m, nil
	}
}

func executeCommand(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	cmd := strings.TrimSpace(strings.ToLower(m.CommandInput.Text))
	m.Mode = state.ModeNormal
	m.CommandInput.Reset()

	switch cmd {
	case keys.CommandCompose:
		m.ActivePanel = state.PanelCompose
		ShowToastNow(m, "✓ Switched to Compose")
	case keys.CommandImages:
		m.ActivePanel = state.PanelImages
	case keys.CommandContainers:
		m.ActivePanel = state.PanelContainers
	case keys.CommandVolumes:
		m.ActivePanel = state.PanelVolumes
	case keys.CommandNetworks:
		m.ActivePanel = state.PanelNetworks
	case keys.CommandLogs:
		if m.LogContainerID != "" {
			m.Mode = state.ModeLogView
		}
	case keys.CommandHelp:
		m.Mode = state.ModeHelp
		m.ActivePanel = state.PanelHelp
	}
	return m, nil
}

func autocompleteCommand(input string) string {
	input = strings.ToLower(input)
	for _, cmd := range keys.Commands() {
		if strings.HasPrefix(cmd, input) {
			return cmd
		}
	}
	return input
}
