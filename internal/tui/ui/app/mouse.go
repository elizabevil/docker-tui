package view

import (
	tea "charm.land/bubbletea/v2"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/history"
)

// IdleCursorPosition returns the terminal position (0-based) where the
// cursor should be parked when no text input widget is active, so IME
// candidate windows render over the footer rail instead of the table.
// ok is false when a text input is active and the cursor must stay
// wherever the renderer last wrote it.
func IdleCursorPosition(m *state.AppModel) (x, y int, ok bool) {
	if m == nil || m.Viewport.Width == 0 || m.Viewport.Height == 0 {
		return 0, 0, false
	}
	class := ClassifyTerminal(m.Viewport.Width, m.Viewport.Height)
	if class != TerminalStandard {
		return 0, 0, false
	}
	if activeTextInput(m) {
		return 0, 0, false
	}
	rep := ResolveLayout(m)
	// Park on the first footer row so the IME candidate window appears
	// at the bottom of the screen. The rightmost column keeps the
	// visible cursor block clear of the shortcut text on the left.
	return m.Viewport.Width - 1, rep.FooterTop, true
}

// activeTextInput reports whether any text-editing widget currently
// owns the keyboard. While one is active the cursor must follow the
// input field, so the idle IME parking must not apply.
func activeTextInput(m *state.AppModel) bool {
	if m == nil {
		return false
	}
	if m.Dialog.Kind != state.DialogNone {
		return true
	}
	switch m.Navigation.Mode {
	case state.ModeFilter, state.ModeSearch, state.ModeCommand,
		state.ModeImagePull, state.ModeRename, state.ModeResourceCreate,
		state.ModeImageWorkflow, state.ModeExecShell, state.ModeContainerForm,
		state.ModeActionBar:
		return true
	}
	return false
}

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

// String returns a short token for tests and debug output.
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

// panelRect describes the resolved top-row of the panel rail together
// with the body geometry available for rows (excluding border + title).
// Tests assert against this struct so the mouse logic stays decoupled
// from the rendering helpers.
//
// bodyLeft and bodyWidth are added for BR-040 (Dialog 居中于 panel
// body): they describe the inner X range of the panel content so
// dialog widgets can center within the panel rather than the whole
// terminal.
type panelRect struct {
	panelTop  int // absolute Y (terminal rows) of the panel border top
	bodyTop   int // absolute Y of the first row inside the panel content
	bodyRows  int // number of data rows the panel can display
	bodyLeft  int // absolute X (terminal columns) of the first column inside the panel content
	bodyWidth int // inner width excluding left + right border
}

// LayoutReport captures the resolved rails + panel geometry for a given
// model. HitTest uses it to answer "what rail does this click belong to"
// in O(1).
type LayoutReport struct {
	Class         TerminalClass
	Panel         panelRect
	HeaderBottom  int
	MessageBottom int
	QueryBottom   int
	FooterTop     int
}

// ResolveLayout returns the layout geometry for the current viewport.
// It does not depend on any rendering, which makes it safe to call from
// the Update path.
func ResolveLayout(m *state.AppModel) LayoutReport {
	class := ClassifyTerminal(m.Viewport.Width, m.Viewport.Height)
	switch class {
	case TerminalUnsupported:
		return LayoutReport{Class: class}
	case TerminalCompact:
		return resolveCompactLayout(m)
	default:
		return resolveStandardLayout(m)
	}
}

func resolveCompactLayout(m *state.AppModel) LayoutReport {
	// No top/bottom percentage margins in compact mode: the rails pack
	// from row 0. Order: header, message, query (only when active),
	// panel, footer.
	plan := planFor(TerminalCompact, m.Viewport.Height)
	queryActive := queryKindFor(m) != queryNone
	if queryActive {
		plan.query = compactQueryLines
	}
	plan.message = 1
	plan.footer = max(1, plan.footer)
	panelH := max(m.Viewport.Height-(plan.header+plan.message+plan.query+plan.footer), 1)
	panelTop := plan.header + plan.message + plan.query
	bodyTop := panelTop + 2 // border + title
	return LayoutReport{
		Class:         TerminalCompact,
		Panel:         panelRect{panelTop: panelTop, bodyTop: bodyTop, bodyRows: panelH - 2, bodyLeft: 2, bodyWidth: m.Viewport.Width - 4},
		HeaderBottom:  plan.header,
		MessageBottom: plan.header + plan.message,
		QueryBottom:   plan.header + plan.message + plan.query,
		FooterTop:     m.Viewport.Height - plan.footer,
	}
}

func resolveStandardLayout(m *state.AppModel) LayoutReport {
	// Replicate the standard layout's margin + rail math so the mouse
	// hit-test stays in sync with what RenderApp draws. The values here
	// intentionally mirror the variable names in layout.go.
	windowCfg := m.Dependencies.Config.UI.Window
	mt := max(windowCfg.MarginTopPercent, 0)
	if mt > 15 {
		mt = 15
	}
	mb := max(windowCfg.MarginBottomPercent, 0)
	if mb > 15 {
		mb = 15
	}
	marginTop := m.Viewport.Height * mt / 100
	marginBot := m.Viewport.Height * mb / 100
	usableH := m.Viewport.Height - marginTop - marginBot
	if usableH < 10 {
		usableH = m.Viewport.Height
		marginTop = 0
		marginBot = 0
	}
	contentWidthPct := windowCfg.ContentWidthPercent
	if contentWidthPct <= 0 || contentWidthPct > 100 {
		contentWidthPct = 90
	}
	usableW := m.Viewport.Width * contentWidthPct / 100
	if usableW < 50 {
		usableW = m.Viewport.Width
	}
	padH := (m.Viewport.Width - usableW) / 2
	rails := calculateRailHeights(usableH)
	plan := railPlan(rails)
	headerBottom := marginTop + plan.header
	messageBottom := headerBottom + plan.message
	queryBottom := messageBottom + plan.query
	panelTop := queryBottom
	footerTop := marginTop + usableH - plan.footer
	bodyTop := panelTop + 2 // border + title
	return LayoutReport{
		Class:         TerminalStandard,
		Panel:         panelRect{panelTop: panelTop, bodyTop: bodyTop, bodyRows: plan.panel - 2, bodyLeft: padH + 2, bodyWidth: usableW - 4},
		HeaderBottom:  headerBottom,
		MessageBottom: messageBottom,
		QueryBottom:   queryBottom,
		FooterTop:     footerTop,
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
	rep := ResolveLayout(m)
	switch {
	case y < rep.HeaderBottom:
		return HitHeader, y
	case y < rep.MessageBottom:
		return HitMessage, y - rep.HeaderBottom
	case y < rep.QueryBottom:
		return HitQuery, y - rep.MessageBottom
	case y < rep.FooterTop:
		row := y - rep.Panel.bodyTop
		return HitPanel, row
	default:
		return HitFooter, y - rep.FooterTop
	}
}

// ApplyMouseClick routes a left-click on the panel body to the right
// cursor action. Non-left clicks and clicks outside the panel body are
// no-ops so the keyboard contract stays intact.
//
// Click semantics (panel rail only):
//   - List panels (containers/images/volumes/networks/audit): move the
//     cursor to the clicked row, clamped to the current data length.
//   - Image panel inside "containers of image" view: routes to the
//     container cursor.
//   - Detail / Log / Top / AuditDetail / Help / RuntimeSelector: scroll
//     to the clicked line by treating the click as a soft scroll.
//
// Click on the panel border or title (row < 0): scroll up by one step
// to give the user a way to "grab" the panel edge to scroll.
func ApplyMouseClick(m *state.AppModel, msg tea.MouseClickMsg) {
	if m == nil {
		return
	}
	if msg.Button != tea.MouseLeft {
		// Right/middle clicks are intentionally ignored to keep the
		// keyboard contract intact; users can still scroll with the
		// wheel and navigate with the arrow keys.
		return
	}
	hit, row := HitTest(m, msg.X, msg.Y)
	if hit != HitPanel {
		return
	}
	if row < 0 {
		scrollActivePanel(m, -1)
		return
	}
	// Transient modes (Detail / Log / Top / AuditDetail) take precedence
	// over the active panel: their body is the page content, not a
	// cursor list.
	if m.Navigation.Mode != state.ModeNormal {
		scrollActivePanel(m, row-scrollStep(m))
		return
	}
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		clickListCursor(m, &m.Resources.Containers.Cursor, &m.Resources.Containers.ViewOffset,
			len(m.Resources.Containers.Items), row)
	case state.PanelImages:
		// The "containers of image" view reuses the image list's
		// Container* fields. We compute the visible row count from the
		// image's own data; if the user clicked the summary list (no
		// ContainersViewID active), fall back to the image cursor.
		if m.Resources.Images.ContainersViewID != "" {
			clickListCursor(m, &m.Resources.Images.ContainerCursor, &m.Resources.Images.ContainerOffset,
				imageContainersCount(m), row)
		} else {
			clickListCursor(m, &m.Resources.Images.Cursor, &m.Resources.Images.ViewOffset,
				len(m.Resources.Images.Items), row)
		}
	case state.PanelVolumes:
		clickListCursor(m, &m.Resources.Volumes.Cursor, &m.Resources.Volumes.ViewOffset,
			len(m.Resources.Volumes.Items), row)
	case state.PanelNetworks:
		clickListCursor(m, &m.Resources.Networks.Cursor, &m.Resources.Networks.ViewOffset,
			len(m.Resources.Networks.Items), row)
	case state.PanelAudit:
		clickListCursor(m, &m.Audit.Cursor, &m.Audit.ViewOffset, m.Audit.Total(), row)
	default:
		scrollActivePanel(m, row-scrollStep(m))
	}
}

// clickListCursor moves a panel cursor to the clicked visible row. When
// the click is below the visible row count the cursor is moved to the
// last item; when the offset pointer is provided it is also kept in
// range so the new cursor stays visible.
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
		rep := ResolveLayout(m)
		bodyRows := max(rep.Panel.bodyRows, 1)
		if *offset > *cursor {
			*offset = *cursor
		}
		if *cursor >= *offset+bodyRows {
			*offset = max(*cursor-bodyRows+1, 0)
		}
	}
	m.Metrics.StatsActive = false
}

// scrollActivePanel scrolls the active scrollable panel by the given
// delta. Positive deltas scroll down (newer rows appear), negative
// deltas scroll up (older rows appear). The sign convention matches the
// existing tea.MouseWheelMsg handler.
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

// scrollStep returns the standard scroll step used by the existing
// wheel handler so click-scroll lines up with wheel-scroll.
func scrollStep(m *state.AppModel) int {
	// Match the +/-3 step used by handleMouseWheel; positive means
	// "scroll content down (forward)".
	_ = m
	return 3
}

// imageContainersCount returns the size of the "containers of image"
// view for the active image. The model does not store the list itself;
// the image panel re-derives it on every render by looking at the
// container list for matching ContainerImageID. We replicate the
// filter here so the mouse click can move the cursor safely. The count
// is intentionally loose: a slightly stale number clamps the cursor
// to the visible rows instead of pushing it past the end.
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
