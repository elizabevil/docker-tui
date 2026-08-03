package state

import "testing"

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

func TestMoveFieldTraversesFieldsAndButtons(t *testing.T) {
	s := &FormState{Fields: []FormField{
		{Key: "a", Kind: FormText},
		{Key: "b", Kind: FormText},
		{Key: "c", Kind: FormText},
	}}
	s.FieldFocus = s.CancelSlot() // default Cancel
	s.MoveField(-1) // Up from Cancel → Confirm
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
