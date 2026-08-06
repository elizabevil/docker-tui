package dialog

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// The constructors return state.FormField impls with correct identity and
// kind. Kind/Key are the two getters every consumer relies on.
func TestNewXxxFieldIdentityAndKind(t *testing.T) {
	cases := []struct {
		name string
		impl state.FormField
		want state.FormFieldKind
	}{
		{"text", NewTextField(TextFieldConfig{Key: "k1"}), state.FormText},
		{"int", NewIntField(IntFieldConfig{Key: "k2"}), state.FormInt},
		{"select", NewSelectField(SelectFieldConfig{Key: "k3"}), state.FormSelect},
		{"bool", NewBoolField(BoolFieldConfig{Key: "k4"}), state.FormBool},
		{"path", NewPathField(PathFieldConfig{Key: "k5"}), state.FormPath},
		{"multiselect", NewMultiSelectField(MultiSelectFieldConfig{Key: "k6"}), state.FormMultiSelect},
		{"radiogroup", NewRadioGroupField(RadioGroupFieldConfig{Key: "k7"}), state.FormRadioGroup},
		{"multiline", NewTextMultiLineField(TextMultiLineFieldConfig{Key: "k8"}), state.FormTextMultiLine},
		{"password", NewTextPasswordField(TextPasswordFieldConfig{Key: "k9"}), state.FormTextPassword},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.impl.Kind() != tc.want {
				t.Fatalf("Kind = %v, want %v", tc.impl.Kind(), tc.want)
			}
			if tc.impl.Key() != "k" && !strings.HasPrefix(tc.impl.Key(), "k") {
				t.Fatalf("Key = %q, want k-prefixed", tc.impl.Key())
			}
		})
	}
}

// Every impl must implement the dialog.FormField contract so the form row
// renderer can treat them uniformly.
func TestImplsSatisfyDialogFormField(t *testing.T) {
	impls := []state.FormField{
		NewTextField(TextFieldConfig{Key: "a", Label: "L1", HelperText: "h1"}),
		NewIntField(IntFieldConfig{Key: "b", Label: "L2", HelperText: "h2"}),
		NewSelectField(SelectFieldConfig{Key: "c", Options: []string{"x"}, Label: "L3"}),
		NewBoolField(BoolFieldConfig{Key: "d", Label: "L4"}),
		NewPathField(PathFieldConfig{Key: "e", Label: "L5", HelperText: "h5"}),
		NewMultiSelectField(MultiSelectFieldConfig{Key: "f", Options: []string{"x"}, Label: "L6"}),
		NewRadioGroupField(RadioGroupFieldConfig{Key: "g", Options: []string{"x"}, Label: "L7"}),
		NewTextMultiLineField(TextMultiLineFieldConfig{Key: "h", Label: "L8"}),
		NewTextPasswordField(TextPasswordFieldConfig{Key: "i", Label: "L9"}),
	}
	for _, impl := range impls {
		df, ok := impl.(FormField)
		if !ok {
			t.Fatalf("%T does not satisfy dialog.FormField", impl)
		}
		if df.Label() == "" {
			t.Fatalf("%T: configured label not returned", impl)
		}
		if df.HelpText() == "" {
			t.Fatalf("%T: HelpText must not be empty", impl)
		}
	}
}

// Config helper text round-trips through HelperText, which the form-row
// renderer reads to build the "Label (helper)" line.
func TestHelperTextRoundTrip(t *testing.T) {
	f, ok := NewTextField(TextFieldConfig{Key: "n", Label: "Name", HelperText: "required"}).(*TextField)
	if !ok {
		t.Fatal("NewTextField must return *TextField")
	}
	if f.HelperText() != "required" {
		t.Fatalf("HelperText = %q, want %q", f.HelperText(), "required")
	}
}

// TextField: HandleKey drives text/cursor, Value trims whitespace.
func TestTextFieldEditAndValue(t *testing.T) {
	f, ok := NewTextField(TextFieldConfig{Key: "name", Required: true}).(*TextField)
	if !ok {
		t.Fatal("NewTextField must return *TextField")
	}
	if _, c := f.HandleKey("h"); !c {
		t.Fatal("typing 'h' must mark the field changed")
	}
	if _, c := f.HandleKey("i"); !c {
		t.Fatal("typing 'i' must mark the field changed")
	}
	if f.Text() != "hi" {
		t.Fatalf("Text = %q, want %q", f.Text(), "hi")
	}
	if v := f.Value().(string); v != "hi" {
		t.Fatalf("Value = %q, want %q", v, "hi")
	}
	if !f.Touched() {
		t.Fatal("typing must set Touched")
	}
	f.SetText("  padded  ")
	if v := f.Value().(string); v != "padded" {
		t.Fatalf("Value must trim: %q", v)
	}
	f.Reset()
	if f.Text() != "" || f.Touched() || f.Error() != "" {
		t.Fatalf("Reset left state: text=%q touched=%v error=%q", f.Text(), f.Touched(), f.Error())
	}
}

// BoolField: Space toggles; arrow/char keys are consumed without toggling.
func TestBoolFieldSpaceToggles(t *testing.T) {
	f, ok := NewBoolField(BoolFieldConfig{Key: "force"}).(*BoolField)
	if !ok {
		t.Fatal("NewBoolField must return *BoolField")
	}
	if f.Toggle() {
		t.Fatal("new BoolField must start off")
	}
	if h, _ := f.HandleKey(keys.KeySpace); !h {
		t.Fatal("Space must be handled")
	}
	if !f.Toggle() || !f.Touched() {
		t.Fatalf("after Space: toggle=%v touched=%v", f.Toggle(), f.Touched())
	}
	if h, c := f.HandleKey(keys.KeyLeft); !h || c {
		t.Fatalf("arrow key: handled=%v changed=%v, want handled without change", h, c)
	}
	if !f.Toggle() {
		t.Fatal("arrow key must not toggle")
	}
}

// PathField is the special form: its path-specific methods are NOT on the
// FormField interface; callers type-assert to *PathField.
func TestPathFieldSpecialMethodsViaTypeAssert(t *testing.T) {
	impl := NewPathField(PathFieldConfig{
		Key:               "target",
		PathSource:        state.PathContainer,
		PathMode:          state.PathSaveFile,
		InitialShowHidden: true,
		InitialTabInput:   "/tmp/partial",
		DependsOn:         "tar",
		DependsEq:         true,
	})
	pf, ok := impl.(*PathField)
	if !ok {
		t.Fatalf("NewPathField must return *PathField, got %T", impl)
	}
	if pf.PathSource() != state.PathContainer {
		t.Fatalf("PathSource = %v, want PathContainer", pf.PathSource())
	}
	if pf.PathMode() != state.PathSaveFile {
		t.Fatalf("PathMode = %v, want PathSaveFile", pf.PathMode())
	}
	if !pf.ShowHidden() {
		t.Fatal("InitialShowHidden must be honored")
	}
	if pf.PathTabInput() != "/tmp/partial" {
		t.Fatalf("PathTabInput = %q", pf.PathTabInput())
	}
	if pf.DependsOn() != "tar" || !pf.DependsEq() {
		t.Fatalf("dependency config lost: on=%q eq=%v", pf.DependsOn(), pf.DependsEq())
	}
	// Common methods still work through the interface.
	if impl.Kind() != state.FormPath || impl.Key() != "target" {
		t.Fatalf("common identity wrong: kind=%v key=%q", impl.Kind(), impl.Key())
	}
	// Suggestions flow through the type-asserted API used by keyboard layer.
	entries := []state.PathEntry{{Name: "a.txt", Path: "/tmp/a.txt"}}
	pf.SetSuggestions(entries)
	if got := pf.Suggestions(); len(got) != 1 || got[0].Name != "a.txt" {
		t.Fatalf("Suggestions = %#v", got)
	}
	pf.SetPathLoading(true)
	if !pf.PathLoading() {
		t.Fatal("PathLoading must be true after SetPathLoading(true)")
	}
	pf.SetPathError("boom")
	if pf.PathError() != "boom" {
		t.Fatalf("PathError = %q", pf.PathError())
	}
}

// PathField: Ctrl+H toggles show-hidden and clears transient completion state.
func TestPathFieldCtrlHTogglesShowHidden(t *testing.T) {
	pf := NewPathField(PathFieldConfig{Key: "src"}).(*PathField)
	pf.SetSuggestions([]state.PathEntry{{Name: "x"}})
	pf.SetPathLoading(true)
	pf.SetPathError("e")
	if h, c := pf.HandleKey(keys.KeyCtrlH); !h || c {
		t.Fatalf("Ctrl+H: handled=%v changed=%v", h, c)
	}
	if !pf.ShowHidden() {
		t.Fatal("Ctrl+H must toggle ShowHidden on")
	}
	if len(pf.Suggestions()) != 0 || pf.PathLoading() || pf.PathError() != "" {
		t.Fatalf("Ctrl+H must clear suggestions/loading/error: sug=%d load=%v err=%q",
			len(pf.Suggestions()), pf.PathLoading(), pf.PathError())
	}
}

// SelectField: Value returns the option at the current index.
func TestSelectFieldValueFollowsIndex(t *testing.T) {
	f, ok := NewSelectField(SelectFieldConfig{
		Key: "level", Options: []string{"low", "mid", "high"}, Index: 1,
	}).(*SelectField)
	if !ok {
		t.Fatal("NewSelectField must return *SelectField")
	}
	if v := f.Value().(string); v != "mid" {
		t.Fatalf("Value = %q, want %q", v, "mid")
	}
	f.SetIndex(2)
	if v := f.Value().(string); v != "high" {
		t.Fatalf("Value = %q, want %q", v, "high")
	}
	if _, ok := f.OpenPopup(); !ok {
		t.Fatal("SelectField with options must open a popup")
	}
}

// IntField: Validate enforces parse + min/max range.
func TestIntFieldValidateRange(t *testing.T) {
	min, max := 1.0, 10.0
	f, ok := NewIntField(IntFieldConfig{Key: "cpu", Min: &min, Max: &max}).(*IntField)
	if !ok {
		t.Fatal("NewIntField must return *IntField")
	}
	f.SetText("5")
	if err := f.Validate(); err != nil {
		t.Fatalf("valid value rejected: %v", err)
	}
	f.SetText("abc")
	if err := f.Validate(); err == nil {
		t.Fatal("non-numeric must fail Validate")
	}
	f.SetText("0")
	if err := f.Validate(); err == nil {
		t.Fatal("below min must fail Validate")
	}
	f.SetText("11")
	if err := f.Validate(); err == nil {
		t.Fatal("above max must fail Validate")
	}
}

// TextPasswordField masks the rendered value; the stored text stays plain.
func TestTextPasswordFieldMasksRender(t *testing.T) {
	f, ok := NewTextPasswordField(TextPasswordFieldConfig{Key: "secret", Text: "abc"}).(*TextPasswordField)
	if !ok {
		t.Fatal("NewTextPasswordField must return *TextPasswordField")
	}
	if f.Text() != "abc" {
		t.Fatalf("stored text = %q, want %q", f.Text(), "abc")
	}
	if !strings.Contains(f.Render(false, 20), "•") {
		t.Fatal("Render must contain bullet mask")
	}
	if strings.Contains(f.Render(false, 20), "abc") {
		t.Fatal("Render must not leak plaintext")
	}
}
