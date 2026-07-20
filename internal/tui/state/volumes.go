package state

import (
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
)

type (
	VolumesLoaded struct {
		Volumes []dockerclient.VolumeItem
		Error   error
	}
)

type VolumeListModel struct {
	Items      []dockerclient.VolumeItem
	Cursor     int
	ViewOffset int
	Loading    bool
	Error      error
	Filter     string
	DetailName string
}

func NewVolumeListModel() *VolumeListModel {
	return &VolumeListModel{
		Items:  make([]dockerclient.VolumeItem, 0),
		Cursor: 0,
	}
}

func (m *VolumeListModel) Selected() *dockerclient.VolumeItem {
	items := m.FilteredItems()
	if len(items) == 0 || m.Cursor < 0 || m.Cursor >= len(items) {
		return nil
	}
	return &items[m.Cursor]
}

func (m *VolumeListModel) Len() int {
	return len(m.Items)
}

func (m *VolumeListModel) SetFilter(query string) {
	m.Filter = query
}

func (m *VolumeListModel) FilterText() string {
	return m.Filter
}

func (m *VolumeListModel) FilteredItems() []dockerclient.VolumeItem {
	if m.Filter == "" {
		return m.Items
	}
	filtered := make([]dockerclient.VolumeItem, 0, len(m.Items))
	for _, v := range m.Items {
		if contains(v.Name, m.Filter) || contains(v.Driver, m.Filter) || contains(v.Scope, m.Filter) || contains(v.Mountpoint, m.Filter) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
