package dialog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"

	"charm.land/lipgloss/v2"
)

func TestRegistryCoversAllKinds(t *testing.T) {
	want := []state.FormFieldKind{
		state.FormText, state.FormInt, state.FormSelect, state.FormBool,
		state.FormPath, state.FormMultiSelect, state.FormRadioGroup,
		state.FormTextMultiLine, state.FormTextPassword,
	}
	for _, k := range want {
		if _, ok := Registry[k]; !ok {
			t.Fatalf("Registry missing constructor for %v", k)
		}
	}
}

func TestFormFieldForUnknownKindFallsBackToText(t *testing.T) {
	f := &state.FormField{Kind: state.FormFieldKind(9999), Label: "x"}
	impl := FormFieldFor(f)
	if impl == nil {
		t.Fatal("FormFieldFor must not return nil for unknown kind")
	}
	if _, ok := impl.(*TextField); !ok {
		t.Fatalf("unknown kind fallback = %T, want *TextField", impl)
	}
}

func TestFormFieldForNilFieldReturnsZeroTextField(t *testing.T) {
	impl := FormFieldFor(nil)
	if _, ok := impl.(*TextField); !ok {
		t.Fatalf("FormFieldFor(nil) = %T, want *TextField", impl)
	}
}

func TestTextFieldRoundtripValueAndValidate(t *testing.T) {
	src := &state.FormField{Kind: state.FormText, Required: true}
	impl := newTextField(src)
	if impl.Kind() != state.FormText {
		t.Fatalf("Kind = %v, want FormText", impl.Kind())
	}
	if d := impl.Default(); d != "" {
		t.Fatalf("Default = %q, want empty", d)
	}
	if err := impl.Validate(); err == nil {
		t.Fatal("Validate on empty Required field must return error")
	}
	impl.SetValue("hello")
	if v := impl.Value().(string); v != "hello" {
		t.Fatalf("Value = %q, want %q", v, "hello")
	}
	if err := impl.Validate(); err != nil {
		t.Fatalf("Validate after SetValue: %v", err)
	}
}

func TestIntFieldValidatesRange(t *testing.T) {
	min, max := 0.0, 100.0
	src := &state.FormField{Kind: state.FormInt, Min: &min, Max: &max}
	impl := newIntField(src)
	impl.SetValue("5")
	if err := impl.Validate(); err != nil {
		t.Fatalf("valid value error: %v", err)
	}
	impl.SetValue("200")
	if err := impl.Validate(); err == nil {
		t.Fatal("value above max must fail Validate")
	}
	impl.SetValue("-1")
	if err := impl.Validate(); err == nil {
		t.Fatal("value below min must fail Validate")
	}
	impl.SetValue("not-a-number")
	if err := impl.Validate(); err == nil {
		t.Fatal("non-numeric must fail Validate")
	}
}

func TestBoolFieldToggleKey(t *testing.T) {
	m := formTestModel(state.FormSpec{Kind: state.FormContainerRemove, Fields: []state.FormField{
		{Key: "force", Kind: state.FormBool, Toggle: false},
		{Key: "child", Kind: state.FormBool, DependsOn: "force", DependsEq: true},
	}})
	impl := newBoolField(&m.Form.Fields[0])
	if v := impl.Value().(bool); v {
		t.Fatal("initial Value = true, want false")
	}
	handled, updated := impl.HandleKey(" ", m)
	if !handled || !updated {
		t.Fatal("Space on Bool must return (true, true)")
	}
	impl.SyncTo(&m.Form.Fields[0])
	if !m.Form.Fields[0].Toggle {
		t.Fatal("Toggle must flip after SyncTo")
	}
	m.Form.RecomputeVisibility()
	if m.Form.Fields[1].Hidden {
		t.Fatal("dependent field must become visible after toggle")
	}
	impl.HandleKey("a", m)    // printable → consumed, no update
	impl.HandleKey("left", m) // Left → consumed, no update
}

func TestTextFieldBackspaceDeleteKey(t *testing.T) {
	src := &state.FormField{Kind: state.FormText, Input: state.NewQueryInput("hello")}
	impl := newTextField(src)
	if handled, updated := impl.HandleKey("backspace", nil); !handled || !updated {
		t.Fatalf("backspace = (%v, %v), want (true, true)", handled, updated)
	}
	if v := impl.Value().(string); v != "hell" {
		t.Fatalf("after backspace Value = %q, want %q", v, "hell")
	}
}

func TestPathFieldCtrlHTogglesHidden(t *testing.T) {
	src := &state.FormField{Kind: state.FormPath, ShowHidden: false, Suggestions: []state.PathEntry{{Name: "a"}}}
	impl := newPathField(src)
	handled, updated := impl.HandleKey("ctrl+h", nil)
	if !handled || updated {
		t.Fatalf("Ctrl+H = (%v, %v), want (true, false)", handled, updated)
	}
	impl.SyncTo(src)
	if !src.ShowHidden {
		t.Fatal("ShowHidden must flip after SyncTo")
	}
	if src.Suggestions != nil {
		t.Fatal("Suggestions must clear")
	}
}

func TestSelectFieldRoundtrip(t *testing.T) {
	src := &state.FormField{Kind: state.FormSelect, Options: []string{"a", "b", "c"}, Index: 1}
	impl := newSelectField(src)
	if v := impl.Value().(string); v != "b" {
		t.Fatalf("Value = %q, want %q", v, "b")
	}
	impl.SetValue("c")
	impl.SyncTo(src)
	if src.Index != 2 {
		t.Fatalf("Index = %d, want 2 after SyncTo", src.Index)
	}
	impl.SetValue("missing") // unknown key → no change
	impl.SyncTo(src)
	if src.Index != 2 {
		t.Fatalf("SetValue with unknown key must not change Index, got %d", src.Index)
	}
}

func TestMultiSelectFieldRoundtrip(t *testing.T) {
	src := &state.FormField{Kind: state.FormMultiSelect, Options: []string{"a", "b"}}
	impl := newMultiSelectField(src)
	m := map[string]bool{"a": true}
	impl.SetValue(m)
	if !impl.selected["a"] {
		t.Fatal("SetValue must populate impl.selected")
	}
	m["b"] = true // mutate original — impl's copy must be isolated
	if impl.selected["b"] {
		t.Fatal("SetValue must deep-copy the map")
	}
}

func TestRadioGroupFieldRenderIncludesCheckGlyph(t *testing.T) {
	src := &state.FormField{
		Kind:    state.FormRadioGroup,
		Label:   "Group",
		Options: []string{"a", "b"},
		Index:   1,
	}
	impl := newRadioGroupField(src)
	out := stripANSI(impl.Render(true, 40))
	if !strings.Contains(out, "b") {
		t.Fatalf("render must contain the selected option %q: %q", "b", out)
	}
	if !strings.Contains(out, component.MarkCheck) {
		t.Fatalf("render must carry radio glyph %q: %q", component.MarkCheck, out)
	}
}

func TestTextPasswordFieldRendersMasked(t *testing.T) {
	src := &state.FormField{Kind: state.FormTextPassword, Input: state.NewQueryInput("hunter2")}
	impl := newTextPasswordField(src)
	out := stripANSI(impl.Render(false, 40))
	if strings.ContainsRune(out, 'h') || strings.Contains(out, "hunter") {
		t.Fatalf("render must mask the raw value: %q", out)
	}
	if !strings.Contains(out, passwordBullet) {
		t.Fatalf("render must contain mask glyph: %q", out)
	}
	raw := stripANSI(impl.Render(true, 40))
	if strings.ContainsRune(raw, 'h') {
		t.Fatalf("focused render must also mask the raw value: %q", raw)
	}
	if v := impl.Value().(string); v != "hunter2" {
		t.Fatalf("Value must preserve raw text, got %q", v)
	}
}

func TestTextMultiLineFieldRendersFirstLineOnly(t *testing.T) {
	src := &state.FormField{Kind: state.FormTextMultiLine, Input: state.NewQueryInput("first\nsecond")}
	impl := newTextMultiLineField(src)
	out := stripANSI(impl.Render(false, 80))
	if strings.Contains(out, "second") {
		t.Fatalf("multi-line render must hide lines after the first: %q", out)
	}
	if !strings.Contains(out, "first") {
		t.Fatalf("multi-line render must show the first line: %q", out)
	}
}

func TestSelectLabelRespectsDisplayOptions(t *testing.T) {
	f := &state.FormField{Options: []string{"a", "b"}, DisplayOptions: []string{"Alpha", ""}}
	if got := selectLabel(f.Options, f.DisplayOptions, 0); got != "Alpha" {
		t.Fatalf("DisplayOptions[0] = %q, want Alpha", got)
	}
	if got := selectLabel(f.Options, f.DisplayOptions, 1); got != "b" {
		t.Fatalf("empty DisplayOptions falls back to Options[1] = %q, want b", got)
	}
	if got := selectLabel(f.Options, f.DisplayOptions, 5); got != "" {
		t.Fatalf("out-of-range index = %q, want empty", got)
	}
}

func TestAppendUnitSuffixEmptyUnitFitsToWidth(t *testing.T) {
	out := appendUnitSuffix("hello", "", 5)
	if w := lipgloss.Width(out); w > 5 {
		t.Fatalf("width = %d, want <= 5: %q", w, out)
	}
	if !strings.Contains(out, "h") {
		t.Fatalf("must preserve content: %q", out)
	}
}

func TestAppendDropdownMarkerAppendsGlyph(t *testing.T) {
	out := appendDropdownMarker("text", component.TriangleDownSmall, 10)
	if !strings.HasSuffix(out, " "+component.TriangleDownSmall) {
		t.Fatalf("marker must be appended with separator: %q", out)
	}
	if appendDropdownMarker("text", "", 10) != "text" {
		t.Fatal("empty marker must leave value unchanged")
	}
}

func TestRenderFormFieldUsesInterfaceImpl(t *testing.T) {
	m := formTestModel(state.FormSpec{
		Kind:   state.FormContainerUpdate,
		Fields: []state.FormField{{Key: "r", Label: "R", Kind: state.FormSelect, Options: []string{"unchanged", "always"}, Index: 1}},
	})
	m.Form.FieldFocus = 0
	out := FormDialog(m, "", LoadDialogConfig(), 0, 0)
	if !strings.Contains(out, component.TriangleDownSmall) {
		t.Fatal("FormDialog must carry the dropdown marker via SelectField.Render")
	}
	if !strings.Contains(out, "always") {
		t.Fatal("FormDialog must show the selected option")
	}
}

func TestSyncToRoundTripsAllImpls(t *testing.T) {
	for i, kind := range []state.FormFieldKind{
		state.FormText, state.FormInt, state.FormBool, state.FormPath,
		state.FormSelect, state.FormMultiSelect, state.FormRadioGroup,
		state.FormTextMultiLine, state.FormTextPassword,
	} {
		kind := kind
		t.Run(fmt.Sprintf("kind-%d", i), func(t *testing.T) {
			ctor, ok := Registry[kind]
			if !ok {
				t.Fatalf("Registry missing constructor for %v", kind)
			}
			// SyncTo must complete without panicking for every kind; the
			// SyncTo assertion only proves the seam exists.
			src := &state.FormField{Kind: kind}
			dst := &state.FormField{Kind: kind}
			ctor(src).SyncTo(dst)
		})
	}
}