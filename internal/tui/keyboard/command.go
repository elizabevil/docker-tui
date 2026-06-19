package keyboard

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

var availableCommands = []string{"compose", "images", "containers", "volumes", "networks", "logs", "help"}

func handleCommandInput(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	switch key {
	case keys.KeyEnter:
		return executeCommand(m)
	case keys.KeyEsc:
		m.Mode = state.ModeNormal
		m.FilterText = ""
		m.FilterCursor = 0
		return m, nil
	case "backspace":
		if len(m.FilterText) > 0 {
			m.FilterText = m.FilterText[:len(m.FilterText)-1]
		}
		m.FilterCursor = len([]rune(m.FilterText))
		return m, nil
	case "tab":
		m.FilterText = autocompleteCommand(m.FilterText)
		m.FilterCursor = len([]rune(m.FilterText))
		return m, nil
	default:
		if len(key) == 1 {
			m.FilterText += key
			m.FilterCursor = len([]rune(m.FilterText))
		}
		return m, nil
	}
}

func executeCommand(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	cmd := strings.TrimSpace(strings.ToLower(m.FilterText))
	m.Mode = state.ModeNormal
	m.FilterText = ""
	m.FilterCursor = 0

	switch cmd {
	case "compose":
		m.ActivePanel = state.PanelCompose
		ShowToastNow(m, "✓ Switched to Compose")
	case "images":
		m.ActivePanel = state.PanelImages
	case "containers":
		m.ActivePanel = state.PanelContainers
	case "volumes":
		m.ActivePanel = state.PanelVolumes
	case "networks":
		m.ActivePanel = state.PanelNetworks
	case "logs":
		if m.LogContainerID != "" {
			m.Mode = state.ModeLogView
		}
	case "help":
		m.Mode = state.ModeHelp
		m.ActivePanel = state.PanelHelp
	}
	return m, nil
}

func autocompleteCommand(input string) string {
	input = strings.ToLower(input)
	for _, cmd := range availableCommands {
		if strings.HasPrefix(cmd, input) {
			return cmd
		}
	}
	return input
}
