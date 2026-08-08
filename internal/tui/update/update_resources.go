package update

import (
	"github.com/elizabevil/docker-tui/internal/tui/keyboard"
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
		if anchor := m.Resources.Containers.SelectionAnchorID; anchor != "" {
			for i, container := range msg.Containers {
				if container.ID == anchor {
					m.Resources.Containers.Cursor = i
					break
				}
			}
			m.Resources.Containers.SelectionAnchorID = ""
		}
		if m.Resources.Containers.Cursor >= len(msg.Containers) {
			m.Resources.Containers.Cursor = 0
		}
	}
	var cmds []tea.Cmd
	// R08-03: refresh the compose project list whenever containers
	// change so the panel reflects the new aggregate.
	if m.Connection.Engine != nil {
		cmds = append(cmds, keyboard.FetchComposeProjects(m.Connection.Engine))
	}
	if m.Connection.Engine != nil && !m.Metrics.StatsActive {
		m.Metrics.StatsActive = true
		cmds = append(cmds, func() tea.Msg { return state.StatsTick{} })
	}
	if len(cmds) == 0 {
		return m, nil
	}
	return m, tea.Batch(cmds...)
}

// handleComposeProjectsLoaded replaces m.Resources.Compose with the
// runtime-supplied summaries. R08-03 F2: completely-empty projects
// stay hidden (ListProjects never synthesises them); an error keeps
// the prior slice so the panel still renders the cached view.
func handleComposeProjectsLoaded(m *state.AppModel, msg state.ComposeProjectsLoaded) (*state.AppModel, tea.Cmd) {
	if msg.Error != nil {
		m.Feedback.RecordError(msg.Error.Error())
		return m, nil
	}
	m.Resources.Compose = msg.Projects
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
