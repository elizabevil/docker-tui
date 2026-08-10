package mouse

import (
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/compose"
	view "github.com/elizabevil/docker-tui/internal/tui/ui/app"
)

// applyComposeSubviewClick switches ComposeFocus based on which compose
// column was clicked. Right clicks return true (cursor updated), left
// clicks return false so the generic list handler updates the project
// cursor.
func applyComposeSubviewClick(m *state.AppModel, x, bodyRow int) bool {
	if m == nil {
		return false
	}
	if m.Navigation.ActivePanel != state.PanelCompose {
		return false
	}
	if m.Compose.ComposeContainerViewID != "" {
		return false
	}
	rep := view.ResolveLayout(m)
	leftW, rightW, bodyH := compose.Geometry(m, rep.Panel.BodyWidth, bodyHeightFromRep(rep))
	rightStart := rep.Panel.BodyLeft + leftW + 1
	if rightW <= 0 || bodyH <= 0 {
		return false
	}
	if x >= rightStart && x < rightStart+rightW {
		m.Compose.Focus(1)
		m.Metrics.StatsActive = false
		services := compose.ServiceNamesFor(m)
		if len(services) == 0 {
			return true
		}
		if bodyRow < 0 || bodyRow >= bodyH {
			return true
		}
		headerRows := composeRightHeaderRows(m, bodyH)
		relative := bodyRow - headerRows
		if relative < 0 || relative >= len(services) {
			return true
		}
		m.Compose.ComposeServiceCursor = relative
		return true
	}
	if x >= rep.Panel.BodyLeft && x < rightStart {
		m.Compose.Focus(0)
	}
	return false
}

func bodyHeightFromRep(rep view.LayoutReport) int {
	return max(rep.FooterTop-rep.Panel.PanelTop, 1)
}

// composeRightHeaderRows returns the number of body rows the right
// column uses before its first data row. It accounts for the panel
// header and the project title strip rendered by compose.
func composeRightHeaderRows(m *state.AppModel, bodyH int) int {
	_ = m
	return 3
}

// ServiceNamesFor is an indirection for the mouse layer so we can
// inject test data. Implementation: gather from the model like
// compose.RenderPanel does.
func composeServiceNamesFor(m *state.AppModel) []string { return compose.ServiceNamesFor(m) }

var _ = component.SelectionInfoEnabled
