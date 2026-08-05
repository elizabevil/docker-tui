package style

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
)

// noColorSentinel is the value returned by lipgloss v2's GetForeground /
// GetBackground when a channel is unset (per lipgloss source, the package
// keeps `var noColor = NoColor{}` and returns it for the unset path).
// Using a package-level var sidesteps a Go 1.26 gofmt/vet parser quirk
// when `lipgloss.NoColor{}` appears as an inline comparison RHS in tests.
var noColorSentinel = lipgloss.NoColor{}

// TestApplyForegroundSkipsTransparent asserts that a fully-transparent color
// (α == 0) is dropped on the foreground channel rather than forwarded to
// lipgloss — Phase 0 evidence shows lipgloss renders NRBGA{A:0} as opaque
// black, defeating the "transparent means no draw" contract.
func TestApplyForegroundSkipsTransparent(t *testing.T) {
	// Given a base style and a fully-transparent NRGBA color.
	base := lipgloss.NewStyle()
	transparent := color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0x00}

	// When ApplyForeground is called.
	got := ApplyForeground(base, transparent)

	// Then the foreground channel must remain unset.
	if fg := got.GetForeground(); fg != noColorSentinel {
		t.Fatalf("expected foreground to be unset (lipgloss.NoColor{}), got %T(%v)", fg, fg)
	}
}

// TestApplyForegroundPassesOpaque asserts that an opaque color is forwarded
// to lipgloss untouched, so the existing happy path is not regressed.
func TestApplyForegroundPassesOpaque(t *testing.T) {
	// Given an opaque NRGBA color.
	base := lipgloss.NewStyle()
	opaque := color.NRGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff}

	// When ApplyForeground is called.
	got := ApplyForeground(base, opaque)

	// Then the foreground channel equals the input color.
	if fg := got.GetForeground(); fg != opaque {
		t.Fatalf("expected foreground to equal input, got %T(%v)", fg, fg)
	}
}

// TestApplyForegroundPassesHalfAlpha asserts that a non-zero alpha is NOT
// skipped — defensive coverage so we do not accidentally drop half-alpha
// colors at the call site when buildStyle's flatten path is bypassed.
func TestApplyForegroundPassesHalfAlpha(t *testing.T) {
	// Given a half-alpha NRGBA color.
	base := lipgloss.NewStyle()
	half := color.NRGBA{R: 0x80, G: 0x40, B: 0x20, A: 0x80}

	// When ApplyForeground is called.
	got := ApplyForeground(base, half)

	// Then the foreground channel is set (not skipped).
	if fg := got.GetForeground(); fg == noColorSentinel {
		t.Fatalf("expected half-alpha foreground to be set, got unset sentinel")
	}
}

// TestApplyForegroundPassesNil asserts that a nil color is dropped safely
// without panicking. color.NRGBAModel.Convert panics on nil, so the helper
// must short-circuit before that call.
func TestApplyForegroundPassesNil(t *testing.T) {
	// Given a nil color.
	base := lipgloss.NewStyle()

	// When ApplyForeground is called (must not panic).
	got := ApplyForeground(base, nil)

	// Then the foreground channel remains unset.
	if fg := got.GetForeground(); fg != noColorSentinel {
		t.Fatalf("expected nil foreground to leave channel unset, got %T(%v)", fg, fg)
	}
}

// TestApplyBackgroundSkipsTransparent is the background counterpart of
// TestApplyForegroundSkipsTransparent.
func TestApplyBackgroundSkipsTransparent(t *testing.T) {
	base := lipgloss.NewStyle()
	transparent := color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0x00}

	got := ApplyBackground(base, transparent)

	if bg := got.GetBackground(); bg != noColorSentinel {
		t.Fatalf("expected background to be unset (lipgloss.NoColor{}), got %T(%v)", bg, bg)
	}
}

// TestApplyBackgroundPassesOpaque is the background counterpart of
// TestApplyForegroundPassesOpaque.
func TestApplyBackgroundPassesOpaque(t *testing.T) {
	base := lipgloss.NewStyle()
	opaque := color.NRGBA{R: 0x00, G: 0xff, B: 0x00, A: 0xff}

	got := ApplyBackground(base, opaque)

	if bg := got.GetBackground(); bg != opaque {
		t.Fatalf("expected background to equal input, got %T(%v)", bg, bg)
	}
}

// TestApplyBackgroundPassesHalfAlpha is the background counterpart of
// TestApplyForegroundPassesHalfAlpha.
func TestApplyBackgroundPassesHalfAlpha(t *testing.T) {
	base := lipgloss.NewStyle()
	half := color.NRGBA{R: 0x80, G: 0x40, B: 0x20, A: 0x80}

	got := ApplyBackground(base, half)

	if bg := got.GetBackground(); bg == noColorSentinel {
		t.Fatalf("expected half-alpha background to be set, got unset sentinel")
	}
}

// TestApplyBackgroundPassesNil is the background counterpart of
// TestApplyForegroundPassesNil.
func TestApplyBackgroundPassesNil(t *testing.T) {
	base := lipgloss.NewStyle()

	got := ApplyBackground(base, nil)

	if bg := got.GetBackground(); bg != noColorSentinel {
		t.Fatalf("expected nil background to leave channel unset, got %T(%v)", bg, bg)
	}
}
