package compose

import (
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

// Geometry returns the rendered compose panel split (left/right column
// widths, total body height) so the mouse click handler can route
// clicks to the active focus column.
func Geometry(m *state.AppModel, panelWidth, panelHeight int) (leftW, rightW, bodyH int) {
	ratioL, ratioR := 4, 6
	if tc.Layout != nil && tc.Layout.Ratio.Left > 0 && tc.Layout.Ratio.Right > 0 {
		ratioL = tc.Layout.Ratio.Left
		ratioR = tc.Layout.Ratio.Right
	}
	leftW = panelWidth * ratioL / (ratioL + ratioR)
	rightW = panelWidth - leftW - 1
	if leftW < 10 {
		leftW = 10
	}
	if rightW < 10 {
		rightW = 10
	}
	bodyH = max(panelHeight-1, 3)
	return leftW, rightW, bodyH
}

// ProjectColumnDescriptors returns the active column set for the
// project list.
func ProjectColumnDescriptors() []tables.ColumnDef { return ProjectTableColumns() }

// ServiceColumnDescriptors returns the active column set for the
// service list.
func ServiceColumnDescriptors() []tables.ColumnDef { return ServiceTableColumns() }
