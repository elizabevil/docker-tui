package state

import (
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type (
	VolumesLoaded struct {
		Volumes []runtimeapi.Volume
		Error   error
	}
)

type VolumeListModel struct {
	Items      []runtimeapi.Volume
	Cursor     int
	ViewOffset int
	Loading    bool
	Error      error
	Filter     string
	DetailName string
}

func NewVolumeListModel() *VolumeListModel {
	return &VolumeListModel{
		Items:  make([]runtimeapi.Volume, 0),
		Cursor: 0,
	}
}

func (m *VolumeListModel) Selected() *runtimeapi.Volume {
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

func (m *VolumeListModel) FilteredItems() []runtimeapi.Volume {
	if m.Filter == "" {
		return m.Items
	}
	filtered := make([]runtimeapi.Volume, 0, len(m.Items))
	for _, v := range m.Items {
		if contains(v.Name, m.Filter) || contains(v.Driver, m.Filter) || contains(v.Scope, m.Filter) || contains(v.Mountpoint, m.Filter) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
