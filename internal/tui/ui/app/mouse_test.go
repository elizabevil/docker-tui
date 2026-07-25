package view

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	tea "charm.land/bubbletea/v2"
)

func newModel(width, height int) *state.AppModel {
	m := state.NewAppModel(config.DefaultConfig(), nil, "test")
	m.Viewport.Width = width
	m.Viewport.Height = height
	return m
}

func TestHitTestStandardLayout(t *testing.T) {
	m := newModel(120, 32)
	rep := ResolveLayout(m)
	if rep.Class != TerminalStandard {
		t.Fatalf("class = %d, want TerminalStandard", rep.Class)
	}
	if rep.Panel.bodyRows < 5 {
		t.Fatalf("panel body rows too small: %d", rep.Panel.bodyRows)
	}
	// Click in the middle of the panel body should map to a body row.
	click := tea.MouseClickMsg{X: 10, Y: rep.Panel.bodyTop + 2}
	hit, row := HitTest(m, click.X, click.Y)
	if hit != HitPanel {
		t.Fatalf("hit = %s, want panel", hit)
	}
	if row != 2 {
		t.Fatalf("row = %d, want 2", row)
	}
}

func TestHitTestCompactLayout(t *testing.T) {
	m := newModel(79, 19)
	rep := ResolveLayout(m)
	if rep.Class != TerminalCompact {
		t.Fatalf("class = %d, want TerminalCompact", rep.Class)
	}
	if rep.Panel.bodyRows < 3 {
		t.Fatalf("panel body rows too small for compact: %d", rep.Panel.bodyRows)
	}
	click := tea.MouseClickMsg{X: 5, Y: rep.Panel.bodyTop + 1}
	hit, row := HitTest(m, click.X, click.Y)
	if hit != HitPanel || row != 1 {
		t.Fatalf("hit=%s row=%d, want panel/1", hit, row)
	}
}

func TestHitTestOutsideViewport(t *testing.T) {
	m := newModel(80, 20)
	if hit, _ := HitTest(m, -1, 0); hit != HitOutside {
		t.Fatalf("negative X should be outside")
	}
	if hit, _ := HitTest(m, 0, m.Viewport.Height); hit != HitOutside {
		t.Fatalf("Y == height should be outside")
	}
}

func TestHitTestBordersAndTitle(t *testing.T) {
	m := newModel(80, 20)
	rep := ResolveLayout(m)
	// Panel border top (panelTop row itself).
	click := tea.MouseClickMsg{X: 5, Y: rep.Panel.panelTop}
	hit, row := HitTest(m, click.X, click.Y)
	if hit != HitPanel {
		t.Fatalf("border click hit = %s, want panel", hit)
	}
	if row >= 0 {
		t.Fatalf("border click row = %d, want negative", row)
	}
	// Title row.
	click = tea.MouseClickMsg{X: 5, Y: rep.Panel.panelTop + 1}
	hit, row = HitTest(m, click.X, click.Y)
	if hit != HitPanel || row != -1 {
		t.Fatalf("title click hit=%s row=%d, want panel/-1", hit, row)
	}
}

func TestApplyMouseClickMovesContainerCursor(t *testing.T) {
	m := newModel(80, 20)
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha"},
		{ID: "b", Name: "bravo"},
		{ID: "c", Name: "charlie"},
		{ID: "d", Name: "delta"},
		{ID: "e", Name: "echo"},
	}
	rep := ResolveLayout(m)
	click := tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.bodyTop + 2,
	}
	ApplyMouseClick(m, click)
	if m.Resources.Containers.Cursor != 2 {
		t.Fatalf("cursor = %d, want 2", m.Resources.Containers.Cursor)
	}
}

func TestApplyMouseClickClampsCursor(t *testing.T) {
	m := newModel(80, 20)
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha"},
		{ID: "b", Name: "bravo"},
	}
	rep := ResolveLayout(m)
	// Click the last visible body row; should clamp to len-1.
	click := tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.bodyTop + rep.Panel.bodyRows - 1,
	}
	ApplyMouseClick(m, click)
	if m.Resources.Containers.Cursor != 1 {
		t.Fatalf("cursor = %d, want 1 (clamped)", m.Resources.Containers.Cursor)
	}
}

func TestApplyMouseClickIgnoresRightButton(t *testing.T) {
	m := newModel(80, 20)
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{{ID: "a"}, {ID: "b"}}
	rep := ResolveLayout(m)
	click := tea.MouseClickMsg{
		Button: tea.MouseRight,
		X:      5,
		Y:      rep.Panel.bodyTop,
	}
	ApplyMouseClick(m, click)
	if m.Resources.Containers.Cursor != 0 {
		t.Fatalf("right click changed cursor: %d", m.Resources.Containers.Cursor)
	}
}

func TestApplyMouseClickIgnoredOutsidePanel(t *testing.T) {
	m := newModel(80, 20)
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{{ID: "a"}, {ID: "b"}}
	// Click in the header rail (well above the panel).
	click := tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      0,
	}
	ApplyMouseClick(m, click)
	if m.Resources.Containers.Cursor != 0 {
		t.Fatalf("header click changed cursor: %d", m.Resources.Containers.Cursor)
	}
}

func TestApplyMouseClickScrollsDetail(t *testing.T) {
	m := newModel(80, 20)
	m.Navigation.Mode = state.ModeDetail
	m.Detail.DetailOffset = 5 // start mid-document so click can scroll either way
	start := m.Detail.DetailOffset
	rep := ResolveLayout(m)
	// Click below the middle of the body; the resulting scroll delta
	// (row - 3) is positive, so the offset must move forward.
	click := tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.bodyTop + 4,
	}
	ApplyMouseClick(m, click)
	if m.Detail.DetailOffset <= start {
		t.Fatalf("detail scroll did not move forward: start=%d now=%d", start, m.Detail.DetailOffset)
	}
}

func TestApplyMouseClickActivePanelVolumes(t *testing.T) {
	m := newModel(80, 20)
	m.Navigation.ActivePanel = state.PanelVolumes
	m.Resources.Volumes.Items = []dockerclient.Volume{
		{Name: "vol1"}, {Name: "vol2"}, {Name: "vol3"},
	}
	rep := ResolveLayout(m)
	click := tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.bodyTop + 1,
	}
	ApplyMouseClick(m, click)
	if m.Resources.Volumes.Cursor != 1 {
		t.Fatalf("volumes cursor = %d, want 1", m.Resources.Volumes.Cursor)
	}
}

func TestApplyMouseClickEmptyListIsNoop(t *testing.T) {
	m := newModel(80, 20)
	rep := ResolveLayout(m)
	click := tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.bodyTop + 1,
	}
	ApplyMouseClick(m, click)
	if m.Resources.Containers.Cursor != 0 {
		t.Fatalf("empty list cursor moved: %d", m.Resources.Containers.Cursor)
	}
}