package tui

import (
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func handleContainersLoaded(m *state.AppModel, msg state.ContainersLoaded) (*state.AppModel, tea.Cmd) {
	m.Resources.Containers.Loading = false
	if msg.Error != nil {
		m.Resources.Containers.Error = msg.Error
		m.Feedback.RecordError(msg.Error.Error())
	} else {
		m.Resources.Containers.Items = msg.Containers
		m.Resources.Containers.Error = nil
		if m.Resources.Containers.Cursor >= len(msg.Containers) {
			m.Resources.Containers.Cursor = 0
		}
	}
	// Start auto-stats polling on first container load
	if m.Connection.Docker != nil && !m.Metrics.StatsActive {
		m.Metrics.StatsActive = true
		return m, func() tea.Msg { return state.StatsTick{} }
	}
	return m, nil
}

func handleImagesLoaded(m *state.AppModel, msg state.ImagesLoaded) (*state.AppModel, tea.Cmd) {
	m.Resources.Images.Loading = false
	if msg.Error != nil {
		m.Resources.Images.Error = msg.Error
		m.Feedback.RecordError(msg.Error.Error())
	} else {
		m.Resources.Images.Items = msg.Images
		m.Resources.Images.Error = nil
	}
	return m, nil
}

func handleVolumesLoaded(m *state.AppModel, msg state.VolumesLoaded) (*state.AppModel, tea.Cmd) {
	m.Resources.Volumes.Loading = false
	if msg.Error != nil {
		m.Resources.Volumes.Error = msg.Error
		m.Feedback.RecordError(msg.Error.Error())
	} else {
		m.Resources.Volumes.Items = msg.Volumes
		m.Resources.Volumes.Error = nil
	}
	return m, nil
}

func handleNetworksLoaded(m *state.AppModel, msg state.NetworksLoaded) (*state.AppModel, tea.Cmd) {
	m.Resources.Networks.Loading = false
	if msg.Error != nil {
		m.Resources.Networks.Error = msg.Error
		m.Feedback.RecordError(msg.Error.Error())
	} else {
		m.Resources.Networks.Items = msg.Networks
		m.Resources.Networks.Error = nil
	}
	return m, nil
}
