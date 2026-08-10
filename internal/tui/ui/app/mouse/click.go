package mouse

import (
	tea "charm.land/bubbletea/v2"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	view "github.com/elizabevil/docker-tui/internal/tui/ui/app"
)

func applyMouseHeaderSort(m *state.AppModel, x, bodyRow int) bool {
	if m.Navigation.ActivePanel == state.PanelCompose {
		return false
	}
	target, ok := mouseListTargetFor(m, view.ResolveLayout(m))
	if !ok {
		return false
	}
	if bodyRow != target.layout.HeaderRow() {
		return false
	}
	rep := view.ResolveLayout(m)
	cols, widths, gap, totalWidth := mouseListColumnGeometry(m, rep)
	if len(cols) == 0 || totalWidth <= 0 {
		return false
	}
	key, ok := component.ResolveColumnHit(rep.Panel.BodyLeft, x, cols, widths, gap)
	if !ok {
		return false
	}
	applyHeaderSort(m, key)
	return true
}

func ApplyMouseClick(m *state.AppModel, msg tea.MouseClickMsg) {
	if m == nil || msg.Button != tea.MouseLeft {
		return
	}
	hit, row := HitTest(m, msg.X, msg.Y)
	if hit != HitPanel || row < 0 {
		return
	}
	if m.Navigation.Mode != state.ModeNormal && m.Navigation.Mode != state.ModeFilter {
		scrollActivePanel(m, row-scrollStep(m))
		return
	}
	if applyMouseHeaderSort(m, msg.X, row) {
		return
	}
	if applyComposeSubviewClick(m, msg.X, row) {
		return
	}
	applyMouseListClick(m, msg.X, row)
}
