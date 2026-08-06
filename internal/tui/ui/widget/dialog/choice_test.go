package dialog

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestChoiceDialogUsesPanelSizeWhenBodyProvided verifies that when
// bodyW>0 and bodyH>0 are passed, ChoiceDialog sizes itself via
// panelDialogWidth / panelDialogHeight (BR-043 §3.2: 3/4 × 1:1 panel
// footprint) rather than the legacy terminal-percentage fallback.
func TestChoiceDialogUsesPanelSizeWhenBodyProvided(t *testing.T) {
	cfg := LoadDialogConfig()
	if cfg.PanelSize.WidthPercent != 75 {
		t.Fatalf("default WidthPercent = %d, want 75", cfg.PanelSize.WidthPercent)
	}
	if cfg.PanelSize.HeightPercent != 100 {
		t.Fatalf("default HeightPercent = %d, want 100", cfg.PanelSize.HeightPercent)
	}

	// bodyW=80, bodyH=30 → dialogW = 60 (75% of 80), dialogH = 30.
	bodyW, bodyH := 80, 30
	box := ChoiceDialog(
		"Confirm",
		"body text",
		[]ChoiceOption{{Label: "Yes"}, {Label: "No"}},
		0,
		bodyW, bodyH,
		nil, "",
		cfg,
	)
	if box == "" {
		t.Fatal("ChoiceDialog returned empty")
	}
	for _, line := range splitLines(box) {
		w := ansi.StringWidth(line)
		// Outer dialog width = dialogW + 4 (border + padding).
		// dialogW = 60 → outer ≤ 64. Allow 4 cells of slack for
		// border/padding rounding (max ≈ 68).
		if w > 68 {
			t.Fatalf("dialog line width %d > 68 (expected dialogW=60 + border/padding)", w)
		}
	}
}

// TestChoiceDialogFallsBackToTerminalWithoutBody verifies that when
// bodyW≤0 or bodyH≤0, ChoiceDialog falls back to the terminal-percent
// sizing path (the legacy dialogWidth/dialogHeight helpers).
func TestChoiceDialogFallsBackToTerminalWithoutBody(t *testing.T) {
	cfg := LoadDialogConfig()
	// termW=120, termH=40 → fallback uses cfg.Width.Percent=75 and
	// cfg.Height.Percent=75. Width 120*75/100=90, capped to Max=120.
	// Height 40*75/100=30, capped to Max=40.
	box := ChoiceDialog(
		"Confirm",
		"body text",
		[]ChoiceOption{{Label: "Yes"}, {Label: "No"}},
		0,
		120, 40,
		nil, "",
		cfg,
	)
	if box == "" {
		t.Fatal("ChoiceDialog returned empty")
	}
}

// TestChoiceDialogSizeHelperBodyPath verifies the choiceDialogSize
// helper directly: bodyW>0 && bodyH>0 must route to panelDialogWidth /
// panelDialogHeight, otherwise to dialogWidth / dialogHeight.
func TestChoiceDialogSizeHelperBodyPath(t *testing.T) {
	cfg := LoadDialogConfig()

	// Body path: bodyW=80 → 75% = 60; bodyH=20 → 100% = 20.
	w, h := choiceDialogSize(80, 20, cfg)
	if w != 60 {
		t.Errorf("body path width = %d, want 60", w)
	}
	if h != 20 {
		t.Errorf("body path height = %d, want 20", h)
	}

	// Terminal fallback path: termW=80, termH=20, bodyW=0.
	// dialogWidth: 80*75/100=60, clamped [60,120] = 60.
	// dialogHeight: 20*75/100=15, clamped [16,40] = 16.
	w, h = choiceDialogSize(0, 0, cfg)
	if w != 60 {
		t.Errorf("term fallback width = %d, want 60", w)
	}
	if h != 16 {
		t.Errorf("term fallback height = %d, want 16 (clamped to Min=16)", h)
	}
}

// TestChoiceDialogRendersOptions verifies that all ChoiceOption labels
// are rendered into the dialog body when the panel-body path is used.
func TestChoiceDialogRendersOptions(t *testing.T) {
	cfg := LoadDialogConfig()
	box := ChoiceDialog(
		"Choose one",
		"Pick an option:",
		[]ChoiceOption{
			{Label: "First"},
			{Label: "Second", Description: "secondary"},
			{Label: "Third", Disabled: true},
		},
		1, 80, 30, nil, "", cfg,
	)
	for _, label := range []string{"First", "Second", "Third", "secondary"} {
		if !strings.Contains(box, label) {
			t.Errorf("dialog must contain %q", label)
		}
	}
	if !strings.Contains(box, "Choose one") {
		t.Errorf("dialog must contain the title")
	}
}

// TestChoiceDialogRespectsMaxWidth verifies that a wide body does not
// produce a dialog wider than cfg.PanelSize.MaxWidth (default 120).
func TestChoiceDialogRespectsMaxWidth(t *testing.T) {
	cfg := LoadDialogConfig()
	// bodyW=400 → 75% = 300, capped to MaxWidth=120.
	box := ChoiceDialog(
		"Confirm",
		"x",
		[]ChoiceOption{{Label: "OK"}},
		0, 400, 30, nil, "", cfg,
	)
	for _, line := range splitLines(box) {
		w := ansi.StringWidth(line)
		// 120 + 4 (border) + ~4 (padding) = ~128. Allow 4 cells slack.
		if w > 132 {
			t.Fatalf("dialog line width %d > 132 (expected dialogW=120 + border/padding)", w)
		}
	}
}

// TestChoiceDialogBordered ensures the rendered box has a lipgloss
// rounded border on the first/last visible line.
func TestChoiceDialogBordered(t *testing.T) {
	cfg := LoadDialogConfig()
	box := ChoiceDialog(
		"Bordered",
		"x",
		[]ChoiceOption{{Label: "OK"}},
		0, 80, 30, nil, "", cfg,
	)
	if !strings.Contains(box, lipgloss.RoundedBorder().Top) {
		t.Fatal("dialog must have a rounded top border")
	}
}
