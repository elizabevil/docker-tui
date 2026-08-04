package view

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/footer"
)

// Terminal-size breakpoints used by RenderApp.
//
//	minimumTerminalWidth / minimumTerminalHeight
//	    hard floor: below this we cannot keep panel + footer readable
//	    and RenderApp returns the "Terminal too small" message.
//
//	compactMinWidth / compactMinHeight
//	    soft floor: between compactMin and minimum, RenderApp falls back
//	    to renderCompactApp which drops the multi-column header and the
//	    rich footer in favour of a single-line status + keymap strip.
//
// compactMin* must be < minimum* and both must be > 0. They are exported
// through the layout_test.go matrix; tune them together.
const (
	minimumTerminalWidth  = 80
	minimumTerminalHeight = 20

	compactMinWidth  = 60
	compactMinHeight = 14

	compactHeaderLines   = 1
	compactFooterLines   = 1
	compactQueryLines    = 3
	compactMessageLines  = 2
	compactPanelOverhead = 3
)

// TerminalClass reports which layout tier the current viewport maps to.
// It is exposed so tests and the overlay path can branch on the same
// classification RenderApp uses.
type TerminalClass int

const (
	TerminalUnsupported TerminalClass = iota // below compactMin*: only error message
	TerminalCompact                          // compact layout, single-line header + footer
	TerminalStandard                         // full header + footer + 3-line query rail
)

// ClassifyTerminal returns the layout tier for the given viewport size.
// Pure function so tests can pin the breakpoints without driving a real
// terminal.
func ClassifyTerminal(width, height int) TerminalClass {
	switch {
	case width < compactMinWidth || height < compactMinHeight:
		return TerminalUnsupported
	case width < minimumTerminalWidth || height < minimumTerminalHeight:
		return TerminalCompact
	default:
		return TerminalStandard
	}
}

// railPlan describes the resolved vertical layout for the current
// terminal tier. Tests assert against this struct rather than the rendered
// string so they can verify both height invariants and that the panel
// always gets at least one body row.
type railPlan struct {
	header  int
	message int
	query   int
	panel   int
	footer  int
}

func (p railPlan) total() int { return p.header + p.message + p.query + p.panel + p.footer }

// planFor returns the rail heights for a tier. The panel section always
// receives the remainder so it shrinks last, never the fixed rails.
func planFor(class TerminalClass, usableHeight int) railPlan {
	switch class {
	case TerminalCompact:
		// In compact mode we always reserve query=0 and grow it on
		// demand inside renderCompactApp (the rail border takes
		// 3 rows so we have to set them only when actually needed).
		fixed := compactHeaderLines + compactMessageLines + compactFooterLines
		return railPlan{
			header:  compactHeaderLines,
			message: compactMessageLines,
			query:   0,
			panel:   max(1, usableHeight-fixed),
			footer:  compactFooterLines,
		}
	default:
		std := calculateRailHeights(usableHeight)
		return railPlan{
			header:  std.header,
			message: std.message,
			query:   std.query,
			panel:   std.panel,
			footer:  std.footer,
		}
	}
}

// renderCompactApp is the small-terminal fallback. It keeps the active
// panel, the in-progress toast, and a one-line footer so core actions
// (navigation + the F-key shortcuts) remain reachable when the user
// shrinks their terminal below 80x20.
//
// The compact view intentionally drops:
//   - the multi-column header (engine / runtime / keystroke log)
//   - the rich footer (two shortcut rows + status bar)
//   - the 3-line query rail collapses to a single line
//
// Those are restored as soon as the viewport grows past the standard
// thresholds; nothing is permanently hidden.
func renderCompactApp(m *state.AppModel) string {
	plan := planFor(TerminalCompact, m.Viewport.Height)
	usableW := m.Viewport.Width
	if usableW < 1 {
		usableW = 1
	}
	status := renderMessageRail(m, usableW)
	header := renderCompactHeader(m, usableW)
	footerLine := renderCompactFooter(usableW)
	query := renderQueryRail(m, usableW)
	// Reserve the query rail only while an input is active; the rail
	// renders with its own border so we have to give it the rows it
	// needs (typically 3: top border + content + bottom border).
	plan.query = 0
	if query != "" {
		plan.query = compactQueryLines
	}
	// Always reserve the two-line message rail (a dim "ready" line when no
	// toast is active) so the rendered height stays stable across the
	// app's lifetime, matching the standard tier invariant.
	plan.message = compactMessageLines
	// Reserve at least one footer row (the navigation strip) so core
	// actions stay discoverable even when shortcut bar is empty.
	plan.footer = max(1, plan.footer)

	panelH := max(1, m.Viewport.Height-headerRows(plan)-footerRows(plan))
	if plan.panel > panelH {
		plan.panel = panelH
	}
	if plan.panel < 1 {
		plan.panel = 1
	}
	body := fitRailHeight(renderMiddlePanel(m, plan.panel, usableW), plan.panel)
	parts := []string{fitRailHeight(header, plan.header)}
	if status != "" {
		parts = append(parts, fitRailHeight(status, plan.message))
	} else {
		parts = append(parts, fitRailHeight(component.GetStyle("dim").Render(i18n.T("msg.ready")), plan.message))
	}
	if plan.query > 0 && query != "" {
		parts = append(parts, fitRailHeight(query, plan.query))
	}
	parts = append(parts, body)
	if footerLine != "" {
		parts = append(parts, fitRailHeight(footerLine, plan.footer))
	} else {
		parts = append(parts, strings.Repeat("\n", plan.footer-1))
	}
	return strings.Join(parts, "\n")
}

// headerRows returns the rail count occupied by the header + status + query
// rails (excluding the panel + footer). Used to recompute the panel when
// the query rail collapses dynamically.
func headerRows(p railPlan) int { return p.header + p.message + p.query }

// footerRows returns the rail count occupied by the footer rail.
func footerRows(p railPlan) int { return p.footer }

// renderCompactHeader produces a single-line "Engine: docker | Containers: 12"
// status row that replaces the multi-column header in compact mode.
func renderCompactHeader(m *state.AppModel, width int) string {
	engine := engineLabel(m)
	host := hostLabel(m)
	summary := summaryLabel(m)
	line := fmt.Sprintf("%s %s | %s | %dx%d", engine, host, summary, m.Viewport.Width, m.Viewport.Height)
	return fitRailHeight(component.GetStyle("headerBar").Render(component.TruncateVisible(line, width)), 1)
}

// renderCompactFooter produces a single-line shortcut strip with the most
// common actions. Other shortcuts remain available via the keymap and
// are documented in the help panel.
func renderCompactFooter(width int) string {
	short := []struct{ k, d string }{
		{keys.KTab, "Panel"}, {keys.KEnter, "Open"}, {keys.KEsc, "Back"}, {":", "Cmd"}, {keys.KF1, keys.ActionLabelHelp},
	}
	var parts []string
	for _, s := range short {
		parts = append(parts,
			component.GetStyle("hintKey").Render(s.k)+
				component.GetStyle("hintDesc").Render(" "+s.d))
	}
	line := strings.Join(parts, component.GetStyle("hintSep").Render(" "+component.BorderLineVertical+" "))
	line = component.PadVisible(component.TruncateVisible(line, width), width)
	return fitRailHeight(component.GetStyle("shortcutBar").Render(line), 1)
}

func engineLabel(m *state.AppModel) string {
	if m == nil {
		return "engine"
	}
	if t := m.Connection.RuntimeType; t != "" {
		return t
	}
	return "docker"
}

func hostLabel(m *state.AppModel) string {
	if m == nil {
		return "-"
	}
	if m.Connection.Pool != nil {
		if active := m.Connection.Pool.Active(); active != nil && active.Host != "" {
			return active.Host
		}
	}
	if m.Connection.ConnectionTarget != "" {
		return m.Connection.ConnectionTarget
	}
	return "-"
}

func summaryLabel(m *state.AppModel) string {
	if m == nil {
		return ""
	}
	counts := []string{}
	if n := len(m.Resources.Containers.Items); n > 0 {
		counts = append(counts, fmt.Sprintf("C:%d", n))
	}
	if n := len(m.Resources.Images.Items); n > 0 {
		counts = append(counts, fmt.Sprintf("I:%d", n))
	}
	if n := len(m.Resources.Volumes.Items); n > 0 {
		counts = append(counts, fmt.Sprintf("V:%d", n))
	}
	if n := len(m.Resources.Networks.Items); n > 0 {
		counts = append(counts, fmt.Sprintf("N:%d", n))
	}
	if len(counts) == 0 {
		return "empty"
	}
	return strings.Join(counts, " ")
}

// renderTerminalError replaces the entire viewport with a single, padded
// error block when the terminal is too small to host any usable layout.
// Kept short and ASCII-only so it stays readable even on 40x12.
func renderTerminalError(width, height int) string {
	msg := fmt.Sprintf("Terminal too small: %dx%d (need %dx%d)", width, height,
		minimumTerminalWidth, minimumTerminalHeight)
	if height >= 3 {
		hint := "Resize the window or press F1 for help."
		return lipgloss.PlaceVertical(height, lipgloss.Center,
			lipgloss.PlaceHorizontal(width, lipgloss.Center, msg+"\n"+hint))
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, msg)
}

// compile-time guard that footer.Render is still referenced; the compact
// fallback only uses a subset of the widget but we keep the import alive
// for parity with RenderApp.
var _ = footer.Render
