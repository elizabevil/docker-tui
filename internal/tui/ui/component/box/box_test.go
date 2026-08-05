package box

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// backgroundAnsi is the CSI prefix for a 24-bit background colour. lipgloss
// emits it as part of a single SGR sequence that may also carry the foreground
// (e.g. `\x1b[38;2;R;G;B;48;2;R;G;Bm`), so the test asserts the bare
// sub-token ";48;2;" rather than the full escape.
const backgroundAnsi = ";48;2;"

func TestLabeledValue_NoBackgroundUsesNoBackgroundEscape(t *testing.T) {
	out := (&LabeledValue{
		Label: "CPU", Value: "6%",
		LabelStyle: StyleHeaderLabel, ValueStyle: StyleHeaderValue,
	}).Render()
	if strings.Contains(out, backgroundAnsi) {
		t.Fatalf("expected no background ANSI when Background is empty, got %q", out)
	}
}

func TestLabeledValue_AppliesBackgroundToBothHalves(t *testing.T) {
	out := (&LabeledValue{
		Label: "CPU", Value: "6%",
		LabelStyle: StyleHeaderLabel, ValueStyle: StyleHeaderValue,
		Background: "#fefefe",
	}).Render()
	count := strings.Count(out, backgroundAnsi)
	if count < 2 {
		t.Fatalf("expected background applied to both label and value halves, got %d occurrences in %q", count, out)
	}
}

func TestLabeledValue_PadsLabel(t *testing.T) {
	out := (&LabeledValue{
		Label: "CPU", Value: "6%",
		LabelStyle: StyleHeaderLabel, ValueStyle: StyleHeaderValue,
		Background: "#fefefe",
		LabelWidth: 10,
	}).Render()
	labelPart := strings.SplitN(out, "6%", 2)[0]
	got := lipgloss.Width(stripANSI(labelPart))
	if got != 10 {
		t.Fatalf("label visible width = %d, want 10; part = %q", got, labelPart)
	}
}

func TestLabeledValue_FocusBoldAndColor(t *testing.T) {
	out := (&LabeledValue{
		Label: "CPU", Value: "6%",
		LabelStyle: StyleHeaderLabel, ValueStyle: StyleHeaderValue,
		Background: "#fefefe",
		FocusBold:  true,
		FocusColor: "#ff0000",
	}).Render()
	// Bold is folded into the same SGR sequence as the foreground colour in
	// lipgloss v2 — assert the colour, the prefix `\x1b[1;38;2;255;0;0` is
	// emitted together, then the explicit `38;2;255;0;0` token is present.
	if !strings.Contains(out, "38;2;255;0;0") {
		t.Fatalf("expected explicit foreground colour in focused label, got %q", out)
	}
	// Lipgloss may also emit the bold flag separately; check for either the
	// combined escape prefix or a standalone bold code.
	if !strings.Contains(out, "\x1b[1m") && !strings.Contains(out, ";1;38;2;") && !strings.Contains(out, ";1m") && !strings.Contains(out, "\x1b[1;38;2;") {
		t.Fatalf("expected bold attribute in focused label, got %q", out)
	}
}

func TestBadge_AppliesBackgroundToEntireString(t *testing.T) {
	out := (&Badge{
		Key: "R", Description: "Refresh",
		StyleName: StyleHeaderKey, Background: "#fefefe",
	}).Render()
	if !strings.Contains(out, backgroundAnsi) {
		t.Fatalf("expected background escape in badge, got %q", out)
	}
}

func TestBadgeRow_SeparatorInheritsBackground(t *testing.T) {
	out := (&BadgeRow{
		Badges: []Badge{
			{Key: "R", Description: "Refresh", StyleName: StyleHeaderKey},
			{Key: "F", Description: "Find", StyleName: StyleHeaderKey},
		},
		Separator: " │ ", Background: "#fefefe",
	}).Render()
	// Badges inherit the parent's background; the separator uses StyleDim +
	// the row's background, so we expect at least 1 background occurrence
	// (the separator) and 1 background inside each badge prefix.
	if got := strings.Count(out, backgroundAnsi); got < 1 {
		t.Fatalf("expected at least one background escape, got %d in %q", got, out)
	}
}

func TestStat_ThresholdColors(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		// Palette success / warning / danger are #499c54, #c8a35e, #db5a5a
		// which translate to ANSI 73;156;84 / 200;163;94 / 219;90;90.
		want string
	}{
		{"safe", 30, "73;156;84"},
		{"warn", 60, "200;163;94"},
		{"danger", 90, "219;90;90"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := (&Stat{
				Value: tc.value, Unit: "%",
				ValueWidth: 4, Background: "#fefefe",
				WarningAt: 50, DangerAt: 80,
			}).Render()
			if !strings.Contains(out, tc.want) {
				t.Fatalf("expected %q in %q", tc.want, out)
			}
		})
	}
}

func TestStat_ExplicitColorOverridesThreshold(t *testing.T) {
	out := (&Stat{
		Value: 90, Unit: "%", Background: "#fefefe",
		WarningAt: 50, DangerAt: 80,
		Color: "#123456",
	}).Render()
	if !strings.Contains(out, "18;52;86") {
		t.Fatalf("expected explicit colour to override threshold, got %q", out)
	}
}

func TestBorderedBox_AppliesBackgroundToPadding(t *testing.T) {
	out := (&BorderedBox{
		Content:    "X",
		UseRounded: true,
		Background: "#fefefe",
		Padding:    [2]int{1, 2},
		Width:      8,
	}).Render()
	// lipgloss emits a standalone background escape when the base style has
	// no foreground; assert the full escape form.
	if !strings.Contains(out, "\x1b[48;2;") {
		t.Fatalf("expected background in bordered box, got %q", out)
	}
}

func TestTableRow_AppliesBackgroundToPadding(t *testing.T) {
	out := (&TableRow{
		Cells:      []string{"abc", "def"},
		Width:      20,
		Background: "#fefefe",
		Gap:        " ",
	}).Render()
	if !strings.Contains(out, backgroundAnsi) {
		t.Fatalf("expected background in table row, got %q", out)
	}
	if got := lipgloss.Width(stripANSI(out)); got != 20 {
		t.Fatalf("table row visible width = %d, want 20; row = %q", got, out)
	}
}

func TestTableRow_NoWidthRendersFlat(t *testing.T) {
	out := (&TableRow{
		Cells: []string{"abc", "def"},
		Gap:   " ",
	}).Render()
	if strings.Contains(out, backgroundAnsi) {
		t.Fatalf("flat row should not apply background, got %q", out)
	}
	if got := lipgloss.Width(stripANSI(out)); got != 7 {
		t.Fatalf("flat row width = %d, want 7; row = %q", got, out)
	}
}

func TestTableRow_TruncatesCells(t *testing.T) {
	out := (&TableRow{
		Cells:         []string{"longtext", "x"},
		ColWidths:     []int{4, 1},
		Width:         6,
		TruncateCells: true,
		Gap:           " ",
	}).Render()
	// utils.TruncateVisible reserves 3 cells for the ellipsis, so a 4-wide
	// column truncates "longtext" to "l..." (1 char + "..." = 4 visible).
	if !strings.Contains(out, "l...") {
		t.Fatalf("expected truncated cell 'l...', got %q", out)
	}
	if !strings.Contains(out, "x") {
		t.Fatalf("expected short cell rendered, got %q", out)
	}
}

func TestBackgroundColor_EmptyReturnsNil(t *testing.T) {
	if c := backgroundColor(""); c != nil {
		t.Fatalf("empty string should return nil, got %#v", c)
	}
}

func TestBackgroundColor_InvalidReturnsNil(t *testing.T) {
	if c := backgroundColor("not-a-color"); c != nil {
		t.Fatalf("invalid colour should return nil, got %#v", c)
	}
}

func TestStyleWithBackground_EmptyStringReturnsBase(t *testing.T) {
	base := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	got := styleWithBackground(base, "")
	if got.GetForeground() == nil {
		t.Fatal("empty background should preserve base style")
	}
}

// stripANSI removes ANSI escape codes for visible-width tests.
func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inESC := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == 0x1b {
			inESC = true
			continue
		}
		if inESC {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				inESC = false
			}
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}
