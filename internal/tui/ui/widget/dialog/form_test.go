package dialog

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/utils"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// stripANSI removes ANSI escape sequences for column-position assertions.
func stripANSI(s string) string { return ansi.Strip(s) }

// formTestModel returns a model with a Form open and a sized viewport.
func formTestModel(spec state.FormSpec) *state.AppModel {
	m := state.NewAppModel(config.DefaultConfig(), nil, "test")
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
	out := FormDialog(m, "", defaultDialogConfig())
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
	out := FormDialog(m, "", defaultDialogConfig())
	if !strings.Contains(out, "always") {
		t.Fatal("collapsed select must show the selected option")
	}
	if !strings.Contains(out, "\u25be") {
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
	out := FormDialog(m, "", defaultDialogConfig())
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
	out := stripANSI(FormDialog(m, "", defaultDialogConfig()))
	if strings.Contains(out, "\u25b6") {
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
					{Name: "backup.tar", Path: "/tmp/backup.tar", IsDir: false},
					{Name: "sub", Path: "/tmp/sub", IsDir: true},
				},
			},
		},
	})
	m.Form.FieldFocus = 0
	m.Form.OpenPopup()
	out := renderFormPopup(m.Form, "box", 60, 20)
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
		return renderFormPopup(m.Form, "box", 60, 20)
	}
	one := render([]state.PathEntry{{Name: "apk", Path: "/etc/apk", IsDir: true}}, false)
	many := render([]state.PathEntry{
		{Name: "a", Path: "/etc/a", IsDir: true}, {Name: "b", Path: "/etc/b", IsDir: true},
		{Name: "c", Path: "/etc/c"}, {Name: "d", Path: "/etc/d"},
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
	out := FormDialog(m, "", defaultDialogConfig())
	if strings.Contains(out, ">") && !strings.Contains(out, "\u25b6") {
		t.Fatal("closed popup must not render cursor rows")
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
					{Name: "alpha.tar", Path: "/tmp/alpha.tar", IsDir: false},
					{Name: "beta", Path: "/tmp/beta", IsDir: true},
					{Name: "gamma", Path: "/tmp/gamma", IsDir: true},
					{Name: "delta.tar", Path: "/tmp/delta.tar", IsDir: false},
				},
			},
		},
	})
	m.Form.FieldFocus = 0
	m.Form.OpenPopup()
	out := renderFormPopup(m.Form, "box", 80, 20)
	clean := stripANSI(out)
	lines := strings.Split(clean, "\n")
	typePos := -1
	for _, line := range lines {
		if strings.Contains(line, "[DIR]") || strings.Contains(line, "[FILE]") {
			idx := strings.Index(line, "[")
			if idx < 0 {
				continue
			}
			if typePos == -1 {
				typePos = idx
			} else if idx != typePos {
				t.Fatalf("type column misaligned: line %q at %d, expected %d", line, idx, typePos)
			}
		}
	}
	if typePos < 0 {
		t.Fatal("popup must contain at least one [DIR]/[FILE] row")
	}
	if !strings.Contains(out, "[DIR]") {
		t.Fatal("popup must mark directories with [DIR]")
	}
	if !strings.Contains(out, "[FILE]") {
		t.Fatal("popup must mark files with [FILE]")
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
	out := FormDialog(m, "", defaultDialogConfig())
	clean := stripANSI(out)
	if !strings.Contains(clean, "\u258f") {
		t.Fatal("focused long path must still show an end caret")
	}
	// The rendered line must fit within the dialog's outer width.
	for _, line := range strings.Split(clean, "\n") {
		if utils.DisplayWidth(line) > 60 {
			t.Fatalf("dialog line %d cells, exceeds inner width: %q", utils.DisplayWidth(line), line)
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
	if got := stripANSI(renderEditableValue(&field, true, 80)); got != path+"\u258f" {
		t.Fatalf("end cursor render = %q, want %q", got, path+"\u258f")
	}
}

func TestFocusedFormFieldDoesNotAddBackground(t *testing.T) {
	field := state.FormField{Label: "Local destination (tar)", Kind: state.FormPath, Input: state.NewQueryInput("/home/debi/archive.tar")}
	form := state.FormState{Fields: []state.FormField{field}, FieldFocus: 0}
	row := renderFormField(form, &form.Fields[0], 0, 24, 50)
	if strings.Contains(row, "\x1b[48") {
		t.Fatalf("focused form row must not set a background colour: %q", row)
	}
}

// TestFormRendersInSmallViewports verifies that no overlap occurs at
// 40x16, 80x24, and 160x40 viewport sizes (BR-041 §6.14).
func TestFormRendersInSmallViewports(t *testing.T) {
	sizes := []struct{ w, h int }{
		{40, 16},
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
		out := FormDialog(m, "", defaultDialogConfig())
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
