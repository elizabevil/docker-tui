package dialog

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestCenterOnPanelNormal(t *testing.T) {
	body := PanelBody{Left: 30, Top: 6, Width: 40, Rows: 12}
	content := buildContent(80, 24, '.')
	dlg := buildDialog(20, 5, '#')

	got := CenterOnPanel(content, dlg, body, 80, 24, DialogConfig{})
	lines := strings.Split(got, "\n")
	if len(lines) != 24 {
		t.Fatalf("line count = %d, want 24", len(lines))
	}

	// startX = 30 + (40-20)/2 = 40
	// startY = 6 + (12-5)/2 = 9
	// 5 dialog rows occupy lines 9..13
	// Outside lines (0..8 and 14..23) must be untouched content.
	for _, lineIdx := range []int{0, 1, 2, 7, 8, 14, 15, 23} {
		if lines[lineIdx] != buildContentRow(80, '.') {
			t.Errorf("line %d was modified; got %q", lineIdx, lines[lineIdx])
		}
	}

	// Lines 9..13: 40 leading dots, dialog, 20 trailing dots.
	for dy := range 5 {
		leftPart := strings.Repeat(".", 40)
		dl := ansi.TruncateWc(dialogRows[dy], 20, "")
		if visW := ansi.StringWidth(dl); visW < 20 {
			dl += strings.Repeat(" ", 20-visW)
		}
		rightPart := strings.Repeat(".", 20)
		want := leftPart + dl + rightPart
		if ansi.Strip(lines[9+dy]) != want {
			t.Errorf("dialog line %d\n  got: %q\n want: %q", dy, lines[9+dy], want)
		}
	}
}

func TestCenterOnPanelEmptyDialog(t *testing.T) {
	body := PanelBody{Left: 10, Top: 5, Width: 50, Rows: 10}
	content := buildContent(80, 20, '.')
	got := CenterOnPanel(content, "", body, 80, 20, DialogConfig{})
	if got != content {
		t.Errorf("empty dialog should return content unchanged")
	}
}

func TestCenterOnPanelClampsWideDialog(t *testing.T) {
	body := PanelBody{Left: 10, Top: 5, Width: 20, Rows: 8}
	content := buildContent(80, 20, '.')
	dlg := buildDialog(60, 3, '#')

	got := CenterOnPanel(content, dlg, body, 80, 20, DialogConfig{})
	lines := strings.Split(got, "\n")

	for _, lineIdx := range []int{0, 6, 10, 19} {
		if lines[lineIdx] != buildContentRow(80, '.') {
			t.Errorf("line %d should be unchanged", lineIdx)
		}
	}
	for _, lineIdx := range []int{7, 8, 9} {
		leftPart := strings.Repeat(".", 10)
		dlPart := strings.Repeat("#", 20)
		rightPart := strings.Repeat(".", 50)
		want := leftPart + dlPart + rightPart
		if ansi.Strip(lines[lineIdx]) != want {
			t.Errorf("line %d\n  got: %q\n want: %q", lineIdx, lines[lineIdx], want)
		}
	}
}

func TestCenterOnPanelClampsTallDialog(t *testing.T) {
	body := PanelBody{Left: 0, Top: 2, Width: 20, Rows: 3}
	content := buildContent(40, 10, '.')
	dlg := buildDialog(10, 6, '#')

	got := CenterOnPanel(content, dlg, body, 40, 10, DialogConfig{})
	lines := strings.Split(got, "\n")

	// Effective rows = min(6, 3) = 3. So only 3 lines modified.
	// startY = 2 + (3-3)/2 = 2.
	for _, lineIdx := range []int{0, 1, 5, 6, 7, 8, 9} {
		if lines[lineIdx] != buildContentRow(40, '.') {
			t.Errorf("line %d should be unchanged; got %q", lineIdx, lines[lineIdx])
		}
	}
}

func TestCenterOnPanelPreservesANSI(t *testing.T) {
	body := PanelBody{Left: 0, Top: 0, Width: 20, Rows: 4}
	content := buildContent(20, 4, '.')
	// Dialog with ANSI color codes.
	dlg := "\x1b[31mred\x1b[0m\n" + // red "red"
		"\x1b[32mgreen\x1b[0m\n" +
		"\x1b[34mblue\x1b[0m"

	got := CenterOnPanel(content, dlg, body, 20, 4, DialogConfig{})
	lines := strings.Split(got, "\n")

	// startX = 0 + (20-5)/2 = 7. startY = 0 + (4-3)/2 = 0.
	// The dialog replaces cols 7..11 (5 cols wide), so cols 0..6 are dots
	// and cols 12..19 are dots.
	// The "red" text has ANSI reset codes around "red"; verify it's still there.
	wantPlainRows := []string{
		strings.Repeat(".", 7) + "red  " + strings.Repeat(".", 8),
		strings.Repeat(".", 7) + "green" + strings.Repeat(".", 8),
		strings.Repeat(".", 7) + "blue " + strings.Repeat(".", 8),
	}
	for dy := range 3 {
		if plain := ansi.Strip(lines[dy]); plain != wantPlainRows[dy] {
			t.Errorf("line %d\n  got: %q\n want: %q", dy, plain, wantPlainRows[dy])
		}
	}
	// Confirm ANSI codes survived
	if !strings.Contains(got, "\x1b[31m") {
		t.Errorf("red ANSI code lost")
	}
	if !strings.Contains(got, "\x1b[32m") {
		t.Errorf("green ANSI code lost")
	}
	if !strings.Contains(got, "\x1b[34m") {
		t.Errorf("blue ANSI code lost")
	}
}

func TestPlaceDialogInPanelDerivesSize(t *testing.T) {
	body := PanelBody{Left: 5, Top: 2, Width: 30, Rows: 8}
	// Content rows are unequal width on purpose.
	content := "row0-1234567890\nrow1-short\n" + buildContentRow(40, '.')
	dlg := buildDialog(10, 3, '#')

	got := PlaceDialogInPanel(content, dlg, body, DialogConfig{})
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("line count = %d, want 3", len(lines))
	}
	// termW derived from widest line = 40. termH = 3.
	// startX = 5 + (30-10)/2 = 15.
	// startY = 2 + (8-3)/2 = 4. But termH = 3 < 4+3, so the 3rd dialog row
	// is dropped. startY clamped to termH-3 = 0.
	if lines[0] == "" || lines[2] == "" {
		t.Errorf("empty lines in result")
	}
}

func TestCenterOnPanelClampsStartXY(t *testing.T) {
	// Body extends past terminal bounds: startY would be 18, termH = 20.
	// effH = 3, startY = 18 + (3-3)/2 = 18. OK.
	// But startX + effW = 5 + 30 + (40-30)/2 = 50. Within termW = 60.
	body := PanelBody{Left: 5, Top: 18, Width: 40, Rows: 5}
	content := buildContent(60, 20, '.')
	dlg := buildDialog(30, 3, '#')
	got := CenterOnPanel(content, dlg, body, 60, 20, DialogConfig{})
	lines := strings.Split(got, "\n")
	if len(lines) != 20 {
		t.Fatalf("line count = %d, want 20", len(lines))
	}
	// Lines 18, 19 should have dialog content.
	if lines[18] == buildContentRow(60, '.') {
		t.Errorf("line 18 should be modified")
	}
	if lines[19] == buildContentRow(60, '.') {
		t.Errorf("line 19 should be modified")
	}
	if lines[0] != buildContentRow(60, '.') {
		t.Errorf("line 0 should be unchanged")
	}
}

// ── test helpers ─────────────────────────────────────────────

var dialogRows []string

func buildDialog(w, h int, ch rune) string {
	dialogRows = make([]string, h)
	var sb strings.Builder
	for i := range h {
		row := strings.Repeat(string(ch), w)
		dialogRows[i] = row
		sb.WriteString(row)
		if i < h-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

func buildContent(w, h int, ch rune) string {
	lines := make([]string, h)
	for i := range h {
		lines[i] = buildContentRow(w, ch)
	}
	return strings.Join(lines, "\n")
}

func buildContentRow(w int, ch rune) string {
	return strings.Repeat(string(ch), w)
}
