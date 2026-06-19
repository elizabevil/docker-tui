package tui

import (
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleContainerActioned(m *state.AppModel, msg state.ContainerActioned) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		keyboard.ShowToastNow(m, "✕ failed: "+msg.Action)
	} else {
		keyboard.ShowToastNow(m, "✓ "+i18n.T("toast."+msg.Action, msg.ID[:12]))
	}
	if m.Docker != nil {
		return m, keyboard.FetchContainers(m.Docker, true)
	}
	return m, nil
}

func handleImageActioned(m *state.AppModel, msg state.ImageActioned) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		keyboard.ShowToastNow(m, "✕ failed: "+msg.Action)
	} else {
		keyboard.ShowToastNow(m, "✓ "+i18n.T("toast."+msg.Action, msg.Ref))
	}
	if m.Docker != nil {
		return m, keyboard.FetchImages(m.Docker)
	}
	return m, nil
}

func handleImageDetailLoaded(m *state.AppModel, msg state.ImageDetailLoaded) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		m.ErrorMessage = msg.Error.Error()
		m.ErrorCount++
	} else {
		m.ImageDetailContent = msg.Content
		if m.Mode == state.ModeDetail {
			m.DetailOffset = 0
		}
	}
	return m, nil
}

func handleGenericActioned(m *state.AppModel, msg state.GenericActioned) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		keyboard.ShowToastNow(m, "✕ failed: "+msg.Action)
	} else {
		keyboard.ShowToastNow(m, "✓ "+i18n.T("toast."+msg.Action, msg.ID))
	}
	if m.Docker != nil {
		return m, tea.Batch(keyboard.FetchAll(m.Docker)...)
	}
	return m, nil
}
