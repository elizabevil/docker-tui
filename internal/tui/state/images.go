package state

import (
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
)

type ImageSortColumn int

const (
	ImageSortByRepo ImageSortColumn = iota
	ImageSortByID
	ImageSortBySize
	ImageSortByCreated
)

type (
	ImagesLoaded struct {
		Images []dockerclient.ImageSummary
		Error  error
	}

	ImageDetailLoaded struct {
		ImageID string
		Content string
		Error   error
	}
)

type ImageListModel struct {
	Items            []dockerclient.ImageSummary
	Cursor           int
	ViewOffset       int
	Loading          bool
	Error            error
	Filter           string
	SortBy           ImageSortColumn
	SortAsc          bool
	ExpandedID       string // deprecated — kept for compatibility
	ContainersViewID string
	ContainerCursor  int
	ContainerOffset  int
}

func NewImageListModel() *ImageListModel {
	return &ImageListModel{
		Items:   make([]dockerclient.ImageSummary, 0),
		Cursor:  0,
		SortAsc: true,
	}
}

func (m *ImageListModel) Selected() *dockerclient.ImageSummary {
	items := m.SortedItems()
	if len(items) == 0 || m.Cursor < 0 || m.Cursor >= len(items) {
		return nil
	}
	return &items[m.Cursor]
}

func (m *ImageListModel) Len() int {
	return len(m.Items)
}

func (m *ImageListModel) SetFilter(query string) {
	m.Filter = query
}

func (m *ImageListModel) FilterText() string {
	return m.Filter
}

func (m *ImageListModel) FilteredItems() []dockerclient.ImageSummary {
	if m.Filter == "" {
		return m.Items
	}
	filtered := make([]dockerclient.ImageSummary, 0, len(m.Items))
	for _, img := range m.Items {
		for _, tag := range img.RepoTags {
			if contains(tag, m.Filter) {
				filtered = append(filtered, img)
				break
			}
		}
		if contains(img.ID, m.Filter) {
			filtered = append(filtered, img)
		}
	}
	return filtered
}

func (m *ImageListModel) SortedItems() []dockerclient.ImageSummary {
	items := m.FilteredItems()
	sortImageSlice(items, m.SortBy, m.SortAsc)
	return items
}

func sortImageSlice(items []dockerclient.ImageSummary, col ImageSortColumn, asc bool) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			less := false
			switch col {
			case ImageSortByRepo:
				a := items[i].Registry + ":" + repoTag(items[i])
				b := items[j].Registry + ":" + repoTag(items[j])
				less = a < b
			case ImageSortByID:
				less = items[i].ID < items[j].ID
			case ImageSortBySize:
				less = items[i].Size < items[j].Size
			case ImageSortByCreated:
				less = items[i].Created < items[j].Created
			}
			if asc != less {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func repoTag(img dockerclient.ImageSummary) string {
	if len(img.RepoTags) > 0 {
		return img.RepoTags[0]
	}
	return "<none>:<none>"
}
