package mouse

import (
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/containers"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/images"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/networks"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/volumes"
	view "github.com/elizabevil/docker-tui/internal/tui/ui/app"
)

// mouseListColumnGeometry returns the column defs, the desired content
// widths and the gutter for the active list panel.
func mouseListColumnGeometry(m *state.AppModel, rep view.LayoutReport) (cols []tables.ColumnDef, widths []int, gap, totalWidth int) {
	gap = component.GetTableLayout().ColumnSpacing
	totalWidth = rep.Panel.BodyWidth
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		cols = containers.ColumnDescriptors(m, totalWidth)
	case state.PanelImages:
		cols = images.ColumnDescriptors(m, totalWidth)
	case state.PanelVolumes:
		cols = volumes.ColumnDescriptors(m, totalWidth)
	case state.PanelNetworks:
		cols = networks.ColumnDescriptors(m, totalWidth)
	default:
		return nil, nil, 0, 0
	}
	widths = make([]int, len(cols))
	for i, def := range cols {
		widths[i] = len([]rune(def.Header))
	}
	return cols, widths, gap, totalWidth
}
