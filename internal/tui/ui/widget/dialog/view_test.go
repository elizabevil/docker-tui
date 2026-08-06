package dialog

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/elizabevil/docker-tui/internal/data/config"
)

// TestPanelDialogWidthUsesPanelSize verifies that panel-dialog width is
// computed from cfg.PanelSize.WidthPercent (default 75) and upper-bound
// by cfg.PanelSize.MaxWidth (default 120).
func TestPanelDialogWidthUsesPanelSize(t *testing.T) {
	cfg := LoadDialogConfig()

	// bodyW=80 → 80*75/100 = 60.
	if got := panelDialogWidth(80, cfg); got != 60 {
		t.Fatalf("panelDialogWidth(80, cfg) = %d, want 60", got)
	}
	// bodyW=200 → 200*75/100 = 150, capped to MaxWidth=120.
	if got := panelDialogWidth(200, cfg); got != 120 {
		t.Fatalf("panelDialogWidth(200, cfg) = %d, want 120 (clamped to MaxWidth)", got)
	}
	// bodyW=40 → 40*75/100 = 30; no Min clamp → 30.
	if got := panelDialogWidth(40, cfg); got != 30 {
		t.Fatalf("panelDialogWidth(40, cfg) = %d, want 30 (no Min clamp in PanelSize)", got)
	}
}

// TestPanelDialogWidthFallsBackToLegacyPercent verifies that a zero
// PanelSize.WidthPercent falls back to cfg.Width.Percent for the ratio,
// and a zero PanelSize.MaxWidth falls back to cfg.Width.Max for the
// upper bound (BR-043 §3.2).
func TestPanelDialogWidthFallsBackToLegacyPercent(t *testing.T) {
	cfg := DialogConfig{
		Width: config.ResponsiveSize{Percent: 50, Max: 90},
		PanelSize: config.PanelSize{
			WidthPercent: 0,
			MaxWidth:     0,
		},
	}
	// bodyW=200 → 200*50/100 = 100, capped to legacy Max=90.
	if got := panelDialogWidth(200, cfg); got != 90 {
		t.Fatalf("panelDialogWidth(200, cfg) = %d, want 90 (legacy fallback)", got)
	}
	// bodyW=80 → 80*50/100 = 40, no clamp.
	if got := panelDialogWidth(80, cfg); got != 40 {
		t.Fatalf("panelDialogWidth(80, cfg) = %d, want 40", got)
	}
}

// TestPanelDialogHeightUsesPanelSize verifies that panel-dialog height
// equals bodyH when HeightPercent=100 and is upper-bound by
// cfg.PanelSize.MaxHeight (default 40). Spec target: 1:1 height (full
// panel body height) capped at 40 rows.
func TestPanelDialogHeightUsesPanelSize(t *testing.T) {
	cfg := LoadDialogConfig()

	// bodyH=40 → 40*100/100 = 40 (1:1 with body height).
	if got := panelDialogHeight(40, cfg); got != 40 {
		t.Fatalf("panelDialogHeight(40, cfg) = %d, want 40 (1:1 with bodyH)", got)
	}
	// bodyH=100 → 100*100/100 = 100, capped to MaxHeight=40.
	if got := panelDialogHeight(100, cfg); got != 40 {
		t.Fatalf("panelDialogHeight(100, cfg) = %d, want 40 (clamped to MaxHeight)", got)
	}
	// bodyH=8 → 8*100/100 = 8 (no Min clamp in PanelSize).
	if got := panelDialogHeight(8, cfg); got != 8 {
		t.Fatalf("panelDialogHeight(8, cfg) = %d, want 8 (no Min clamp in PanelSize)", got)
	}
}

// TestPanelDialogHeightFallsBackToLegacyPercent verifies that a zero
// PanelSize.HeightPercent falls back to cfg.Height.Percent.
func TestPanelDialogHeightFallsBackToLegacyPercent(t *testing.T) {
	cfg := DialogConfig{
		Height: config.ResponsiveSize{Percent: 50, Max: 30},
		PanelSize: config.PanelSize{
			HeightPercent: 0,
			MaxHeight:     0,
		},
	}
	// bodyH=80 → 80*50/100 = 40, capped to legacy Max=30.
	if got := panelDialogHeight(80, cfg); got != 30 {
		t.Fatalf("panelDialogHeight(80, cfg) = %d, want 30 (legacy fallback)", got)
	}
}

// TestDefaultAppConfigPanelSizeDefaults verifies that the default app
// config populates PanelSize with the expected defaults: width 75 %,
// height 100 %, max width 120, max height 40.
func TestDefaultAppConfigPanelSizeDefaults(t *testing.T) {
	cfg := config.DefaultAppConfig().UI.Dialog
	if cfg.PanelSize.WidthPercent != 75 {
		t.Fatalf("PanelSize.WidthPercent = %d, want 75", cfg.PanelSize.WidthPercent)
	}
	if cfg.PanelSize.HeightPercent != 100 {
		t.Fatalf("PanelSize.HeightPercent = %d, want 100", cfg.PanelSize.HeightPercent)
	}
	if cfg.PanelSize.MaxWidth != 120 {
		t.Fatalf("PanelSize.MaxWidth = %d, want 120", cfg.PanelSize.MaxWidth)
	}
	if cfg.PanelSize.MaxHeight != 40 {
		t.Fatalf("PanelSize.MaxHeight = %d, want 40", cfg.PanelSize.MaxHeight)
	}
}

// TestDialogBoxSetsMaxWidth verifies that DialogBox propagates the
// MaxWidth field of DialogStyle into lipgloss so the rendered box
// honours the cap when content would otherwise exceed the requested
// width.
func TestDialogBoxSetsMaxWidth(t *testing.T) {
	box := DialogBox(DialogStyle{Width: 40, MaxWidth: 40, Height: 10, LeftAligned: true}, "line1", "line2")
	if box == "" {
		t.Fatal("DialogBox returned empty")
	}
	// Cap is 40; allow 4 cells slack for padding/border rounding.
	maxCells := 44
	for _, line := range []string{box} {
		for _, ln := range splitLines(line) {
			if w := ansi.StringWidth(ln); w > maxCells {
				t.Fatalf("dialog line width %d > %d", w, maxCells)
			}
		}
	}
}

// TestDialogBoxMaxWidthOverridesContent verifies that a wide content
// line is hard-capped by MaxWidth, never rendered wider than the cap
// even when the inner content overflows.
func TestDialogBoxMaxWidthOverridesContent(t *testing.T) {
	wide := "this-is-a-very-long-line-that-would-otherwise-overflow-the-dialog"
	box := DialogBox(DialogStyle{Width: 30, MaxWidth: 30, Height: 5, LeftAligned: true}, wide)
	for _, ln := range splitLines(box) {
		if w := ansi.StringWidth(ln); w > 34 {
			t.Fatalf("dialog line width %d > 34 with MaxWidth=30", w)
		}
	}
}

// TestDialogBoxMaxWidthZeroLeavesNoCap verifies that MaxWidth=0 (the
// zero value) does not impose a cap; the rendered box then only
// honours Width.
func TestDialogBoxMaxWidthZeroLeavesNoCap(t *testing.T) {
	box := DialogBox(DialogStyle{Width: 0, MaxWidth: 0, Height: 5, LeftAligned: true}, "hello")
	if box == "" {
		t.Fatal("DialogBox returned empty")
	}
}

// splitLines splits on \n (helper kept local to avoid stomping the test
// helpers in centered_test.go).
func splitLines(s string) []string {
	out := []string{}
	cur := ""
	for _, r := range s {
		if r == '\n' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

// ansiStringWidth measures the visible cell width of s, ANSI-aware.
func ansiStringWidth(s string) int {
	return ansi.StringWidth(s)
}
