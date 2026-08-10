package mouse

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/app"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/compose"
)

func TestRenderActionBarInStandardAndCompactLayouts(t *testing.T) {
	for _, viewport := range [][2]int{{120, 32}, {79, 19}} {
		m := newModel(viewport[0], viewport[1])
		m.Navigation.Mode = state.ModeActionBar
		m.Navigation.ActionBar.Open()
		rendered := view.RenderApp(m)
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
	rep := view.ResolveLayout(m)
	if rep.Class != view.TerminalStandard {
		t.Fatalf("class = %d, want view.TerminalStandard", rep.Class)
	}
	if rep.Panel.BodyRows < 5 {
		t.Fatalf("panel body rows too small: %d", rep.Panel.BodyRows)
	}
	// Click in the middle of the panel body should map to a body row.
	click := tea.MouseClickMsg{X: 10, Y: rep.Panel.BodyTop + 2}
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
	rep := view.ResolveLayout(m)
	if rep.Class != view.TerminalStandard {
		t.Fatalf("class = %d, want view.TerminalStandard", rep.Class)
	}
	if rep.Panel.BodyLeft < 0 || rep.Panel.BodyLeft >= m.Viewport.Width {
		t.Fatalf("bodyLeft = %d, want 0..%d", rep.Panel.BodyLeft, m.Viewport.Width)
	}
	if rep.Panel.BodyWidth <= 0 {
		t.Fatalf("bodyWidth = %d, want positive", rep.Panel.BodyWidth)
	}
	if rep.Panel.BodyLeft+rep.Panel.BodyWidth > m.Viewport.Width {
		t.Errorf("panel body extends past terminal: bodyLeft=%d + bodyWidth=%d > termW=%d",
			rep.Panel.BodyLeft, rep.Panel.BodyWidth, m.Viewport.Width)
	}
	if rep.Panel.BodyTop+rep.Panel.BodyRows > rep.FooterTop {
		t.Errorf("panel body extends into footer: bodyTop=%d + bodyRows=%d > FooterTop=%d",
			rep.Panel.BodyTop, rep.Panel.BodyRows, rep.FooterTop)
	}
}

// TestPanelBodyGeometryCompact verifies the same fields for the compact tier.
func TestPanelBodyGeometryCompact(t *testing.T) {
	m := newModel(79, 19)
	rep := view.ResolveLayout(m)
	if rep.Class != view.TerminalCompact {
		t.Fatalf("class = %d, want view.TerminalCompact", rep.Class)
	}
	if rep.Panel.BodyLeft != 2 {
		t.Errorf("compact bodyLeft = %d, want 2", rep.Panel.BodyLeft)
	}
	expectedWidth := m.Viewport.Width - 4
	if rep.Panel.BodyWidth != expectedWidth {
		t.Errorf("compact bodyWidth = %d, want %d", rep.Panel.BodyWidth, expectedWidth)
	}
	if rep.Panel.BodyWidth <= 0 {
		t.Fatalf("bodyWidth must be positive: %d", rep.Panel.BodyWidth)
	}
}

func TestHitTestCompactLayout(t *testing.T) {
	m := newModel(79, 19)
	rep := view.ResolveLayout(m)
	if rep.Class != view.TerminalCompact {
		t.Fatalf("class = %d, want view.TerminalCompact", rep.Class)
	}
	if rep.Panel.BodyRows < 3 {
		t.Fatalf("panel body rows too small for compact: %d", rep.Panel.BodyRows)
	}
	click := tea.MouseClickMsg{X: 5, Y: rep.Panel.BodyTop + 1}
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
	rep := view.ResolveLayout(m)
	// Panel border top (panelTop row itself).
	click := tea.MouseClickMsg{X: 5, Y: rep.Panel.PanelTop}
	hit, row := HitTest(m, click.X, click.Y)
	if hit != HitPanel {
		t.Fatalf("border click hit = %s, want panel", hit)
	}
	if row >= 0 {
		t.Fatalf("border click row = %d, want negative", row)
	}
	// Title row.
	click = tea.MouseClickMsg{X: 5, Y: rep.Panel.PanelTop + 1}
	hit, row = HitTest(m, click.X, click.Y)
	if hit != HitPanel || row != -1 {
		t.Fatalf("title click hit=%s row=%d, want panel/-1", hit, row)
	}
}

func TestApplyMouseClickSelectsFirstVisibleContainerRow(t *testing.T) {
	// Given
	m := newModel(80, 20)
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha"},
		{ID: "b", Name: "bravo"},
		{ID: "c", Name: "charlie"},
	}
	rep := view.ResolveLayout(m)

	// When: clicking the first rendered data row after preview, page info,
	// and header rows.
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + 4,
	})

	// Then
	if m.Resources.Containers.Cursor != 0 {
		t.Fatalf("cursor = %d, want first data item 0", m.Resources.Containers.Cursor)
	}
}

func TestApplyMouseClickUsesContainerViewOffset(t *testing.T) {
	// Given
	m := newModel(80, 20)
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha"},
		{ID: "b", Name: "bravo"},
		{ID: "c", Name: "charlie"},
		{ID: "d", Name: "delta"},
		{ID: "e", Name: "echo"},
	}
	m.Resources.Containers.ViewOffset = 2
	rep := view.ResolveLayout(m)

	// When
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + 4,
	})

	// Then
	if m.Resources.Containers.Cursor != 2 {
		t.Fatalf("cursor = %d, want visible row translated to offset 2", m.Resources.Containers.Cursor)
	}
}

func TestApplyMouseClickIgnoresContainerTableHeader(t *testing.T) {
	m := newModel(80, 30)
	m.Resources.Containers.SortBy = state.ContainerSortByID
	m.Resources.Containers.SortAsc = true
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha"},
		{ID: "b", Name: "bravo"},
	}
	rep := view.ResolveLayout(m)
	target, ok := mouseListTargetFor(m, rep)
	if !ok {
		t.Fatal("missing target")
	}
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + 2,
	})
	if m.Resources.Containers.Cursor != 0 {
		t.Fatalf("header click changed cursor to %d", m.Resources.Containers.Cursor)
	}
	if target.layout.DataStartRow() < 3 {
		t.Fatalf("dataStart = %d, want header row protected", target.layout.DataStartRow())
	}
}

func TestApplyMouseClickAccountsForMarkedBanner(t *testing.T) {
	// Given
	m := newModel(80, 20)
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha"},
		{ID: "b", Name: "bravo"},
	}
	m.Selection.Toggle(state.PanelContainers, "a")
	rep := view.ResolveLayout(m)

	// When: the banner adds one line before the table's first data row.
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + 5,
	})

	// Then
	if m.Resources.Containers.Cursor != 0 {
		t.Fatalf("marked-list first data click selected %d, want 0", m.Resources.Containers.Cursor)
	}
}

func TestApplyMouseClickSelectsContainerOwnerForPortRows(t *testing.T) {
	// Given
	m := newModel(80, 30)
	m.Resources.Containers.SortBy = state.ContainerSortByID
	m.Resources.Containers.SortAsc = true
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha", PortBindings: []dockerclient.PortBinding{
			{ContainerPort: 80, Protocol: "tcp", HostPort: 8080, HostIP: "0.0.0.0"},
			{ContainerPort: 443, Protocol: "tcp", HostPort: 8443, HostIP: "0.0.0.0"},
		}},
		{ID: "b", Name: "bravo"},
	}
	rep := view.ResolveLayout(m)

	// When: clicking the second visual port row belonging to alpha.
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + 5,
	})

	// Then
	if m.Resources.Containers.Cursor != 0 {
		t.Fatalf("second port row selected container %d, want owner 0", m.Resources.Containers.Cursor)
	}
}

func TestApplyMouseClickFilterModeHasNoSelectionPreview(t *testing.T) {
	// Given
	m := newModel(80, 20)
	m.Navigation.Mode = state.ModeFilter
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha"},
		{ID: "b", Name: "bravo"},
	}
	rep := view.ResolveLayout(m)

	// When: filter mode removes the selection preview, so data starts after
	// page info and header only.
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + 2,
	})

	// Then
	if m.Resources.Containers.Cursor != 0 {
		t.Fatalf("filter-mode first data click selected %d, want 0", m.Resources.Containers.Cursor)
	}
}
func TestApplyMouseClickContainerSortStaysSorted(t *testing.T) {
	m := newModel(80, 30)
	m.Resources.Containers.SortBy = state.ContainerSortByName
	m.Resources.Containers.SortAsc = true
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "c1", Name: "alpha"},
		{ID: "c2", Name: "alpha"},
		{ID: "c3", Name: "bravo"},
	}
	rep := view.ResolveLayout(m)
	target, ok := mouseListTargetFor(m, rep)
	if !ok {
		t.Fatal("missing target")
	}
	clickRow := target.layout.DataStartRow() + 1
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + clickRow,
	})
	if m.Resources.Containers.Cursor < 0 || m.Resources.Containers.Cursor >= 3 {
		t.Fatalf("cursor = %d, want in [0,3)", m.Resources.Containers.Cursor)
	}
	items := m.Resources.Containers.SortedItems()
	selected := items[m.Resources.Containers.Cursor]
	if selected.Name != "alpha" {
		t.Fatalf("selected name = %q, want alpha (stable secondary key broke)", selected.Name)
	}
	prev := items[0]
	stable := true
	for i := 1; i < len(items); i++ {
		if items[i].Name == prev.Name && items[i].ID < prev.ID {
			stable = false
		}
		prev = items[i]
	}
	if !stable {
		t.Fatalf("sort is not stable: %+v", items)
	}
}

func TestApplyMouseClickMovesContainerCursor(t *testing.T) {
	m := newModel(80, 30)
	m.Resources.Containers.SortBy = state.ContainerSortByID
	m.Resources.Containers.SortAsc = true
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha"},
		{ID: "b", Name: "bravo"},
		{ID: "c", Name: "charlie"},
		{ID: "d", Name: "delta"},
		{ID: "e", Name: "echo"},
	}
	rep := view.ResolveLayout(m)
	target, ok := mouseListTargetFor(m, rep)
	if !ok {
		t.Fatal("missing target")
	}
	clickRow := target.layout.DataStartRow() + 2

	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + clickRow,
	})

	if m.Resources.Containers.Cursor != 2 {
		t.Fatalf("cursor = %d, want 2", m.Resources.Containers.Cursor)
	}
}

func TestApplyMouseClickClampsCursor(t *testing.T) {
	m := newModel(80, 30)
	m.Resources.Containers.SortBy = state.ContainerSortByID
	m.Resources.Containers.SortAsc = true
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha"},
		{ID: "b", Name: "bravo"},
	}
	rep := view.ResolveLayout(m)
	target, ok := mouseListTargetFor(m, rep)
	if !ok {
		t.Fatal("missing target")
	}
	clickRow := target.layout.DataStartRow() + 1
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + clickRow,
	})
	if m.Resources.Containers.Cursor != 1 {
		t.Fatalf("cursor = %d, want 1 (clamped to last item)", m.Resources.Containers.Cursor)
	}
}

func TestApplyMouseClickIgnoresRightButton(t *testing.T) {
	m := newModel(80, 20)
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{{ID: "a"}, {ID: "b"}}
	rep := view.ResolveLayout(m)
	click := tea.MouseClickMsg{
		Button: tea.MouseRight,
		X:      5,
		Y:      rep.Panel.BodyTop,
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
	rep := view.ResolveLayout(m)
	// Click below the middle of the body; the resulting scroll delta
	// (row - 3) is positive, so the offset must move forward.
	click := tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + 4,
	}
	ApplyMouseClick(m, click)
	if m.Detail.DetailOffset <= start {
		t.Fatalf("detail scroll did not move forward: start=%d now=%d", start, m.Detail.DetailOffset)
	}
}

func TestApplyMouseClickSortsContainerByHeaderClick(t *testing.T) {
	m := newModel(80, 30)
	m.Resources.Containers.SortBy = state.ContainerSortByName
	m.Resources.Containers.SortAsc = true
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha"},
		{ID: "b", Name: "bravo"},
	}
	rep := view.ResolveLayout(m)
	target, ok := mouseListTargetFor(m, rep)
	if !ok {
		t.Fatal("missing target")
	}
	cols, widths, gap, _ := mouseListColumnGeometry(m, rep)
	x, ok := columnX(rep.Panel.BodyLeft, cols, widths, gap, "id")
	if !ok {
		t.Fatal("missing column X for id")
	}
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      x,
		Y:      rep.Panel.BodyTop + target.layout.HeaderRow(),
	})
	if m.Resources.Containers.SortBy != state.ContainerSortByID {
		t.Fatalf("container sort = %d, want ID", m.Resources.Containers.SortBy)
	}
	if !m.Resources.Containers.SortAsc {
		t.Fatal("container sort asc = false, want true on new column")
	}
}

func TestApplyMouseClickHeaderGapSnapsToNearestColumn(t *testing.T) {
	m := newModel(120, 32)
	m.Navigation.ActivePanel = state.PanelNetworks
	m.Resources.Networks.Items = []dockerclient.Network{
		{Name: "alpha", Driver: "bridge"},
		{Name: "bravo", Driver: "bridge"},
	}
	rep := view.ResolveLayout(m)
	target, ok := mouseListTargetFor(m, rep)
	if !ok {
		t.Fatal("missing target")
	}
	cols, widths, gap, _ := mouseListColumnGeometry(m, rep)
	gapX := rep.Panel.BodyLeft + widths[0]
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      gapX,
		Y:      rep.Panel.BodyTop + target.layout.HeaderRow(),
	})
	if m.Resources.Networks.SortBy != state.NetworkSortByName {
		t.Fatalf("network sort = %d, want Name after gap click", m.Resources.Networks.SortBy)
	}
	_ = cols
	_ = gap
}

func TestApplyMouseClickHeaderHitImageByID(t *testing.T) {
	m := newModel(120, 32)
	m.Navigation.ActivePanel = state.PanelImages
	m.Resources.Images.Items = []dockerclient.ImageSummary{
		{ID: "a", RepoTags: []string{"a:latest"}},
		{ID: "b", RepoTags: []string{"b:latest"}},
	}
	rep := view.ResolveLayout(m)
	target, ok := mouseListTargetFor(m, rep)
	if !ok {
		t.Fatal("missing target")
	}
	cols, widths, gap, _ := mouseListColumnGeometry(m, rep)
	x, ok := columnX(rep.Panel.BodyLeft, cols, widths, gap, "id")
	if !ok {
		t.Fatal("missing column X for id")
	}
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      x,
		Y:      rep.Panel.BodyTop + target.layout.HeaderRow(),
	})
	if m.Resources.Images.SortBy != state.ImageSortByID {
		t.Fatalf("image sort = %d, want ID", m.Resources.Images.SortBy)
	}
}

func TestApplyMouseClickFlipsContainerSortDirection(t *testing.T) {
	m := newModel(80, 30)
	m.Resources.Containers.SortBy = state.ContainerSortByName
	m.Resources.Containers.SortAsc = true
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha"},
	}
	rep := view.ResolveLayout(m)
	target, ok := mouseListTargetFor(m, rep)
	if !ok {
		t.Fatal("missing target")
	}
	cols, widths, gap, _ := mouseListColumnGeometry(m, rep)
	x, ok := columnX(rep.Panel.BodyLeft, cols, widths, gap, "name")
	if !ok {
		t.Fatal("missing column X for name")
	}
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      x,
		Y:      rep.Panel.BodyTop + target.layout.HeaderRow(),
	})
	if m.Resources.Containers.SortAsc {
		t.Fatal("container sort asc = true, want false after click on active column")
	}
}

func TestApplyMouseClickSelectsComposeService(t *testing.T) {
	m := newModel(120, 32)
	m.Navigation.ActivePanel = state.PanelCompose
	m.Compose.ComposeFocus = 1
	m.Compose.ComposeServiceCursor = 0
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "c1", Name: "web-a", Image: "img-1", ComposeProject: "shop", ComposeService: "web"},
		{ID: "c2", Name: "web-b", Image: "img-1", ComposeProject: "shop", ComposeService: "web"},
		{ID: "c3", Name: "api", Image: "img-1", ComposeProject: "shop", ComposeService: "api"},
	}
	rep := view.ResolveLayout(m)
	leftW, rightW, _ := compose.Geometry(m, rep.Panel.BodyWidth, rep.FooterTop-rep.Panel.PanelTop)
	rightStart := rep.Panel.BodyLeft + leftW + 1
	names := composeServiceNamesFor(m)
	if len(names) < 2 {
		t.Fatalf("compose service names = %v, want >= 2", names)
	}
	clickY := rep.Panel.BodyTop + composeRightHeaderRows(m, max(rep.FooterTop-rep.Panel.PanelTop, 1)) + 1
	t.Logf("rightStart=%d rightW=%d clickY=%d names=%v", rightStart, rightW, clickY, names)
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      rightStart + 1,
		Y:      clickY,
	})
	if m.Compose.ComposeServiceCursor != 1 {
		t.Fatalf("compose service cursor = %d, want 1", m.Compose.ComposeServiceCursor)
	}
	_ = rightW
}

func TestApplyMouseClickComposeRightSwitchesFocusFromLeft(t *testing.T) {
	m := newModel(120, 32)
	m.Navigation.ActivePanel = state.PanelCompose
	m.Compose.ComposeFocus = 0
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "c1", Name: "web-a", Image: "img-1", ComposeProject: "shop", ComposeService: "web"},
		{ID: "c2", Name: "api", Image: "img-1", ComposeProject: "shop", ComposeService: "api"},
	}
	rep := view.ResolveLayout(m)
	leftW, rightW, bodyH := compose.Geometry(m, rep.Panel.BodyWidth, rep.FooterTop-rep.Panel.PanelTop)
	rightStart := rep.Panel.BodyLeft + leftW + 1
	clickY := rep.Panel.BodyTop + composeRightHeaderRows(m, bodyH) + 1
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      rightStart + 1,
		Y:      clickY,
	})
	if m.Compose.ComposeFocus != 1 {
		t.Fatalf("compose focus = %d, want 1 after right-side click", m.Compose.ComposeFocus)
	}
	if m.Compose.ComposeServiceCursor != 1 {
		t.Fatalf("compose service cursor = %d, want 1 (api)", m.Compose.ComposeServiceCursor)
	}
	_ = rightW
}

func TestApplyMouseClickComposeLeftSwitchesFocusFromRight(t *testing.T) {
	m := newModel(120, 32)
	m.Navigation.ActivePanel = state.PanelCompose
	m.Compose.ComposeFocus = 1
	m.Compose.ComposeServiceCursor = 0
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "c1", Name: "web-a", Image: "img-1", ComposeProject: "shop", ComposeService: "web"},
	}
	rep := view.ResolveLayout(m)
	leftW, _, _ := compose.Geometry(m, rep.Panel.BodyWidth, rep.FooterTop-rep.Panel.PanelTop)
	clickY := rep.Panel.BodyTop + 4
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      rep.Panel.BodyLeft + leftW/2,
		Y:      clickY,
	})
	if m.Compose.ComposeFocus != 0 {
		t.Fatalf("compose focus = %d, want 0 after left-side click", m.Compose.ComposeFocus)
	}
}

func TestApplyMouseClickSelectsImageContainerSubview(t *testing.T) {
	m := newModel(80, 30)
	m.Navigation.ActivePanel = state.PanelImages
	m.Resources.Images.ContainersViewID = "img-1"
	m.Resources.Images.ContainersViewRef = "img-1"
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "c1", Name: "alpha", Image: "img-1"},
		{ID: "c2", Name: "bravo", Image: "img-1"},
		{ID: "c3", Name: "charlie", Image: "other"},
	}
	rep := view.ResolveLayout(m)
	target, ok := mouseListTargetFor(m, rep)
	if !ok {
		t.Fatal("missing target")
	}
	clickRow := target.layout.DataStartRow() + 1
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + clickRow,
	})
	if m.Resources.Images.ContainerCursor != 1 {
		t.Fatalf("image subview cursor = %d, want 1", m.Resources.Images.ContainerCursor)
	}
}

func TestApplyMouseClickActivePanelVolumes(t *testing.T) {
	m := newModel(80, 30)
	m.Navigation.ActivePanel = state.PanelVolumes
	m.Resources.Volumes.Items = []dockerclient.Volume{
		{Name: "vol1"}, {Name: "vol2"}, {Name: "vol3"},
	}
	rep := view.ResolveLayout(m)
	target, ok := mouseListTargetFor(m, rep)
	if !ok {
		t.Fatal("missing target")
	}
	clickRow := target.layout.DataStartRow() + 1
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + clickRow,
	})
	if m.Resources.Volumes.Cursor != 1 {
		t.Fatalf("volumes cursor = %d, want 1", m.Resources.Volumes.Cursor)
	}
}

func TestApplyMouseClickEmptyListIsNoop(t *testing.T) {
	m := newModel(80, 20)
	rep := view.ResolveLayout(m)
	click := tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      5,
		Y:      rep.Panel.BodyTop + 1,
	}
	ApplyMouseClick(m, click)
	if m.Resources.Containers.Cursor != 0 {
		t.Fatalf("empty list cursor moved: %d", m.Resources.Containers.Cursor)
	}
}

// Regression: clicking a new sort column must not reset the cursor to 0;
// the cursor follows the selected container by ID to its new sort position.
func TestApplyMouseClickHeaderSortKeepsSelection(t *testing.T) {
	m := newModel(120, 32)
	m.Resources.Containers.SortBy = state.ContainerSortByID
	m.Resources.Containers.SortAsc = true
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{
		{ID: "a", Name: "alpha"},
		{ID: "b", Name: "bravo"},
		{ID: "c", Name: "charlie"},
	}
	m.Resources.Containers.Cursor = 2
	m.Resources.Containers.ViewOffset = 0
	rep := view.ResolveLayout(m)
	target, ok := mouseListTargetFor(m, rep)
	if !ok {
		t.Fatal("missing target")
	}
	cols, widths, gap, _ := mouseListColumnGeometry(m, rep)
	x, ok := columnX(rep.Panel.BodyLeft, cols, widths, gap, "name")
	if !ok {
		t.Fatal("missing column X for name")
	}
	ApplyMouseClick(m, tea.MouseClickMsg{
		Button: tea.MouseLeft,
		X:      x,
		Y:      rep.Panel.BodyTop + target.layout.HeaderRow(),
	})
	if m.Resources.Containers.SortBy != state.ContainerSortByName {
		t.Fatalf("sort = %d, want Name after header click", m.Resources.Containers.SortBy)
	}
	if m.Resources.Containers.Cursor == 0 {
		t.Fatalf("cursor reset to 0 — user selection lost; got %d", m.Resources.Containers.Cursor)
	}
	sorted := m.Resources.Containers.SortedItems()
	if sorted[m.Resources.Containers.Cursor].ID != "c" {
		t.Fatalf("cursor = %d, want row where ID=c (charlie) at index %d",
			m.Resources.Containers.Cursor, indexOfID(sorted, "c"))
	}
}

func indexOfID(items []dockerclient.ContainerSummary, id string) int {
	for i, it := range items {
		if it.ID == id {
			return i
		}
	}
	return -1
}
