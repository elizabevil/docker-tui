package view

import (
	"github.com/elizabevil/docker-tui/internal/tui/state"
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
	return m.Viewport.Width - 1, rep.FooterTop, true
}

// activeTextInput reports whether any text-editing widget currently
// owns the keyboard.
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

// PanelRect describes the resolved top-row of the panel rail together
// with the body geometry available for rows (excluding border + title).
// Exported so the mouse sub-package can read panel geometry.
type PanelRect struct {
	PanelTop  int
	BodyTop   int
	BodyRows  int
	BodyLeft  int
	BodyWidth int
}

// LayoutReport captures the resolved rails + panel geometry for a given
// model. HitTest uses it to answer "what rail does this click belong to"
// in O(1).
type LayoutReport struct {
	Class         TerminalClass
	Panel         PanelRect
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
	plan := planFor(TerminalCompact, m.Viewport.Height)
	queryActive := queryKindFor(m) != queryNone
	if queryActive {
		plan.query = compactQueryLines
	}
	plan.message = 1
	plan.footer = max(1, plan.footer)
	panelH := max(m.Viewport.Height-(plan.header+plan.message+plan.query+plan.footer), 1)
	panelTop := plan.header + plan.message + plan.query
	bodyTop := panelTop + 2
	return LayoutReport{
		Class:         TerminalCompact,
		Panel:         PanelRect{PanelTop: panelTop, BodyTop: bodyTop, BodyRows: panelH - 2, BodyLeft: 2, BodyWidth: m.Viewport.Width - 4},
		HeaderBottom:  plan.header,
		MessageBottom: plan.header + plan.message,
		QueryBottom:   plan.header + plan.message + plan.query,
		FooterTop:     m.Viewport.Height - plan.footer,
	}
}

func resolveStandardLayout(m *state.AppModel) LayoutReport {
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
	bodyTop := panelTop + 2
	return LayoutReport{
		Class:         TerminalStandard,
		Panel:         PanelRect{PanelTop: panelTop, BodyTop: bodyTop, BodyRows: plan.panel - 2, BodyLeft: padH + 2, BodyWidth: usableW - 4},
		HeaderBottom:  headerBottom,
		MessageBottom: messageBottom,
		QueryBottom:   messageBottom + plan.query,
		FooterTop:     footerTop,
	}
}