// Standalone probe for lipgloss v2 + NRGBA{A:0} behavior.
// Run from the repo root:
//
//	go run ./test/diagnostics/lipgloss-transparent-probe/
//
// Prints the raw ANSI bytes lipgloss emits for each case plus a verdict.
package main

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"
)

func dump(label, rendered string) {
	fmt.Printf("%-46s raw=%q  size=%d\n", label, rendered, len(rendered))
}

func main() {
	red := color.NRGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}
	transBlack := color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00}
	transRed := color.NRGBA{R: 0xff, G: 0x00, B: 0x00, A: 0x00}

	fmt.Println("=== lipgloss v2 transparent probe ===")

	dump("control: opaque red background",
		lipgloss.NewStyle().Background(red).Render("x"))

	dump("A:0 background, RGB=(0,0,0)",
		lipgloss.NewStyle().Background(transBlack).Render("x"))

	dump("A:0 background, RGB=(255,0,0)",
		lipgloss.NewStyle().Background(transRed).Render("x"))

	dump("A:0 foreground",
		lipgloss.NewStyle().Foreground(transBlack).Render("x"))

	dump("nil background (only safe sentinel)",
		lipgloss.NewStyle().Background(nil).Render("x"))

	dump("A:0 background + Width(8) padding",
		lipgloss.NewStyle().Background(transBlack).Width(8).Render("x"))

	fmt.Println()
	fmt.Println("verdict: lipgloss v2 does NOT honor A==0 as transparent.")
	fmt.Println("  NRGBA.A:0 renders as opaque black (Go's NRGBA.RGBA()")
	fmt.Println("  premultiplies by alpha, zeroing the RGB channels).")
	fmt.Println("  The only safe sentinel is `nil`.")
	fmt.Println("  Plan Phase 1 must return a typed sentinel, not NRGBA{A:0}.")
}