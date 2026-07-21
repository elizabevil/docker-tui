package keyboard

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// networkRemoveCmd returns a tea.Cmd that removes a network.
func networkRemoveCmd(client *docker.Client, id string) tea.Cmd {
	return func() tea.Msg {
		err := client.RemoveNetwork(id)
		return state.GenericActioned{Action: state.ActionRemoved, ID: id, Success: err == nil, Error: err}
	}
}

// doNetworkSort cycles through sort columns for the networks panel.
func doNetworkSort(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.ActivePanel != state.PanelNetworks {
		return m, nil
	}
	if m.Networks.SortBy >= state.NetworkSortByCreated {
		m.Networks.SortBy = state.NetworkSortByName
		m.Networks.SortAsc = !m.Networks.SortAsc
	} else {
		m.Networks.SortBy++
	}
	m.Networks.Cursor = 0
	m.Networks.ViewOffset = 0
	labels := map[state.NetworkSortColumn]string{
		state.NetworkSortByName:    "name",
		state.NetworkSortByDriver:  "driver",
		state.NetworkSortByCreated: "created",
	}
	m.InfoMessage = fmt.Sprintf("Sort by %s (%v)", labels[m.Networks.SortBy], m.Networks.SortAsc)
	return m, nil
}

// doNetworkInspect opens the detail view for the selected network.
func doNetworkInspect(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil || m.ActivePanel != state.PanelNetworks {
		return m, nil
	}
	net := m.Networks.Selected()
	if net == nil {
		return m, nil
	}
	rawJSON, err := m.Docker.InspectNetwork(net.ID)
	if err != nil {
		m.FeedbackState.RecordError(err.Error())
		return m, nil
	}
	m.DetailState.SetRaw(state.ResourceNetwork, rawJSON)
	ToDetail(m, i18n.T("detail.title.network", net.Name), "")
	return m, nil
}

func doNetworkRemove(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil || m.ActivePanel != state.PanelNetworks {
		return m, nil
	}
	net := m.Networks.Selected()
	if net == nil {
		return m, nil
	}
	confirmAction(m, "network-remove", net.ID, fmt.Sprintf("Remove network %s?", net.Name))
	m.ConfirmAudit = beginAudit(m, "resource.network.delete", audit.NetworkTarget{ID: net.ID, Name: net.Name, Meta: audit.NetworkMeta{Driver: net.Driver}}, "Remove network "+net.Name)
	return m, nil
}
