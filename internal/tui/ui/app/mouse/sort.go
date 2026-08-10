package mouse

import (
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// applyHeaderSort toggles the sort column for the active panel when the
// click hit matches a column header. Returns true if a sort change was
// dispatched. Asc/desc follow the BR-032 contract: clicking the active
// column flips the direction; clicking a new column sets it ascending.
func applyHeaderSort(m *state.AppModel, key string) bool {
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		target, ok := containerSortColumnForKey(key)
		if !ok {
			return false
		}
		toggleContainerSort(m, target)
		return true
	case state.PanelImages:
		target, ok := imageSortColumnForKey(key)
		if !ok {
			return false
		}
		toggleImageSort(m, target)
		return true
	case state.PanelNetworks:
		target, ok := networkSortColumnForKey(key)
		if !ok {
			return false
		}
		toggleNetworkSort(m, target)
		return true
	}
	return false
}

func containerSortColumnForKey(key string) (state.ContainerSortColumn, bool) {
	switch key {
	case "name":
		return state.ContainerSortByName, true
	case "id":
		return state.ContainerSortByID, true
	case "state":
		return state.ContainerSortByState, true
	case "created":
		return state.ContainerSortByCreated, true
	case "cpu":
		return state.ContainerSortByCPU, true
	case "mem":
		return state.ContainerSortByMem, true
	}
	return 0, false
}

func toggleContainerSort(m *state.AppModel, col state.ContainerSortColumn) {
	if m.Resources.Containers == nil {
		return
	}
	cm := m.Resources.Containers
	prevID := containerIDAtCursor(cm)
	prevBy := cm.SortBy
	toggleSortColumn(&cm.SortBy, &cm.SortAsc, col)
	if prevBy != col && prevID != "" {
		items := cm.SortedItems()
		cm.Cursor = cm.TrackCursor(items, prevID)
		cm.ViewOffset = 0
	}
}

func containerIDAtCursor(cm *state.ContainerListModel) string {
	if cm.Cursor < 0 || cm.Cursor >= len(cm.Items) {
		return ""
	}
	return cm.Items[cm.Cursor].ID
}

func imageSortColumnForKey(key string) (state.ImageSortColumn, bool) {
	switch key {
	case "name", "registry":
		return state.ImageSortByRepo, true
	case "id":
		return state.ImageSortByID, true
	case "size":
		return state.ImageSortBySize, true
	case "created":
		return state.ImageSortByCreated, true
	}
	return 0, false
}

func toggleImageSort(m *state.AppModel, col state.ImageSortColumn) {
	if m.Resources.Images == nil {
		return
	}
	toggleSortColumn(&m.Resources.Images.SortBy, &m.Resources.Images.SortAsc, col)
}

func networkSortColumnForKey(key string) (state.NetworkSortColumn, bool) {
	switch key {
	case "name":
		return state.NetworkSortByName, true
	case "driver":
		return state.NetworkSortByDriver, true
	case "created":
		return state.NetworkSortByCreated, true
	}
	return 0, false
}

func toggleNetworkSort(m *state.AppModel, col state.NetworkSortColumn) {
	if m.Resources.Networks == nil {
		return
	}
	toggleSortColumn(&m.Resources.Networks.SortBy, &m.Resources.Networks.SortAsc, col)
}

// sortColumnID accepts any named int type used for sortable column IDs.
type sortColumnID interface {
	~int
}

// toggleSortColumn applies the BR-032 header-click contract:
//   - clicking the active column flips the asc flag
//   - clicking a new column switches to it and resets asc to true
//
// The cursor and viewport are intentionally untouched so SortedItems()
// can track the selected item by ID through the new sort order.
func toggleSortColumn[T sortColumnID](sortBy *T, sortAsc *bool, col T) {
	if *sortBy == col {
		*sortAsc = !*sortAsc
		return
	}
	*sortBy = col
	*sortAsc = true
}