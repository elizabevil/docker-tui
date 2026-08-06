package dialog

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// stripANSI removes ANSI escape sequences for column-position assertions.
func stripANSI(s string) string { return ansi.Strip(s) }

// formTestModel returns a model with a Form open and a sized viewport.
func formTestModel(spec state.FormSpec) *state.AppModel {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Viewport.Width = 100
	m.Viewport.Height = 30
	m.Form.Open(spec)
	return m
}

func TestFormDialogTwoColumnLayout(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind:  state.FormContainerCopy,
		Title: "Copy file from container",
		Fields: []state.FormField{
			{Key: "source", Label: "Container path", Kind: state.FormText, Input: state.QueryInputState{Text: "/etc/app.conf"}},
			{Key: "destination", Label: "Local destination (tar)", Kind: state.FormPath, Input: state.QueryInputState{Text: "/tmp/backup.tar"}},
		},
	})
	out := FormDialog(m, "", LoadDialogConfig(), 0, 0)
	lines := strings.Split(out, "\n")

	titleSeen := false
	labelSeen := false
	for _, line := range lines {
		if strings.Contains(line, "Copy file from container") {
			titleSeen = true
		}
		if strings.Contains(line, "Local destination (tar)") && !strings.Contains(line, "Local destination (tar):") {
			labelSeen = true
		}
	}
	if !titleSeen {
		t.Fatal("dialog must render the form title")
	}
	if !labelSeen {
		t.Fatal("dialog must render the destination label without punctuation")
	}
}

func TestFormDialogSelectCollapsedWithMarker(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerUpdate,
		Fields: []state.FormField{
			{
				Key:     "restart",
				Label:   "Restart policy",
				Kind:    state.FormSelect,
				Options: []string{"unchanged", "no", "always"},
				Index:   2,
			},
		},
	})
	out := FormDialog(m, "", LoadDialogConfig(), 0, 0)
	if !strings.Contains(out, "always") {
		t.Fatal("collapsed select must show the selected option")
	}
	if !strings.Contains(out, component.TriangleDownSmall) {
		t.Fatal("select cell must carry a dropdown marker")
	}
}

func TestFormDialogFieldErrorAligned(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerCopy,
		Fields: []state.FormField{
			{Key: "source", Label: "Container path", Kind: state.FormText, Required: true, Error: "required"},
			{Key: "destination", Label: "Local destination", Kind: state.FormPath},
		},
	})
	out := FormDialog(m, "", LoadDialogConfig(), 0, 0)
	if !strings.Contains(out, "required") {
		t.Fatal("field error must be rendered below the field row")
	}
}

func TestFormDialogDoesNotHighlightButtonsWhileFieldFocused(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind:   state.FormContainerCommit,
		Fields: []state.FormField{{Key: "repository", Label: "Repository", Kind: state.FormText}},
	})
	m.Form.FieldFocus = 0
	out := stripANSI(FormDialog(m, "", LoadDialogConfig(), 0, 0))
	if strings.Contains(out, component.ButtonIndicator) {
		t.Fatalf("field focus leaked into button selection: %q", out)
	}
}

func TestEditableValueHiddenCursorPreservesWidth(t *testing.T) {
	field := state.FormField{Kind: state.FormText, Input: state.QueryInputState{Text: "demo", Cursor: 4}}
	visible := stripANSI(renderEditableValue(&field, true, 20, true))
	hidden := stripANSI(renderEditableValue(&field, true, 20, false))
	if lipgloss.Width(visible) != lipgloss.Width(hidden) {
		t.Fatalf("cursor blink changed width: visible=%q hidden=%q", visible, hidden)
	}
}

func TestRenderPathValueShortStaysSingleLine(t *testing.T) {
	field := state.FormField{Kind: state.FormPath, Input: state.NewQueryInput("/tmp/backup.tar")}
	got := stripANSI(renderEditableValue(&field, false, 24))
	if strings.Contains(got, "\n") {
		t.Fatalf("short path wrapped unexpectedly: %q", got)
	}
}

func TestRenderPathValueLongWraps(t *testing.T) {
	const width = 12
	path := "/tmp/very-long-backup-archive-2026.tar.gz"
	field := state.FormField{Kind: state.FormPath, Input: state.NewQueryInput(path)}
	got := stripANSI(renderEditableValue(&field, false, width))
	if lines := strings.Split(got, "\n"); len(lines) < 3 {
		t.Fatalf("long path rendered in %d lines, want at least 3: %q", len(lines), got)
	}
	if strings.ReplaceAll(got, "\n", "") != path {
		t.Fatalf("wrapped path was altered: got %q, want %q", got, path)
	}
	if strings.Contains(got, "…") {
		t.Fatalf("wrapped path must not be truncated: %q", got)
	}
}

func TestRenderPathValuePreservesTail(t *testing.T) {
	path := "/very/long/path/to/some/deep/backup-archive-2026.tar.gz"
	field := state.FormField{Kind: state.FormPath, Input: state.NewQueryInput(path)}
	got := stripANSI(renderEditableValue(&field, false, 16))
	if !strings.Contains(strings.ReplaceAll(got, "\n", ""), "backup-archive-2026.tar.gz") {
		t.Fatalf("wrapped path lost its tail: %q", got)
	}
}

func TestRenderPathValueFocusedKeepsCursorWindow(t *testing.T) {
	path := "/very/long/path/to/some/deep/backup-archive-2026.tar.gz"
	field := state.FormField{Kind: state.FormPath, Input: state.NewQueryInput(path)}
	got := stripANSI(renderEditableValue(&field, true, 16))
	if strings.Contains(got, "\n") {
		t.Fatalf("focused path must remain a single-line cursor window: %q", got)
	}
	if utils.DisplayWidth(got) > 16 {
		t.Fatalf("focused path width = %d, want <= 16: %q", utils.DisplayWidth(got), got)
	}
}

func TestRenderPathValueRespectsMaxWidth(t *testing.T) {
	const width = 10
	path := "/very/long/path/to/some/deep/backup-archive-2026.tar.gz"
	field := state.FormField{Kind: state.FormPath, Input: state.NewQueryInput(path)}
	got := stripANSI(renderEditableValue(&field, false, width))
	for _, line := range strings.Split(got, "\n") {
		if lineWidth := utils.DisplayWidth(line); lineWidth > width {
			t.Fatalf("wrapped line width = %d, want <= %d: %q", lineWidth, width, line)
		}
	}
}

func TestEditableTextAndIntRemainSingleLineTruncated(t *testing.T) {
	for _, kind := range []state.FormFieldKind{state.FormText, state.FormInt} {
		field := state.FormField{Kind: kind, Input: state.NewQueryInput("12345678901234567890")}
		got := stripANSI(renderEditableValue(&field, false, 8))
		if strings.Contains(got, "\n") {
			t.Fatalf("kind %v wrapped unexpectedly: %q", kind, got)
		}
		if utils.DisplayWidth(got) > 8 {
			t.Fatalf("kind %v width = %d, want <= 8: %q", kind, utils.DisplayWidth(got), got)
		}
	}
}

func TestFormLayoutWidthsOutsideIn(t *testing.T) {
	fields := []state.FormField{
		{Label: "Container path"},
		{Label: "Local destination (tar)"},
		{Label: "Max retries"},
	}
	labelW, valueW := formLayout(fields, 40)
	if labelW <= 0 || valueW <= 0 || labelW+valueW+1 != 40 {
		t.Fatalf("layout must fill inner width exactly: label=%d value=%d sum=%d", labelW, valueW, labelW+valueW+1)
	}
	// Long labels are capped so the value column keeps a usable share.
	if labelW > 20 {
		t.Fatalf("label column too wide: %d", labelW)
	}
}

func TestRenderFormPopupSelectRows(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerUpdate,
		Fields: []state.FormField{
			{
				Key:     "restart",
				Label:   "Restart policy",
				Kind:    state.FormSelect,
				Options: []string{"no", "always", "on-failure"},
			},
		},
	})
	m.Form.FieldFocus = 0
	m.Form.OpenPopup()
	out := renderFormPopup(m.Form, "box", 60, 20)
	if !strings.Contains(out, "no") || !strings.Contains(out, "always") || !strings.Contains(out, "on-failure") {
		t.Fatal("popup must list every select option")
	}
}

func TestRenderFormPopupMultiSelectMarks(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerUpdate,
		Fields: []state.FormField{
			{
				Key:      "multi",
				Label:    "Multi",
				Kind:     state.FormMultiSelect,
				Options:  []string{"a", "b", "c"},
				Selected: map[string]bool{"b": true},
			},
		},
	})
	m.Form.FieldFocus = 0
	m.Form.OpenPopup()
	out := renderFormPopup(m.Form, "box", 60, 20)
	if !strings.Contains(out, "[x] b") {
		t.Fatal("checked multi option must carry an [x] marker")
	}
}

func TestRenderFormPopupPathRows(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerExport,
		Fields: []state.FormField{
			{
				Key:     "destination",
				Label:   "Local destination",
				Kind:    state.FormPath,
				Options: nil,
				Suggestions: []state.PathEntry{
					{Name: "backup.tar", Path: "/tmp/backup.tar", IsDir: false, Type: state.PathEntryFile},
					{Name: "sub", Path: "/tmp/sub", IsDir: true, Type: state.PathEntryDir},
				},
			},
		},
	})
	m.Form.FieldFocus = 0
	m.Form.OpenPopup()
	out := renderFormPopup(m.Form, "box", 100, 20)
	if !strings.Contains(out, "backup.tar") {
		t.Fatal("popup must list path candidates")
	}
	if !strings.Contains(out, "sub/") {
		t.Fatal("directory candidates must carry a trailing separator")
	}
}

func TestPathPopupGeometryStaysFixedAcrossDirectoryLoads(t *testing.T) {
	render := func(suggestions []state.PathEntry, loading bool) string {
		m := formTestModel(state.FormSpec{Kind: state.FormContainerCopy, Fields: []state.FormField{{
			Key: "source", Label: "Container path", Kind: state.FormPath,
			PathSource: state.PathContainer, Suggestions: suggestions, PathLoading: loading,
		}}})
		m.Form.FieldFocus = 0
		m.Form.OpenPopup()
		return renderFormPopup(m.Form, "box", 100, 20)
	}
	one := render([]state.PathEntry{{Name: "apk", Path: "/etc/apk", IsDir: true, Type: state.PathEntryDir}}, false)
	many := render([]state.PathEntry{
		{Name: "a", Path: "/etc/a", IsDir: true, Type: state.PathEntryDir},
		{Name: "b", Path: "/etc/b", IsDir: true, Type: state.PathEntryDir},
		{Name: "c", Path: "/etc/c", Type: state.PathEntryFile},
		{Name: "d", Path: "/etc/d", Type: state.PathEntryFile},
	}, false)
	loading := render(nil, true)
	wantLines := len(strings.Split(one, "\n"))
	if got := len(strings.Split(many, "\n")); got != wantLines {
		t.Fatalf("many-candidate popup height = %d, want %d", got, wantLines)
	}
	if got := len(strings.Split(loading, "\n")); got != wantLines {
		t.Fatalf("loading popup height = %d, want %d", got, wantLines)
	}
	if lipgloss.Width(one) != lipgloss.Width(many) || lipgloss.Width(one) != lipgloss.Width(loading) {
		t.Fatalf("popup width changed: one=%d many=%d loading=%d", lipgloss.Width(one), lipgloss.Width(many), lipgloss.Width(loading))
	}
}

func TestFormDialogNoPopupWhenClosed(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind:   state.FormContainerExport,
		Fields: []state.FormField{{Key: "destination", Label: "Local destination", Kind: state.FormPath}},
	})
	out := FormDialog(m, "", LoadDialogConfig(), 0, 0)
	if strings.Contains(out, ">") && !strings.Contains(out, component.ButtonIndicator) {
		t.Fatal("closed popup must not render cursor rows")
	}
}

// TestFormDialogUnfocusedPathAnchoredAtStart verifies that when a path
// field is not focused (e.g. immediately after the form opens with a
// default Local destination tar path) the rendered text is anchored at
// the start so the user can see which directory the file will land in,
// instead of being scrolled to the cursor at the end of the path.
func TestFormDialogUnfocusedPathAnchoredAtStart(t *testing.T) {
	long := "/home/u/p/snapshot-2026-08-05-snapshot-1234.tar"
	m := formTestModel(state.FormSpec{
		Kind:   state.FormContainerExport,
		Fields: []state.FormField{{Key: "destination", Label: "Local destination", Kind: state.FormPath, Input: state.QueryInputState{Text: long, Cursor: len(long)}}},
	})
	// Move focus to a sibling text field so the path field is unfocused.
	m.Form.Fields = append(m.Form.Fields, state.FormField{Key: "buffer", Label: "Buffer", Kind: state.FormText})
	m.Form.FieldFocus = 1
	out := stripANSI(FormDialog(m, "", LoadDialogConfig(), 0, 0))
	if !strings.Contains(out, "/home/u/p") {
		t.Fatalf("unfocused long path must show leading directory, got %q", out)
	}
}

// TestFormDialogSelectPopupRendersInline verifies that pressing Enter on a
// FormSelect / FormMultiSelect field shows the option list as a dropdown
// inside the same dialog rather than replacing the form with a bordered
// popup window. The dropdown is recognised by every option text appearing
// directly in FormDialog output (not requiring a separate DialogBox render).
func TestFormDialogSelectPopupRendersInline(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerUpdate,
		Fields: []state.FormField{
			{Key: "memory", Label: "Memory", Kind: state.FormInt},
			{
				Key:     "restart",
				Label:   "Restart policy",
				Kind:    state.FormSelect,
				Options: []string{"unchanged", "always", "on-failure", "unless-stopped", "no"},
			},
		},
	})
	m.Form.FieldFocus = 1
	m.Form.OpenPopup()
	out := FormDialog(m, "", LoadDialogConfig(), 0, 0)
	if !strings.Contains(out, "always") || !strings.Contains(out, "on-failure") {
		t.Fatalf("inline dropdown must include every option, got %q", out)
	}
	if !strings.Contains(out, "Memory") {
		t.Fatalf("inline dropdown must keep prior field rows visible, got %q", out)
	}
}

// TestPathPopupTypeColumnAligned verifies that every path row starts at the
// same column and that directories carry the [DIR] marker while files carry
// the [FILE] marker (BR-041 §4.4).
func TestPathPopupTypeColumnAligned(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerExport,
		Fields: []state.FormField{
			{
				Key:  "destination",
				Kind: state.FormPath,
				Suggestions: []state.PathEntry{
					{Name: "alpha.tar", Path: "/tmp/alpha.tar", IsDir: false, Type: state.PathEntryFile},
					{Name: "beta", Path: "/tmp/beta", IsDir: true, Type: state.PathEntryDir},
					{Name: "gamma", Path: "/tmp/gamma", IsDir: true, Type: state.PathEntryDir},
					{Name: "delta.tar", Path: "/tmp/delta.tar", IsDir: false, Type: state.PathEntryFile},
				},
			},
		},
	})
	m.Form.FieldFocus = 0
	m.Form.OpenPopup()
	out := renderFormPopup(m.Form, "box", 100, 20)
	clean := stripANSI(out)
	lines := strings.Split(clean, "\n")
	typePos := -1
	for _, line := range lines {
		idx := -1
		for _, marker := range []string{"D ", "F ", "L "} {
			if i := strings.Index(line, marker); i >= 0 {
				idx = i
				break
			}
		}
		if idx < 0 {
			continue
		}
		if typePos == -1 {
			typePos = idx
		} else if idx != typePos {
			t.Fatalf("type column misaligned: line %q at %d, expected %d", line, idx, typePos)
		}
	}
	if typePos < 0 {
		t.Fatal("popup must contain at least one D/F/L row")
	}
	if !strings.Contains(out, "D ") {
		t.Fatal("popup must mark directories with D")
	}
	if !strings.Contains(out, "F ") {
		t.Fatal("popup must mark files with F")
	}
}

// TestMultiSelectCursorRowKeepsCheckbox verifies that the cursor row still
// shows an [x] or [ ] marker (BR-041 §4.3).
func TestMultiSelectCursorRowKeepsCheckbox(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerUpdate,
		Fields: []state.FormField{
			{
				Key:      "caps",
				Kind:     state.FormMultiSelect,
				Options:  []string{"read", "write", "inspect"},
				Selected: map[string]bool{"write": true},
			},
		},
	})
	m.Form.FieldFocus = 0
	m.Form.OpenPopup()
	// Cursor = 1 (write). Working copy is empty so write shows [ ] until
	// toggled. Either way the row must still display a checkbox.
	m.Form.Popup.Cursor = 1
	out := renderFormPopup(m.Form, "box", 80, 20)
	clean := stripANSI(out)
	if !strings.Contains(clean, "[ ] write") && !strings.Contains(clean, "[x] write") {
		t.Fatalf("cursor row must still carry a checkbox: %q", clean)
	}
}

// TestLongPathCursorStaysVisible ensures a focused path field with a long
// input keeps the cursor in view by horizontal scrolling (BR-041 §9.6 / §3.2).
func TestLongPathCursorStaysVisible(t *testing.T) {
	long := strings.Repeat("a", 80) + ".tar"
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerExport,
		Fields: []state.FormField{
			{
				Key:   "destination",
				Kind:  state.FormPath,
				Input: state.QueryInputState{Text: long, Cursor: len(long)},
			},
		},
	})
	m.Form.FieldFocus = 0
	out := FormDialog(m, "", LoadDialogConfig(), 0, 0)
	clean := stripANSI(out)
	if !strings.Contains(clean, component.NarrowCursor) {
		t.Fatal("focused long path must still show an end caret")
	}
	// The rendered line must fit within the dialog's outer width.
	// The dialog default MinWidth is 60, so a 60-cell viewport with a
	// long path must still fit within the box, allowing one extra cell
	// of slack for lipgloss padding rounding.
	m.Viewport.Width = 60
	m.Viewport.Height = 18
	out = FormDialog(m, "", LoadDialogConfig(), 0, 0)
	clean = stripANSI(out)
	const maxBoxCells = 62
	for _, line := range strings.Split(clean, "\n") {
		if utils.DisplayWidth(line) > maxBoxCells {
			t.Fatalf("dialog line %d cells, exceeds %d: %q", utils.DisplayWidth(line), maxBoxCells, line)
		}
	}
}

func TestEditablePathCursorDoesNotShiftCharacters(t *testing.T) {
	const path = "/home/debi/IdeaProjects/docker-tui/archive.tar"
	for _, cursor := range []int{0, 1, 6, 10, len([]rune(path)) - 1} {
		field := state.FormField{Kind: state.FormPath, Input: state.QueryInputState{Text: path, Cursor: cursor}}
		got := stripANSI(renderEditableValue(&field, true, 80))
		if got != path {
			t.Fatalf("cursor %d changed rendered path: got %q, want %q", cursor, got, path)
		}
	}

	field := state.FormField{Kind: state.FormPath, Input: state.NewQueryInput(path)}
	if got := stripANSI(renderEditableValue(&field, true, 80)); got != path+component.NarrowCursor {
		t.Fatalf("end cursor render = %q, want %q", got, path+component.NarrowCursor)
	}
}

func TestFocusedFormFieldHasBackground(t *testing.T) {
	field := state.FormField{Label: "Local destination (tar)", Kind: state.FormPath, Input: state.NewQueryInput("/home/debi/archive.tar")}
	form := state.FormState{Fields: []state.FormField{field}, FieldFocus: 0}
	row := renderFormField(form, &form.Fields[0], 0, 24, 50)
	if !strings.Contains(row, "\x1b[48") {
		t.Fatalf("focused form row must paint the input cell background: %q", row)
	}
}

// TestFormRendersInSmallViewports verifies that no overlap occurs at
// 60x16, 80x24, and 160x40 viewport sizes (BR-041 §6.14). The smallest
// dimension matches the dialog default MinWidth so the form layout has
// space to render both fields and the buttons.
func TestFormRendersInSmallViewports(t *testing.T) {
	sizes := []struct{ w, h int }{
		{60, 16},
		{80, 24},
		{160, 40},
	}
	for _, s := range sizes {
		m := formTestModel(state.FormSpec{
			Kind:  state.FormContainerCopy,
			Title: "Copy file from container",
			Fields: []state.FormField{
				{Key: "source", Label: "Container path", Kind: state.FormText, Input: state.QueryInputState{Text: "/etc/app.conf"}},
				{Key: "destination", Label: "Local destination (tar)", Kind: state.FormPath, Input: state.QueryInputState{Text: "/tmp/backup.tar"}},
			},
		})
		m.Viewport.Width = s.w
		m.Viewport.Height = s.h
		out := FormDialog(m, "", LoadDialogConfig(), 0, 0)
		lines := strings.Split(out, "\n")
		if len(lines) > s.h+2 {
			t.Fatalf("viewport %dx%d produced %d lines, expected at most %d", s.w, s.h, len(lines), s.h+2)
		}
		// Title row and button row must not exceed the configured width.
		for _, line := range lines {
			if w := utils.DisplayWidth(stripANSI(line)); w > s.w {
				t.Fatalf("viewport %dx%d line width %d > %d: %q", s.w, s.h, w, s.w, line)
			}
		}
	}
}

// TestFormDialogHeightEqualsBodyH verifies that with the default
// PanelSize (HeightPercent=100, MaxHeight=40) the form dialog height
// returned by formDialogSize equals bodyH when bodyH is between the
// requiredH floor and MaxHeight (BR-043 §3.2: 1:1 height with the
// panel body).
func TestFormDialogHeightEqualsBodyH(t *testing.T) {
	cfg := LoadDialogConfig()
	if cfg.PanelSize.HeightPercent != 100 {
		t.Fatalf("default PanelSize.HeightPercent = %d, want 100 (BR-043 §3.2)",
			cfg.PanelSize.HeightPercent)
	}
	if cfg.PanelSize.MaxHeight != 40 {
		t.Fatalf("default PanelSize.MaxHeight = %d, want 40", cfg.PanelSize.MaxHeight)
	}

	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerCopy,
		Fields: []state.FormField{
			{Key: "source", Label: "Container path", Kind: state.FormText, Input: state.QueryInputState{Text: "/etc/app.conf"}},
		},
	})
	// Viewport must be ≥ bodyH+2 so formDialogSize's maxH clamp
	// (Viewport.Height-2) does not lower the result.
	for _, bodyH := range []int{16, 20, 30, 40} {
		m.Viewport.Width = 100
		m.Viewport.Height = bodyH + 2
		dialogW, dialogH := formDialogSize(m, cfg, 80, bodyH)
		if dialogW <= 0 {
			t.Errorf("bodyH=%d: dialogW = %d, want > 0", bodyH, dialogW)
		}
		if dialogH != bodyH {
			t.Errorf("bodyH=%d: dialogH = %d, want %d (1:1 with bodyH)", bodyH, dialogH, bodyH)
		}
	}
}

// TestFormDialogHeightClampsToMaxHeight verifies that a bodyH larger
// than PanelSize.MaxHeight is capped at MaxHeight, so the dialog never
// exceeds the configured cap regardless of the active panel.
func TestFormDialogHeightClampsToMaxHeight(t *testing.T) {
	cfg := LoadDialogConfig()
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerCopy,
		Fields: []state.FormField{
			{Key: "source", Label: "Container path", Kind: state.FormText, Input: state.QueryInputState{Text: "/etc/app.conf"}},
		},
	})
	// bodyH=100 with default MaxHeight=40 → dialogH must be 40.
	m.Viewport.Width = 200
	m.Viewport.Height = 102
	_, dialogH := formDialogSize(m, cfg, 200, 100)
	if dialogH != 40 {
		t.Fatalf("bodyH=100: dialogH = %d, want 40 (capped to PanelSize.MaxHeight)", dialogH)
	}
}

// TestFormDialogWidthClampsToMaxWidth verifies that a wide panel body
// produces a dialog capped to PanelSize.MaxWidth (default 120 cells).
func TestFormDialogWidthClampsToMaxWidth(t *testing.T) {
	cfg := LoadDialogConfig()
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerCopy,
		Fields: []state.FormField{
			{Key: "source", Label: "Container path", Kind: state.FormText, Input: state.QueryInputState{Text: "/etc/app.conf"}},
		},
	})
	m.Viewport.Width = 400
	m.Viewport.Height = 60
	dialogW, _ := formDialogSize(m, cfg, 400, 60)
	// 400*75/100 = 300, capped to MaxWidth=120.
	if dialogW != 120 {
		t.Fatalf("bodyW=400: dialogW = %d, want 120 (capped to PanelSize.MaxWidth)", dialogW)
	}
}

// TestRenderSelectCellUsesDisplayOptions verifies the Phase 3 wire/display
// split: when a FormField carries both Options (Wire keys) and DisplayOptions
// (localized labels), the user sees the label, not the raw key.
func TestRenderSelectCellUsesDisplayOptions(t *testing.T) {
	field := state.FormField{
		Kind:           state.FormSelect,
		Options:        []string{"a", "b"},
		DisplayOptions: []string{"Alpha", "Beta"},
		Index:          1,
	}
	value, _ := renderSelectCell(&field, false, 80)
	clean := stripANSI(value)
	if !strings.Contains(clean, "Beta") {
		t.Fatalf("renderSelectCell must show DisplayOptions[1] = \"Beta\", got %q", clean)
	}
	if strings.Contains(clean, " b") || strings.HasSuffix(strings.TrimSpace(clean), "b") {
		// The raw key "b" must not leak through.
		trimmed := strings.TrimSpace(clean)
		if trimmed == "b" || strings.HasSuffix(trimmed, " b") {
			t.Fatalf("renderSelectCell leaked raw key \"b\" into output: %q", clean)
		}
	}
}

// TestRenderSelectCellFallbackToOptions ensures fields without DisplayOptions
// (legacy forms, dynamically-built selects) still render Options directly.
func TestRenderSelectCellFallbackToOptions(t *testing.T) {
	field := state.FormField{
		Kind:    state.FormSelect,
		Options: []string{"a", "b"},
		Index:   1,
	}
	value, _ := renderSelectCell(&field, false, 80)
	clean := stripANSI(value)
	if !strings.Contains(clean, "b") {
		t.Fatalf("renderSelectCell must fall back to Options[1] = \"b\", got %q", clean)
	}
}

// TestRenderSelectCellKeepsChevron locks the dropdown marker (component
// glyph: TriangleDownSmall) on every render path so visibility regressions
// (Phase 1 max-width handling, Phase 6 PTY capture) are immediately caught.
func TestRenderSelectCellKeepsChevron(t *testing.T) {
	cases := []state.FormField{
		{Kind: state.FormSelect, Options: []string{"a"}, DisplayOptions: []string{"Alpha"}},
		{Kind: state.FormSelect, Options: []string{"a"}},
		{Kind: state.FormMultiSelect, Options: []string{"a", "b"}, Selected: map[string]bool{"a": true}},
	}
	for i, f := range cases {
		_, marker := renderSelectCell(&f, false, 40)
		if marker != component.TriangleDownSmall {
			t.Errorf("case %d: marker = %q, want %q", i, marker, component.TriangleDownSmall)
		}
	}
	// End-to-end: the chevron must also appear in the rendered form output.
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerUpdate,
		Fields: []state.FormField{
			{
				Key:            "restart",
				Label:          "Restart policy",
				Kind:           state.FormSelect,
				Options:        []string{"unchanged", "no", "always"},
				DisplayOptions: []string{"Leave unchanged", "No restart", "Always"},
				Index:          2,
			},
		},
	})
	out := FormDialog(m, "", LoadDialogConfig(), 0, 0)
	if !strings.Contains(out, component.TriangleDownSmall) {
		t.Fatalf("update form must show the dropdown chevron in output: %q", out)
	}
	if !strings.Contains(out, "Always") {
		t.Fatalf("update form must show the DisplayOptions label \"Always\": %q", out)
	}
}

func TestRenderFormFieldShowsHelperText(t *testing.T) {
	field := state.FormField{Label: "Memory", Kind: state.FormInt, HelperText: "MB"}
	form := state.FormState{Fields: []state.FormField{field}}
	row := stripANSI(renderFormField(form, &form.Fields[0], 0, 24, 40))
	if !strings.Contains(row, " (MB)") {
		t.Fatalf("renderFormField must append HelperText in parentheses: %q", row)
	}
}

func TestRenderFormFieldOmitsHelperTextWhenEmpty(t *testing.T) {
	field := state.FormField{Label: "Memory", Kind: state.FormInt}
	form := state.FormState{Fields: []state.FormField{field}}
	row := stripANSI(renderFormField(form, &form.Fields[0], 0, 24, 40))
	if strings.Contains(row, "()") || strings.Contains(row, "(nil)") {
		t.Fatalf("empty HelperText must not render stray parens: %q", row)
	}
	if !strings.Contains(row, "Memory") {
		t.Fatalf("raw label must remain visible: %q", row)
	}
}

func TestRenderFormFieldShowsUnit(t *testing.T) {
	field := state.FormField{Label: "CPUs", Kind: state.FormInt, Unit: "cores", Input: state.NewQueryInput("1.5")}
	form := state.FormState{Fields: []state.FormField{field}}
	row := stripANSI(renderFormField(form, &form.Fields[0], 0, 24, 40))
	if !strings.Contains(row, "cores") {
		t.Fatalf("renderFormField must surface Unit suffix in output: %q", row)
	}
}
