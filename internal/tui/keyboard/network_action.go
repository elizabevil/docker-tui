package keyboard

import (
	"context"
	"fmt"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// networkRemoveCmd returns a tea.Cmd that removes a network.
func networkRemoveCmd(client runtimeapi.Engine, id string) tea.Cmd {
	return func() tea.Msg {
		err := client.Networks().Remove(context.Background(), id)
		return state.GenericActioned{Action: state.ActionRemoved, ID: id, Success: err == nil, Error: err}
	}
}

// doNetworkSort cycles through sort columns for the networks panel.
func doNetworkSort(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Navigation.ActivePanel != state.PanelNetworks {
		return m, nil
	}
	if m.Resources.Networks.SortBy >= state.NetworkSortByCreated {
		m.Resources.Networks.SortBy = state.NetworkSortByName
		m.Resources.Networks.SortAsc = !m.Resources.Networks.SortAsc
	} else {
		m.Resources.Networks.SortBy++
	}
	m.Resources.Networks.Cursor = 0
	m.Resources.Networks.ViewOffset = 0
	labels := map[state.NetworkSortColumn]string{
		state.NetworkSortByName:    "name",
		state.NetworkSortByDriver:  "driver",
		state.NetworkSortByCreated: "created",
	}
	m.Feedback.InfoMessage = fmt.Sprintf("Sort by %s (%v)", labels[m.Resources.Networks.SortBy], m.Resources.Networks.SortAsc)
	return m, nil
}

// doNetworkInspect opens the detail view for the selected network.
func doNetworkInspect(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelNetworks {
		return m, nil
	}
	net := m.Resources.Networks.Selected()
	if net == nil {
		return m, nil
	}
	detail, err := m.Connection.Engine.Networks().Inspect(context.Background(), net.ID)
	if err != nil {
		m.Feedback.RecordError(err.Error())
		return m, nil
	}
	m.Detail.SetNetworkDetail(detail)
	ToDetail(m, i18n.T("detail.title.network", net.Name), "")
	return m, nil
}

func doNetworkRemove(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelNetworks {
		return m, nil
	}
	net := m.Resources.Networks.Selected()
	if net == nil {
		return m, nil
	}
	confirmAction(m, "network-remove", net.ID, fmt.Sprintf("Remove network %s?", net.Name))
	m.Confirm.ConfirmAudit = beginAudit(m, "resource.network.delete", audit.NetworkTarget{ID: net.ID, Name: net.Name, Meta: audit.NetworkMeta{Driver: net.Driver}}, "Remove network "+net.Name)
	return m, nil
}
