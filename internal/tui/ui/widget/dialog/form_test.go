package dialog

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"

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
			NewTextField(TextFieldConfig{Key: "source", Label: "Container path", Text: "/etc/app.conf"}),
			NewPathField(PathFieldConfig{Key: "destination", Label: "Local destination (tar)", Text: "/tmp/backup.tar"}),
		},
	})
	out := FormDialog(m, "", LoadDialogConfig(), 0, 0)
	if !strings.Contains(out, "Copy file from container") {
		t.Fatal("dialog must render the form title")
	}
	if !strings.Contains(out, "Local destination (tar)") {
		t.Fatal("dialog must render the destination label")
	}
	if !strings.Contains(stripANSI(out), "/etc/app.conf") {
		t.Fatal("dialog must render the source value")
	}
}

func TestFormDialogSelectCollapsedWithMarker(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerUpdate,
		Fields: []state.FormField{
			NewSelectField(SelectFieldConfig{
				Key: "restart", Label: "Restart policy",
				Options: []string{"unchanged", "no", "always"}, Index: 2,
			}),
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
	source := NewTextField(TextFieldConfig{Key: "source", Label: "Container path"})
	source.SetError("required")
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerCopy,
		Fields: []state.FormField{
			source,
			NewPathField(PathFieldConfig{Key: "destination", Label: "Local destination"}),
		},
	})
	out := FormDialog(m, "", LoadDialogConfig(), 0, 0)
	if !strings.Contains(out, "required") {
		t.Fatal("field error must be rendered below the field row")
	}
}

func TestFormDialogNoPopupWhenClosed(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind:   state.FormContainerExport,
		Fields: []state.FormField{NewPathField(PathFieldConfig{Key: "destination", Label: "Local destination"})},
	})
	out := FormDialog(m, "", LoadDialogConfig(), 0, 0)
	if strings.Contains(out, ">") && !strings.Contains(out, component.ButtonIndicator) {
		t.Fatal("closed popup must not render cursor rows")
	}
}

func TestFormLayoutWidthsOutsideIn(t *testing.T) {
	fields := []state.FormField{
		NewTextField(TextFieldConfig{Key: "a", Label: "Container path"}),
		NewPathField(PathFieldConfig{Key: "b", Label: "Local destination (tar)"}),
		NewIntField(IntFieldConfig{Key: "c", Label: "Max retries"}),
	}
	labelW, valueW := formLayout(fields, 40)
	if labelW <= 0 || valueW <= 0 || labelW+valueW+1 != 40 {
		t.Fatalf("layout must fill inner width exactly: label=%d value=%d sum=%d", labelW, valueW, labelW+valueW+1)
	}
	if labelW > 20 {
		t.Fatalf("label column too wide: %d", labelW)
	}
}

func TestRenderFormFieldShowsHelperText(t *testing.T) {
	field := NewIntField(IntFieldConfig{Key: "memory", Label: "Memory", HelperText: "MB"})
	form := state.FormState{Fields: []state.FormField{field}}
	row := stripANSI(renderFormField(form, field, 0, 24, 40, true, formOperationStyles{}))
	if !strings.Contains(row, " (MB)") {
		t.Fatalf("renderFormField must append HelperText in parentheses: %q", row)
	}
}

func TestRenderFormFieldOmitsHelperTextWhenEmpty(t *testing.T) {
	field := NewIntField(IntFieldConfig{Key: "memory", Label: "Memory"})
	form := state.FormState{Fields: []state.FormField{field}}
	row := stripANSI(renderFormField(form, field, 0, 24, 40, true, formOperationStyles{}))
	if strings.Contains(row, "()") {
		t.Fatalf("empty HelperText must not render stray parens: %q", row)
	}
	if !strings.Contains(row, "Memory") {
		t.Fatalf("raw label must remain visible: %q", row)
	}
}

func TestRenderSelectCellUsesDisplayOptions(t *testing.T) {
	value, _ := renderSelectCellValue(
		[]string{"a", "b"}, []string{"Alpha", "Beta"}, 1, nil, state.FormSelect, false, 80,
	)
	clean := stripANSI(value)
	if !strings.Contains(clean, "Beta") {
		t.Fatalf("renderSelectCellValue must show DisplayOptions[1] = \"Beta\", got %q", clean)
	}
}

func TestRenderSelectCellKeepsChevron(t *testing.T) {
	_, marker := renderSelectCellValue(
		[]string{"a", "b"}, []string{"Alpha", "Beta"}, 1, nil, state.FormSelect, false, 40,
	)
	if marker != component.TriangleDownSmall {
		t.Fatalf("marker = %q, want %q", marker, component.TriangleDownSmall)
	}
}

func TestRenderFormPopupSelectRows(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerUpdate,
		Fields: []state.FormField{
			NewSelectField(SelectFieldConfig{
				Key: "restart", Label: "Restart policy",
				Options: []string{"no", "always", "on-failure"},
			}),
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
			NewMultiSelectField(MultiSelectFieldConfig{
				Key: "multi", Label: "Multi",
				Options:  []string{"a", "b", "c"},
				Selected: map[string]bool{"b": true},
			}),
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
	dest := NewPathField(PathFieldConfig{Key: "destination", Label: "Local destination"})
	pf := dest.(*PathField)
	pf.SetSuggestions([]state.PathEntry{
		{Name: "backup.tar", Path: "/tmp/backup.tar", IsDir: false, Type: state.PathEntryFile},
		{Name: "sub", Path: "/tmp/sub", IsDir: true, Type: state.PathEntryDir},
	})
	m := formTestModel(state.FormSpec{
		Kind:   state.FormContainerExport,
		Fields: []state.FormField{dest},
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

func TestRenderPathValueLongWraps(t *testing.T) {
	const width = 12
	path := "/tmp/very-long-backup-archive-2026.tar.gz"
	pf := NewPathField(PathFieldConfig{Key: "dest", Text: path}).(*PathField)
	got := stripANSI(pf.Render(false, width))
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

func TestRenderPathValueFocusedKeepsCursorWindow(t *testing.T) {
	path := "/very/long/path/to/some/deep/backup-archive-2026.tar.gz"
	pf := NewPathField(PathFieldConfig{Key: "dest", Text: path, Cursor: len(path)}).(*PathField)
	got := stripANSI(pf.Render(true, 16))
	if strings.Contains(got, "\n") {
		t.Fatalf("focused path must remain a single-line cursor window: %q", got)
	}
	if utils.DisplayWidth(got) > 16 {
		t.Fatalf("focused path width = %d, want <= 16: %q", utils.DisplayWidth(got), got)
	}
}

func TestFocusedFormFieldHasBackground(t *testing.T) {
	field := NewPathField(PathFieldConfig{Key: "destination", Label: "Local destination (tar)", Text: "/home/debi/archive.tar"})
	form := state.FormState{Fields: []state.FormField{field}, FieldFocus: 0}
	row := renderFormField(form, field, 0, 24, 50, true, formOperationStyles{})
	if !strings.Contains(row, "\x1b[48") {
		t.Fatalf("focused form row must paint the input cell background: %q", row)
	}
}

func TestFormDialogHeightEqualsBodyH(t *testing.T) {
	cfg := LoadDialogConfig()
	if cfg.PanelSize.HeightPercent != 100 {
		t.Fatalf("default PanelSize.HeightPercent = %d, want 100", cfg.PanelSize.HeightPercent)
	}
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerCopy,
		Fields: []state.FormField{
			NewTextField(TextFieldConfig{Key: "source", Label: "Container path", Text: "/etc/app.conf"}),
		},
	})
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

func TestFormDialogHeightClampsToMaxHeight(t *testing.T) {
	cfg := LoadDialogConfig()
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerCopy,
		Fields: []state.FormField{
			NewTextField(TextFieldConfig{Key: "source", Label: "Container path", Text: "/etc/app.conf"}),
		},
	})
	m.Viewport.Width = 200
	m.Viewport.Height = 102
	_, dialogH := formDialogSize(m, cfg, 200, 100)
	if dialogH != 40 {
		t.Fatalf("bodyH=100: dialogH = %d, want 40 (capped to PanelSize.MaxHeight)", dialogH)
	}
}

func TestFormDialogWidthClampsToMaxWidth(t *testing.T) {
	cfg := LoadDialogConfig()
	m := formTestModel(state.FormSpec{
		Kind: state.FormContainerCopy,
		Fields: []state.FormField{
			NewTextField(TextFieldConfig{Key: "source", Label: "Container path", Text: "/etc/app.conf"}),
		},
	})
	m.Viewport.Width = 400
	m.Viewport.Height = 60
	dialogW, _ := formDialogSize(m, cfg, 400, 60)
	if dialogW != 120 {
		t.Fatalf("bodyW=400: dialogW = %d, want 120 (capped to PanelSize.MaxWidth)", dialogW)
	}
}
