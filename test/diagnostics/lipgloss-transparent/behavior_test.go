// Package transparent is a Phase-0 diagnostic for the color-transparency
// plan. Run with:
//
//	go test -v ./test/diagnostics/lipgloss-transparent/
//
// The cases are observational only (no asserts). See the package-level
// verdict below for the finding that invalidates the plan as written.
package transparent

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
)

// Phase-0 verdict (2026-08-05): lipgloss v2 does NOT honor A==0 as
// "transparent". Any NRGBA{A:0} is rendered as opaque black (Go's NRGBA.
// RGBA() premultiplies by alpha, zeroing the RGB channels — confirmed by
// the red-A:0 subtest painting (0,0,0) not (255,0,0)). The only safe
// sentinel is `nil` (see nil_background case: 1-byte output, no SGR).
// Phase 1 must therefore make ParseColor("transparent") return a typed
// sentinel that callers detect via IsTransparent and substitute with nil
// before passing to lipgloss — NOT a color.Color.

func dumpBytes(t *testing.T, label, rendered string) {
	t.Logf("%s\n  raw : %q\n  size: %d bytes", label, rendered, len(rendered))
}

func TestLipglossTransparentBehavior(t *testing.T) {
	red := color.NRGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
	transBlack := color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00}
	transRed := color.NRGBA{R: 0xff, G: 0x00, B: 0x00, A: 0x00}

	t.Run("control_opaque_background", func(t *testing.T) {
		dumpBytes(t, "A:255 Background (control)",
			lipgloss.NewStyle().Background(red).Render("x"))
	})

	t.Run("transparent_black_background", func(t *testing.T) {
		dumpBytes(t, "A:0 Background, RGB=(0,0,0)",
			lipgloss.NewStyle().Background(transBlack).Render("x"))
	})

	t.Run("transparent_red_background", func(t *testing.T) {
		dumpBytes(t, "A:0 Background, RGB=(255,0,0) — footgun demo",
			lipgloss.NewStyle().Background(transRed).Render("x"))
	})

	t.Run("transparent_foreground", func(t *testing.T) {
		dumpBytes(t, "A:0 Foreground",
			lipgloss.NewStyle().Foreground(transBlack).Render("x"))
	})

	t.Run("nil_background", func(t *testing.T) {
		dumpBytes(t, "nil Background (safe sentinel)",
			lipgloss.NewStyle().Background(nil).Render("x"))
	})

	t.Run("transparent_background_with_width_padding", func(t *testing.T) {
		dumpBytes(t, "A:0 Background + Width(8) (padding path)",
			lipgloss.NewStyle().Background(transBlack).Width(8).Render("x"))
	})
}