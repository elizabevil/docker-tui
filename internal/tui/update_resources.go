package tui

import (
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleContainersLoaded(m *state.AppModel, msg state.ContainersLoaded) (*state.AppModel, tea.Cmd) {
	m.Containers.Loading = false
	if msg.Error != nil {
		m.Containers.Error = msg.Error
		m.FeedbackState.RecordError(msg.Error.Error())
	} else {
		m.Containers.Items = msg.Containers
		m.Containers.Error = nil
		if m.Containers.Cursor >= len(msg.Containers) {
			m.Containers.Cursor = 0
		}
	}
	// Start auto-stats polling on first container load
	if m.Docker != nil && !m.StatsActive {
		m.StatsActive = true
		return m, func() tea.Msg { return state.StatsTick{} }
	}
	return m, nil
}

func handleImagesLoaded(m *state.AppModel, msg state.ImagesLoaded) (*state.AppModel, tea.Cmd) {
	m.Images.Loading = false
	if msg.Error != nil {
		m.Images.Error = msg.Error
		m.FeedbackState.RecordError(msg.Error.Error())
	} else {
		m.Images.Items = msg.Images
		m.Images.Error = nil
	}
	return m, nil
}

func handleVolumesLoaded(m *state.AppModel, msg state.VolumesLoaded) (*state.AppModel, tea.Cmd) {
	m.Volumes.Loading = false
	if msg.Error != nil {
		m.Volumes.Error = msg.Error
		m.FeedbackState.RecordError(msg.Error.Error())
	} else {
		m.Volumes.Items = msg.Volumes
		m.Volumes.Error = nil
	}
	return m, nil
}

func handleNetworksLoaded(m *state.AppModel, msg state.NetworksLoaded) (*state.AppModel, tea.Cmd) {
	m.Networks.Loading = false
	if msg.Error != nil {
		m.Networks.Error = msg.Error
		m.FeedbackState.RecordError(msg.Error.Error())
	} else {
		m.Networks.Items = msg.Networks
		m.Networks.Error = nil
	}
	return m, nil
}
