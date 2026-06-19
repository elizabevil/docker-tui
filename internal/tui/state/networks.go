package state

import (
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
)

type (
	NetworksLoaded struct {
		Networks []dockerclient.NetworkItem
		Error    error
	}
)

type NetworkSortColumn int

const (
	NetworkSortByName NetworkSortColumn = iota
	NetworkSortByDriver
	NetworkSortByCreated
)

type NetworkListModel struct {
	Items      []dockerclient.NetworkItem
	Cursor     int
	ViewOffset int
	Loading    bool
	Error      error
	Filter     string
	SortBy     NetworkSortColumn
	SortAsc    bool
}

func NewNetworkListModel() *NetworkListModel {
	return &NetworkListModel{
		Items:   make([]dockerclient.NetworkItem, 0),
		Cursor:  0,
		SortAsc: true,
	}
}

func (m *NetworkListModel) Selected() *dockerclient.NetworkItem {
	items := m.FilteredItems()
	if len(items) == 0 || m.Cursor < 0 || m.Cursor >= len(items) {
		return nil
	}
	return &items[m.Cursor]
}

func (m *NetworkListModel) Len() int {
	return len(m.Items)
}

func (m *NetworkListModel) SetFilter(query string) {
	m.Filter = query
}

func (m *NetworkListModel) FilterText() string {
	return m.Filter
}

func (m *NetworkListModel) FilteredItems() []dockerclient.NetworkItem {
	if m.Filter == "" {
		return m.SortedItems()
	}
	filtered := make([]dockerclient.NetworkItem, 0, len(m.Items))
	for _, n := range m.Items {
		if contains(n.Name, m.Filter) || contains(n.ID, m.Filter) || contains(n.Driver, m.Filter) {
			filtered = append(filtered, n)
		}
	}
	return sortNetworkItems(filtered, m.SortBy, m.SortAsc)
}

func (m *NetworkListModel) SortedItems() []dockerclient.NetworkItem {
	return sortNetworkItems(m.Items, m.SortBy, m.SortAsc)
}

func sortNetworkItems(items []dockerclient.NetworkItem, col NetworkSortColumn, asc bool) []dockerclient.NetworkItem {
	sorted := make([]dockerclient.NetworkItem, len(items))
	copy(sorted, items)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			less := false
			switch col {
			case NetworkSortByDriver:
				less = sorted[i].Driver < sorted[j].Driver
			case NetworkSortByCreated:
				less = sorted[i].Created < sorted[j].Created
			default: // NetworkSortByName
				less = sorted[i].Name < sorted[j].Name
			}
			if asc != less {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}
