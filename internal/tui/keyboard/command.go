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
		m.Navigation.Mode = state.ModeNormal
		m.Navigation.CommandInput.Reset()
		return m, nil
	case keys.KeyTab:
		m.Navigation.CommandInput.Set(autocompleteCommand(m.Navigation.CommandInput.Text))
		return m, nil
	default:
		editQueryInput(key, &m.Navigation.CommandInput)
		return m, nil
	}
}

func executeCommand(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	cmd := strings.TrimSpace(strings.ToLower(m.Navigation.CommandInput.Text))
	m.Navigation.Mode = state.ModeNormal
	m.Navigation.CommandInput.Reset()

	switch cmd {
	case keys.CommandCompose:
		m.Navigation.ActivePanel = state.PanelCompose
		ShowToastNow(m, "✓ Switched to Compose")
	case keys.CommandImages:
		m.Navigation.ActivePanel = state.PanelImages
	case keys.CommandContainers:
		m.Navigation.ActivePanel = state.PanelContainers
	case keys.CommandVolumes:
		m.Navigation.ActivePanel = state.PanelVolumes
	case keys.CommandNetworks:
		m.Navigation.ActivePanel = state.PanelNetworks
	case keys.CommandLogs:
		if m.Log.LogContainerID != "" {
			m.Navigation.Mode = state.ModeLogView
		}
	case keys.CommandHelp:
		m.Navigation.Mode = state.ModeHelp
		m.Navigation.ActivePanel = state.PanelHelp
	case keys.CommandRename:
		return openRenameDialog(m)
	case keys.CommandTop:
		return openTopView(m)
	case keys.CommandPort:
		return openPortDetail(m)
	case keys.CommandConnInfo:
		return showConnectionInfo(m)
	}
	if m.Navigation.Mode == state.ModeNormal {
		m.Selection.ClearMarks()
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
