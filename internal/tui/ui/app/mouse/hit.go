package mouse

import (
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/history"
	view "github.com/elizabevil/docker-tui/internal/tui/ui/app"
)

// LayoutHit identifies which rail a mouse coordinate belongs to. The
// handler in internal/tui/update uses it to decide what action to take
// without duplicating the layout math used by RenderApp.
type LayoutHit int

const (
	HitOutside LayoutHit = iota
	HitHeader
	HitMessage
	HitQuery
	HitPanel
	HitFooter
)

func (h LayoutHit) String() string {
	switch h {
	case HitHeader:
		return "header"
	case HitMessage:
		return "message"
	case HitQuery:
		return "query"
	case HitPanel:
		return "panel"
	case HitFooter:
		return "footer"
	default:
		return "outside"
	}
}

// HitTest returns the rail the (x, y) coordinate belongs to along with
// the relative row inside that rail. y=0 is the top terminal row.
// A click on the panel border (top/bottom) or padding rows is reported
// as HitPanel with a negative row so the caller can decide whether to
// treat it as a body click.
func HitTest(m *state.AppModel, x, y int) (LayoutHit, int) {
	if m == nil || m.Viewport.Width == 0 || m.Viewport.Height == 0 {
		return HitOutside, 0
	}
	if y < 0 || x < 0 || x >= m.Viewport.Width || y >= m.Viewport.Height {
		return HitOutside, 0
	}
	rep := view.ResolveLayout(m)
	switch {
	case y < rep.HeaderBottom:
		return HitHeader, y
	case y < rep.MessageBottom:
		return HitMessage, y - rep.HeaderBottom
	case y < rep.QueryBottom:
		return HitQuery, y - rep.MessageBottom
	case y < rep.FooterTop:
		row := y - rep.Panel.BodyTop
		return HitPanel, row
	default:
		return HitFooter, y - rep.FooterTop
	}
}

// scrollActivePanel applies a wheel delta to the panel that currently
// owns the keyboard focus. Matches the tea.MouseWheelMsg handler so
// click-scroll and wheel-scroll use the same targets.
func scrollActivePanel(m *state.AppModel, delta int) {
	switch m.Navigation.Mode {
	case state.ModeDetail:
		m.Detail.Scroll(delta)
	case state.ModeLogView:
		m.Log.Scroll(delta)
	case state.ModeHistory:
		total := len(history.FilterLayers(m.History.Layers, m.History.Filter))
		m.History.MoveCursor(delta, total, max(1, m.Viewport.Height-18))
	case state.ModeEvents:
		total := len(m.EventPanel.FilteredEvents())
		m.EventPanel.MoveCursor(delta, total, max(1, m.Viewport.Height-18))
	}
}

func scrollStep(m *state.AppModel) int {
	return 3
}

// clickListCursor moves a panel cursor to the clicked visible row,
// clamping to range and adjusting the offset so the new cursor stays
// in view.
func clickListCursor(m *state.AppModel, cursor *int, offset *int, total, row int) {
	if total <= 0 {
		*cursor = 0
		return
	}
	if row >= total {
		row = total - 1
	}
	if row < 0 {
		row = 0
	}
	*cursor = row
	if offset != nil {
		rep := view.ResolveLayout(m)
		bodyRows := max(rep.Panel.BodyRows, 1)
		if *offset > *cursor {
			*offset = *cursor
		}
		if *cursor >= *offset+bodyRows {
			*offset = max(*cursor-bodyRows+1, 0)
		}
	}
	m.Metrics.StatsActive = false
}

// imageContainersCount returns the size of the "containers of image"
// view for the active image. The model does not store the list itself;
// the image panel re-derives it on every render by looking at the
// container list for matching ContainerImageID. We replicate the
// filter here so the mouse click can move the cursor safely.
func imageContainersCount(m *state.AppModel) int {
	if m == nil || m.Resources.Images.ContainersViewID == "" {
		return 0
	}
	count := 0
	for _, c := range m.Resources.Containers.Items {
		if c.Image == m.Resources.Images.ContainersViewID {
			count++
		}
	}
	return count
}