// Package box provides self-contained, component-level rendering primitives.
//
// Each component manages its own background as a property of the box itself
// rather than relying on character-level ANSI reset barriers. The "transparent"
// semantic is: a component with no background color let the terminal default
// show through; it does not rely on a parent wrapper to provide the fill.
package box

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/utils"
)

// StyleName is a stable identifier into the component package's style cache.
// Keep these names aligned with the keys in globalStyleRefs (styles_load.go).
type StyleName string

// Common style names. Components accept a StyleName so the caller can pick the
// right semantic role for the slice of content without coupling to a single
// hard-coded style.
const (
	StyleHeaderLabel StyleName = "headerLabel"
	StyleHeaderValue StyleName = "headerBar"
	StyleHeaderKey   StyleName = "keyBadge"
	StyleHeaderLast  StyleName = "keyLast"
	StylePanelTitle  StyleName = "panelTitle"
	StyleDim         StyleName = "dim"
	StyleRowNormal   StyleName = "rowNormal" // reserved; falls back to safe fallback
	StyleRowSelected StyleName = "selectedRow"
	StyleRowMarked   StyleName = "marked"
	StyleRowAlt      StyleName = "alt"
)

// backgroundColor returns the parsed color or nil when bg is empty.
func backgroundColor(bg string) color.Color {
	if bg == "" {
		return nil
	}
	if c, ok := utils.ParseColor(bg); ok {
		return c
	}
	return nil
}

// styleWithBackground returns the style with an optional background applied.
// Centralised so components never inject background codes at the character
// level — the background is part of the box's lipgloss.Style.
func styleWithBackground(base lipgloss.Style, bg string) lipgloss.Style {
	if bg == "" {
		return base
	}
	if c := backgroundColor(bg); c != nil {
		return base.Background(c)
	}
	return base
}
