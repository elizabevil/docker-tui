package mouse

import (
	"strings"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/containers"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/images"
	view "github.com/elizabevil/docker-tui/internal/tui/ui/app"
)

type mouseListTarget struct {
	cursor      *int
	offset      *int
	total       int
	visibleRows int
	rowCount    func(int) int
	layout      component.TableHitLayout
}

func applyMouseListClick(m *state.AppModel, x, rawRow int) bool {
	rep := view.ResolveLayout(m)
	if x < rep.Panel.BodyLeft || x >= rep.Panel.BodyLeft+rep.Panel.BodyWidth {
		return false
	}
	target, ok := mouseListTargetFor(m, rep)
	if !ok {
		return false
	}
	visibleRow, ok := target.layout.DataRowAt(rawRow, target.visibleRows)
	if !ok {
		return false
	}
	item, ok := itemAtVisualRow(*target.offset, target.total, visibleRow, target.rowCount)
	if !ok {
		return false
	}
	*target.cursor = item
	ensureMouseListVisible(target)
	m.Metrics.StatsActive = false
	return true
}
func mouseListTargetFor(m *state.AppModel, rep view.LayoutReport) (mouseListTarget, bool) {
	panelHeight := max(rep.FooterTop-rep.Panel.PanelTop, 1)
	bodyHeight := view.PanelBodyHeight(panelHeight)
	selectionPreview := m.Navigation.Mode != state.ModeFilter && component.SelectionInfoEnabled()
	panelWidth := max(rep.Panel.BodyWidth, 1)

	newTarget := func(cursor, offset *int, total, visibleRows, bannerLines int, rowCount func(int) int) mouseListTarget {
		return mouseListTarget{
			cursor:      cursor,
			offset:      offset,
			total:       total,
			visibleRows: visibleRows,
			rowCount:    rowCount,
			layout: component.NewTableHitLayout(component.TableHitLayoutInput{
				SelectionPreview:    selectionPreview,
				SelectionPreviewPad: component.SelectionPreviewPadding(),
				Total:               total,
				BannerLines:         bannerLines,
			}),
		}
	}

	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		items := m.Resources.Containers.SortedItems()
		compactPorts := panelWidth < 80
		return newTarget(
			&m.Resources.Containers.Cursor,
			&m.Resources.Containers.ViewOffset,
			len(items),
			component.CalcTableRowHeight(bodyHeight, selectionPreview),
			markedBannerLines(m, state.PanelContainers),
			func(index int) int {
				if index < 0 || index >= len(items) {
					return 1
				}
				return len(containers.FormatPorts(items[index].PortBindings, compactPorts))
			},
		), true
	case state.PanelImages:
		if m.Resources.Images.ContainersViewID != "" {
			items := imageContainersForMouse(m)
			compactPorts := panelWidth < 80
			return newTarget(
				&m.Resources.Images.ContainerCursor,
				&m.Resources.Images.ContainerOffset,
				len(items),
				component.CalcRowHeight(bodyHeight),
				0,
				func(index int) int {
					if index < 0 || index >= len(items) {
						return 1
					}
					return len(containers.FormatPorts(items[index].PortBindings, compactPorts))
				},
			), true
		}
		return newTarget(
			&m.Resources.Images.Cursor,
			&m.Resources.Images.ViewOffset,
			len(m.Resources.Images.SortedItems()),
			component.CalcTableRowHeight(bodyHeight, selectionPreview),
			0,
			oneRow,
		), true
	case state.PanelVolumes:
		return newTarget(
			&m.Resources.Volumes.Cursor,
			&m.Resources.Volumes.ViewOffset,
			len(m.Resources.Volumes.FilteredItems()),
			component.CalcTableRowHeight(bodyHeight, selectionPreview),
			0,
			oneRow,
		), true
	case state.PanelNetworks:
		return newTarget(
			&m.Resources.Networks.Cursor,
			&m.Resources.Networks.ViewOffset,
			len(m.Resources.Networks.FilteredItems()),
			component.CalcTableRowHeight(bodyHeight, selectionPreview),
			0,
			oneRow,
		), true
	case state.PanelAudit:
		return newTarget(
			&m.Audit.Cursor,
			&m.Audit.ViewOffset,
			m.Audit.Total(),
			component.CalcTableRowHeight(bodyHeight, false),
			0,
			oneRow,
		), true
	default:
		return mouseListTarget{}, false
	}
}

func markedBannerLines(m *state.AppModel, panel state.PanelType) int {
	if m == nil || len(m.Selection.PanelMarks[panel]) == 0 {
		return 0
	}
	switch panel {
	case state.PanelContainers:
		return strings.Count(component.MarkedItemsBanner(panel, m.Selection.PanelMarks[panel]), "\n") + 1
	}
	return 0
}

func oneRow(int) int {
	return 1
}

func itemAtVisualRow(offset, total, visualRow int, rowCount func(int) int) (int, bool) {
	if offset < 0 {
		offset = 0
	}
	if visualRow < 0 || offset >= total {
		return 0, false
	}
	for index := offset; index < total; index++ {
		rows := max(1, rowCount(index))
		if visualRow < rows {
			return index, true
		}
		visualRow -= rows
	}
	return 0, false
}

func ensureMouseListVisible(target mouseListTarget) {
	if target.cursor == nil || target.offset == nil || target.total <= 0 {
		return
	}
	if *target.cursor < 0 {
		*target.cursor = 0
	}
	if *target.cursor >= target.total {
		*target.cursor = target.total - 1
	}
	if *target.offset < 0 {
		*target.offset = 0
	}
	if *target.offset >= target.total {
		*target.offset = target.total - 1
	}
	if *target.cursor < *target.offset {
		*target.offset = *target.cursor
		return
	}
	visibleRow := visualRowForItem(*target.offset, *target.cursor, target.rowCount)
	if visibleRow >= target.visibleRows {
		*target.offset = *target.cursor
	}
}

func visualRowForItem(offset, cursor int, rowCount func(int) int) int {
	row := 0
	for index := offset; index < cursor; index++ {
		row += max(1, rowCount(index))
	}
	return row
}

func imageContainersForMouse(m *state.AppModel) []dockerclient.ContainerSummary {
	return images.MatchedContainersByImage(m.Resources.Images, m.Resources.Containers, m.Resources.Images.ContainersViewID)
}
