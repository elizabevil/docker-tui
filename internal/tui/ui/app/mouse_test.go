package view

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestRenderActionBarInStandardAndCompactLayouts(t *testing.T) {
	for _, viewport := range [][2]int{{120, 32}, {79, 19}} {
		m := newModel(viewport[0], viewport[1])
		m.Navigation.Mode = state.ModeActionBar
		m.Navigation.ActionBar.Open()
		rendered := RenderApp(m)
		if !strings.Contains(rendered, "Action Bar") {
			t.Fatalf("Action Bar missing at %dx%d", viewport[0], viewport[1])
		}
	}
}

func newModel(width, height int) *state.AppModel {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Viewport.Width = width
	m.Viewport.Height = height
	return m
}

func TestHistoryMouseWheelMovesCursor(t *testing.T) {
	m := newModel(120, 32)
	m.Navigation.Mode = state.ModeHistory
	m.History.Layers = make([]dockerclient.ImageHistoryLayer, 20)
	scrollActivePanel(m, 3)
	if m.History.Cursor != 3 {
		t.Fatalf("history cursor = %d, want 3", m.History.Cursor)
	}
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

// TestPanelBodyGeometryStandard verifies BR-040 fields (bodyLeft / bodyWidth)
// are populated correctly for the standard tier.
func TestPanelBodyGeometryStandard(t *testing.T) {
	m := newModel(120, 32)
	rep := ResolveLayout(m)
	if rep.Class != TerminalStandard {
		t.Fatalf("class = %d, want TerminalStandard", rep.Class)
	}
	if rep.Panel.bodyLeft < 0 || rep.Panel.bodyLeft >= m.Viewport.Width {
		t.Fatalf("bodyLeft = %d, want 0..%d", rep.Panel.bodyLeft, m.Viewport.Width)
	}
	if rep.Panel.bodyWidth <= 0 {
		t.Fatalf("bodyWidth = %d, want positive", rep.Panel.bodyWidth)
	}
	if rep.Panel.bodyLeft+rep.Panel.bodyWidth > m.Viewport.Width {
		t.Errorf("panel body extends past terminal: bodyLeft=%d + bodyWidth=%d > termW=%d",
			rep.Panel.bodyLeft, rep.Panel.bodyWidth, m.Viewport.Width)
	}
	if rep.Panel.bodyTop+rep.Panel.bodyRows > rep.FooterTop {
		t.Errorf("panel body extends into footer: bodyTop=%d + bodyRows=%d > FooterTop=%d",
			rep.Panel.bodyTop, rep.Panel.bodyRows, rep.FooterTop)
	}
}

// TestPanelBodyGeometryCompact verifies the same fields for the compact tier.
func TestPanelBodyGeometryCompact(t *testing.T) {
	m := newModel(79, 19)
	rep := ResolveLayout(m)
	if rep.Class != TerminalCompact {
		t.Fatalf("class = %d, want TerminalCompact", rep.Class)
	}
	if rep.Panel.bodyLeft != 2 {
		t.Errorf("compact bodyLeft = %d, want 2", rep.Panel.bodyLeft)
	}
	expectedWidth := m.Viewport.Width - 4
	if rep.Panel.bodyWidth != expectedWidth {
		t.Errorf("compact bodyWidth = %d, want %d", rep.Panel.bodyWidth, expectedWidth)
	}
	if rep.Panel.bodyWidth <= 0 {
		t.Fatalf("bodyWidth must be positive: %d", rep.Panel.bodyWidth)
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
