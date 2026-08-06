package state

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/utils"
)

func TestToggleMulti(t *testing.T) {
	f := &FormField{Kind: FormMultiSelect, Options: []string{"a", "b"}}
	if f.IsMulti() {
		t.Fatal("empty multi must report no selection")
	}
	f.ToggleMulti("a")
	f.ToggleMulti("a") // toggle off
	if f.IsMulti() {
		t.Fatal("double toggle must clear selection")
	}
	f.ToggleMulti("b")
	if !f.IsMulti() {
		t.Fatal("expected a selection after adding b")
	}
}

func TestOpenPopupByKind(t *testing.T) {
	s := &FormState{}
	s.Open(FormSpec{
		Kind: FormContainerUpdate,
		Fields: []FormField{
			{Key: "mem", Kind: FormInt},
			{Key: "restart", Kind: FormSelect, Options: []string{"a", "b", "c"}, Index: 1},
			{Key: "dir", Kind: FormPath, Suggestions: []PathEntry{{Name: "x", Path: "/x"}}},
		},
	})
	s.FieldFocus = 1
	s.OpenPopup()
	if s.Popup.Kind != PopupSelect || s.Popup.Cursor != 1 || !s.Popup.Open {
		t.Fatalf("select popup = %#v, want PopupSelect cursor=1 open", s.Popup)
	}

	s.FieldFocus = 2
	s.OpenPopup()
	if s.Popup.Kind != PopupPath || !s.Popup.Open {
		t.Fatalf("path popup = %#v, want PopupPath open", s.Popup)
	}

	s.FieldFocus = 0
	s.OpenPopup()
	if s.Popup.Open {
		t.Fatal("text field must not open a popup")
	}
}

func TestPopupCountAndCursorWrap(t *testing.T) {
	s := &FormState{}
	s.Popup = FormPopupState{Kind: PopupSelect, Field: 0, Open: true}
	// To exercise PopupCount/Cursor independently of field focus, use a form
	// with the popup field at index 0.
	s.Fields = []FormField{{Kind: FormSelect, Options: []string{"a", "b", "c"}}}
	if got := s.PopupCount(); got != 3 {
		t.Fatalf("PopupCount = %d, want 3", got)
	}
	s.PopupCursor(1)
	if s.Popup.Cursor != 1 {
		t.Fatalf("cursor after +1 = %d, want 1", s.Popup.Cursor)
	}
	s.PopupCursor(-3) // wrap to same
	if s.Popup.Cursor != 1 {
		t.Fatalf("cursor wrap = %d, want 1", s.Popup.Cursor)
	}
	s.ClosePopup()
	if s.Popup.Open {
		t.Fatal("ClosePopup must clear the popup")
	}
}

func TestPopupOnMissingFieldNoPanic(t *testing.T) {
	s := &FormState{}
	s.Popup = FormPopupState{Kind: PopupPath, Field: 7, Open: true}
	if f := s.PopupField(); f != nil {
		t.Fatalf("expected nil popup field for out-of-range index, got %#v", f)
	}
	if n := s.PopupCount(); n != 0 {
		t.Fatalf("PopupCount on missing field = %d, want 0", n)
	}
}

func TestMultiPopupCommitsOrCancelsWorkingSelection(t *testing.T) {
	s := &FormState{Fields: []FormField{{Kind: FormMultiSelect, Options: []string{"a", "b"}, Selected: map[string]bool{"a": true}}}, FieldFocus: 0}
	s.OpenPopup()
	s.TogglePopupMulti("b")
	if s.Fields[0].Selected["b"] {
		t.Fatal("popup edits must not mutate the field before commit")
	}
	s.ClosePopup()
	if !s.Fields[0].Selected["a"] || s.Fields[0].Selected["b"] {
		t.Fatalf("cancel changed selection: %#v", s.Fields[0].Selected)
	}
	s.OpenPopup()
	s.TogglePopupMulti("b")
	s.CommitPopupMulti()
	if !s.Fields[0].Selected["a"] || !s.Fields[0].Selected["b"] {
		t.Fatalf("commit selection = %#v", s.Fields[0].Selected)
	}
}

func TestPopupCursorHomeEndPage(t *testing.T) {
	s := &FormState{
		Fields: []FormField{{Kind: FormSelect, Options: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}}},
	}
	s.Popup = FormPopupState{Kind: PopupSelect, Field: 0, Open: true, Cursor: 4}
	s.PopupCursorEnd()
	if s.Popup.Cursor != 9 {
		t.Fatalf("end cursor = %d, want 9", s.Popup.Cursor)
	}
	s.PopupCursorHome()
	if s.Popup.Cursor != 0 {
		t.Fatalf("home cursor = %d, want 0", s.Popup.Cursor)
	}
	s.PopupCursorPage(1, 3)
	if s.Popup.Cursor != 3 {
		t.Fatalf("PgDn page cursor = %d, want 3", s.Popup.Cursor)
	}
	s.PopupCursorPage(-1, 3)
	if s.Popup.Cursor != 0 {
		t.Fatalf("PgUp page cursor = %d, want 0", s.Popup.Cursor)
	}
	s.PopupCursorPage(1, 100)
	if s.Popup.Cursor != 9 {
		t.Fatalf("PgDn clamp = %d, want 9 (last)", s.Popup.Cursor)
	}
}

func TestFocusedButtonReportsSlot(t *testing.T) {
	s := &FormState{Fields: []FormField{{Key: "x", Kind: FormText}}}
	if got := s.FocusedButton(); got != "" {
		t.Fatalf("field focus must not be a button: %q", got)
	}
	s.FieldFocus = s.CancelSlot()
	if got := s.FocusedButton(); got != "cancel" {
		t.Fatalf("CancelSlot button = %q, want cancel", got)
	}
	s.FieldFocus = s.ConfirmSlot()
	if got := s.FocusedButton(); got != "confirm" {
		t.Fatalf("ConfirmSlot button = %q, want confirm", got)
	}
}

func TestOpenDangerousSpecDefaultsFocusToCancel(t *testing.T) {
	s := &FormState{}
	s.Open(FormSpec{
		Kind:      FormContainerRemove,
		Dangerous: true,
		Fields: []FormField{
			{Key: "force", Kind: FormBool},
			{Key: "removeVolumes", Kind: FormBool, DependsOn: "force", DependsEq: true},
		},
	})
	if s.FieldFocus != s.CancelSlot() {
		t.Fatalf("Dangerous form must open on Cancel: focus=%d, want %d", s.FieldFocus, s.CancelSlot())
	}
}

func TestOpenNonDangerousSpecDefaultsFocusToFirstField(t *testing.T) {
	s := &FormState{}
	s.Open(FormSpec{
		Kind: FormContainerCopy,
		Fields: []FormField{
			{Key: "source", Kind: FormPath},
			{Key: "destination", Kind: FormPath},
		},
	})
	if s.FieldFocus != 0 {
		t.Fatalf("non-dangerous form must open on first field: focus=%d, want 0", s.FieldFocus)
	}
}

func TestResetClearsFieldEditState(t *testing.T) {
	s := &FormState{}
	s.Open(FormSpec{
		Kind: FormContainerRemove,
		Fields: []FormField{
			{Key: "force", Kind: FormBool},
			{Key: "removeVolumes", Kind: FormBool, DependsOn: "force", DependsEq: true},
		},
	})
	force := s.Get("force")
	force.Toggle = true
	force.Touched = true
	volumes := s.Get("removeVolumes")
	volumes.Toggle = true
	volumes.Touched = true
	volumes.Error = "previous error"
	s.Popup = FormPopupState{Kind: PopupSelect, Field: 0, Open: true}
	s.Loading = true
	s.RecomputeVisibility()
	if volumes.Hidden {
		t.Fatal("setup: RemoveVolumes must be visible after Force toggle")
	}

	s.Reset()

	if force.Toggle || force.Touched {
		t.Fatalf("Reset must clear Force toggle/touched: toggle=%v touched=%v", force.Toggle, force.Touched)
	}
	if volumes.Toggle || volumes.Touched || volumes.Error != "" {
		t.Fatalf("Reset must clear RemoveVolumes state: toggle=%v touched=%v error=%q", volumes.Toggle, volumes.Touched, volumes.Error)
	}
	if s.Popup.Open || s.Loading {
		t.Fatalf("Reset must clear popup and loading: popup.Open=%v loading=%v", s.Popup.Open, s.Loading)
	}
}

func TestFormKindRemoveConstantsAreDistinct(t *testing.T) {
	if FormImageRemove == FormContainerRemove || FormContainerRemove == FormVolumeRemove || FormImageRemove == FormVolumeRemove {
		t.Fatalf("Remove FormKinds must be distinct: image=%d container=%d volume=%d", FormImageRemove, FormContainerRemove, FormVolumeRemove)
	}
}

func TestMoveFieldTraversesFieldsAndButtons(t *testing.T) {
	s := &FormState{Fields: []FormField{
		{Key: "a", Kind: FormText},
		{Key: "b", Kind: FormText},
		{Key: "c", Kind: FormText},
	}}
	s.FieldFocus = s.CancelSlot() // default Cancel
	s.MoveField(-1)               // Up from Cancel → Confirm
	if s.FocusedButton() != "confirm" {
		t.Fatalf("Up from Cancel = %s, want confirm", s.FocusedButton())
	}
	s.MoveField(-1) // Up from Confirm → last field (c)
	if s.FieldFocus != 2 {
		t.Fatalf("Up from Confirm = %d, want 2", s.FieldFocus)
	}
	s.MoveField(1) // Down from c → Confirm
	if s.FocusedButton() != "confirm" {
		t.Fatalf("Down from last = %s, want confirm", s.FocusedButton())
	}
	s.MoveField(1) // Down from Confirm → Cancel
	if s.FocusedButton() != "cancel" {
		t.Fatalf("Down from Confirm = %s, want cancel", s.FocusedButton())
	}
	s.MoveField(1) // Down from Cancel wraps to field 0
	if s.FieldFocus != 0 {
		t.Fatalf("Down from Cancel wraps = %d, want 0", s.FieldFocus)
	}
	s.MoveField(-1) // Up from field 0 wraps to Cancel
	if s.FocusedButton() != "cancel" {
		t.Fatalf("Up from first wraps = %s, want cancel", s.FocusedButton())
	}
}

func TestMoveButtonIsNoOp(t *testing.T) {
	s := &FormState{Fields: []FormField{{Kind: FormText}}}
	s.FieldFocus = s.ConfirmSlot()
	s.MoveButton(1)
	if s.FieldFocus != s.ConfirmSlot() {
		t.Fatalf("MoveButton must not change focus, got %d", s.FieldFocus)
	}
	s.MoveButton(-1)
	if s.FieldFocus != s.ConfirmSlot() {
		t.Fatalf("MoveButton must not change focus, got %d", s.FieldFocus)
	}
}

func TestToggleBoolIsSharedAndKindSafe(t *testing.T) {
	field := FormField{Kind: FormBool}
	if !field.ToggleBool() || !field.Toggle || !field.Touched {
		t.Fatalf("first toggle = %#v", field)
	}
	if !field.ToggleBool() || field.Toggle {
		t.Fatalf("second toggle = %#v", field)
	}
	text := FormField{Kind: FormText}
	if text.ToggleBool() || text.Toggle || text.Touched {
		t.Fatalf("non-Bool field changed: %#v", text)
	}
}

func TestRecomputeVisibilityHidesFieldUntilDependencyMatches(t *testing.T) {
	s := &FormState{}
	s.Open(FormSpec{
		Kind: FormContainerCommit,
		Fields: []FormField{
			{Key: "repo", Kind: FormText},
			{Key: "tar", Kind: FormBool},
			{Key: "archive", Kind: FormPath, DependsOn: "tar", DependsEq: true},
		},
	})
	if !s.Fields[2].Hidden {
		t.Fatal("archive must start hidden when tar=false")
	}
	s.Fields[1].Toggle = true
	s.RecomputeVisibility()
	if s.Fields[2].Hidden {
		t.Fatal("archive must become visible when tar=true")
	}
	s.Fields[1].Toggle = false
	s.RecomputeVisibility()
	if !s.Fields[2].Hidden {
		t.Fatal("archive must hide again when tar=false")
	}
}

func TestRecomputeVisibilityMovesFocusWhenFocusedFieldHides(t *testing.T) {
	s := &FormState{}
	s.Open(FormSpec{
		Kind: FormContainerCommit,
		Fields: []FormField{
			{Key: "repo", Kind: FormText},
			{Key: "tar", Kind: FormBool, Toggle: true},
			{Key: "archive", Kind: FormPath, DependsOn: "tar", DependsEq: true},
		},
	})
	s.FieldFocus = 2 // archive (currently visible)
	if s.Fields[2].Hidden {
		t.Fatal("setup: archive must be visible")
	}
	s.Fields[1].Toggle = false
	s.RecomputeVisibility()
	if s.FieldFocus == 2 {
		t.Fatal("focus must move off archive after it hides")
	}
	if s.FieldFocus >= 0 && s.FieldFocus < len(s.Fields) && s.Fields[s.FieldFocus].Hidden {
		t.Fatalf("focus landed on hidden field %d", s.FieldFocus)
	}
}

func TestMoveFieldSkipsHiddenFields(t *testing.T) {
	s := &FormState{Fields: []FormField{
		{Key: "a", Kind: FormText},
		{Key: "hidden", Kind: FormText, Hidden: true},
		{Key: "c", Kind: FormText},
	}, FieldFocus: 0}
	s.MoveField(1)
	if s.FieldFocus != 2 {
		t.Fatalf("Down from 0 must skip hidden idx=1, got %d", s.FieldFocus)
	}
	s.MoveField(-1)
	if s.FieldFocus != 0 {
		t.Fatalf("Up from 2 must skip hidden idx=1, got %d", s.FieldFocus)
	}
}

func TestFormFieldHelpers(t *testing.T) {
	if got := FormFieldMin(0.5); got == nil || *got != 0.5 {
		t.Fatalf("FormFieldMin(0.5) = %v, want pointer to 0.5", got)
	}
	if got := FormFieldMax(8192); got == nil || *got != 8192 {
		t.Fatalf("FormFieldMax(8192) = %v, want pointer to 8192", got)
	}
	// Helpers must not alias the same backing variable across calls; each
	// call must return a freshly-allocated pointer so field literals don't
	// accidentally share state through package-level captures.
	if FormFieldMin(1) == FormFieldMin(1) {
		t.Fatal("FormFieldMin must return a new pointer per call")
	}
}

func TestFormFieldStructFields(t *testing.T) {
	f := FormField{
		Key:         "mem",
		HelperText:  "Container memory limit",
		Unit:        "MB",
		Min:         FormFieldMin(64),
		Max:         FormFieldMax(65536),
		Placeholder: "1024",
	}
	if f.HelperText != "Container memory limit" {
		t.Fatalf("HelperText = %q, want %q", f.HelperText, "Container memory limit")
	}
	if f.Unit != "MB" {
		t.Fatalf("Unit = %q, want %q", f.Unit, "MB")
	}
	if f.Min == nil || *f.Min != 64 {
		t.Fatalf("Min = %v, want pointer to 64", f.Min)
	}
	if f.Max == nil || *f.Max != 65536 {
		t.Fatalf("Max = %v, want pointer to 65536", f.Max)
	}
	if f.Placeholder != "1024" {
		t.Fatalf("Placeholder = %q, want %q", f.Placeholder, "1024")
	}
}

func TestFormFieldJSONRoundTrip(t *testing.T) {
	orig := FormField{
		Key:        "mem",
		HelperText: "Memory limit",
		Unit:       "MB",
		Min:        FormFieldMin(64),
		Max:        FormFieldMax(8192),
	}
	body, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, want := range []string{`"helperText":"Memory limit"`, `"unit":"MB"`, `"min":64`, `"max":8192`} {
		if !strings.Contains(s, want) {
			t.Fatalf("marshal missing %q in %s", want, s)
		}
	}
	var got FormField
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.HelperText != orig.HelperText || got.Unit != orig.Unit || got.Placeholder != orig.Placeholder {
		t.Fatalf("string fields round-trip mismatch: %+v vs %+v", got, orig)
	}
	if got.Min == nil || *got.Min != *orig.Min {
		t.Fatalf("Min round-trip mismatch: %v vs %v", got.Min, orig.Min)
	}
	if got.Max == nil || *got.Max != *orig.Max {
		t.Fatalf("Max round-trip mismatch: %v vs %v", got.Max, orig.Max)
	}
}

func TestFormFieldJSONCEmptyFieldsOmitted(t *testing.T) {
	// A zero-valued FormField must serialize without the Phase-2 keys so
	// legacy layouts (no HelperText/Min/Max/...) do not gain noise.
	body, err := utils.MarshalJSONCStd(FormField{Key: "tag"}, utils.JSONCHeader{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(body)
	for _, banned := range []string{"helperText", `"unit"`, `"min"`, `"max"`, "placeholder"} {
		if strings.Contains(s, banned) {
			t.Fatalf("empty field must omit %q, got: %s", banned, s)
		}
	}
	// And a JSONC body with comments must round-trip into a populated struct.
	src := []byte(`{
		// phase-2 layout
		"key": "mem",
		"helperText": "Memory limit",
		"unit": "MB",
		"min": 64,
		"max": 8192,
		"placeholder": "1024"
	}`)
	var got FormField
	if err := utils.UnmarshalJSONCStd(src, &got); err != nil {
		t.Fatalf("unmarshal jsonc: %v", err)
	}
	if got.HelperText != "Memory limit" || got.Unit != "MB" || got.Placeholder != "1024" {
		t.Fatalf("string fields: %+v", got)
	}
	if got.Min == nil || *got.Min != 64 {
		t.Fatalf("Min = %v, want 64", got.Min)
	}
	if got.Max == nil || *got.Max != 8192 {
		t.Fatalf("Max = %v, want 8192", got.Max)
	}
}
