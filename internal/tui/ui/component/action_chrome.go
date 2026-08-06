package component

import (
	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

// WrapActionWindow paints the given content inside the action scope's
// window chrome (theme.action.<scope>.window). It is the single seam
// that lets a Page-mode or Async-mode body render inside a frame whose
// border + background are theme-driven.
//
// The wrapper uses a rounded border (matching chrome.borderKind's
// default) and the width passed in by the caller. width is treated as
// the *outer* width; the inner content gets 2 cells less to make room
// for the border.
//
// An unrecognised scope (or one the theme did not declare) returns the
// raw content so callers can fall back gracefully.
func WrapActionWindow(content string, scope config.OperationScope, width int) string {
	windowStyle, borderStyle, ok := actionWindowStyles(scope)
	if !ok {
		return content
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderStyle.GetForeground()).
		Background(windowStyle.GetBackground()).
		Width(max(0, width-2)).
		Padding(0, 1).
		Render(content)
}

// actionWindowStyles returns the (window, border) style pair for the
// given scope. ok reports whether the scope is one of the recognised
// values (Container / Image) — callers fall back to plain content
// when ok is false so unrecognised scopes never accidentally inherit
// another scope's chrome.
func actionWindowStyles(scope config.OperationScope) (lipgloss.Style, lipgloss.Style, bool) {
	switch scope {
	case config.OperationScopeContainer:
		return GetStyle(StyleActionContainerWindow), GetStyle(StyleActionContainerWindowBorder), true
	case config.OperationScopeImage:
		return GetStyle(StyleActionImageWindow), GetStyle(StyleActionImageWindowBorder), true
	}
	return lipgloss.Style{}, lipgloss.Style{}, false
}