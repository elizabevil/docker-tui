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
	s.FieldFocus = -1
	if got := s.FocusedButton(); got != "cancel" {
		t.Fatalf("default button = %q, want cancel", got)
	}
	s.OnConfirm = true
	if got := s.FocusedButton(); got != "confirm" {
		t.Fatalf("OnConfirm button = %q, want confirm", got)
	}
}

func TestMoveFieldTraversesFieldsAndButtons(t *testing.T) {
	s := &FormState{Fields: []FormField{
		{Key: "a", Kind: FormText},
		{Key: "b", Kind: FormText},
		{Key: "c", Kind: FormText},
	}, FieldFocus: -1, OnConfirm: false}
	s.MoveField(-1) // Up from default Cancel → last field (index 2)
	if s.FieldFocus != 2 {
		t.Fatalf("Up from Cancel = %d, want 2", s.FieldFocus)
	}
	s.MoveField(-1)
	if s.FieldFocus != 1 {
		t.Fatalf("Up from 2 = %d, want 1", s.FieldFocus)
	}
	s.MoveField(1)
	if s.FieldFocus != 2 {
		t.Fatalf("Down to last = %d, want 2", s.FieldFocus)
	}
	s.MoveField(1) // last → button area, default Cancel
	if s.FieldFocus != -1 || s.OnConfirm {
		t.Fatalf("Down from last = focus=%d onConfirm=%v, want -1/false", s.FieldFocus, s.OnConfirm)
	}
	s.MoveField(1) // Down in button area stays
	if s.FieldFocus != -1 || s.OnConfirm {
		t.Fatalf("Down in buttons must stay: focus=%d onConfirm=%v", s.FieldFocus, s.OnConfirm)
	}
	s.MoveField(-1) // Up from buttons → last field
	if s.FieldFocus != 2 {
		t.Fatalf("Up from buttons = %d, want last field 2", s.FieldFocus)
	}
	s.FieldFocus = 0
	s.MoveField(-1) // Up from first → button area
	if s.FieldFocus != -1 || s.OnConfirm {
		t.Fatalf("Up from first = focus=%d onConfirm=%v, want -1/false", s.FieldFocus, s.OnConfirm)
	}
}

func TestMoveButtonTogglesCancelConfirm(t *testing.T) {
	s := &FormState{Fields: []FormField{{Kind: FormText}}}
	s.FieldFocus = -1
	if s.OnConfirm {
		t.Fatal("default must be Cancel")
	}
	s.MoveButton(1) // Right
	if !s.OnConfirm {
		t.Fatal("Right from Cancel must select Confirm")
	}
	s.MoveButton(-1) // Left
	if s.OnConfirm {
		t.Fatal("Left from Confirm must select Cancel")
	}
	s.MoveButton(-1)
	if s.OnConfirm {
		t.Fatal("Left at Cancel must stay on Cancel")
	}
	s.MoveButton(1)
	s.MoveButton(1)
	if !s.OnConfirm {
		t.Fatal("Right at Confirm must stay on Confirm")
	}
	s.MoveButton(0) // no-op
	if !s.OnConfirm {
		t.Fatal("zero delta must preserve selection")
	}
	// Field focus: MoveButton is a no-op.
	s.FieldFocus = 0
	s.OnConfirm = false
	s.MoveButton(1)
	if s.OnConfirm {
		t.Fatal("MoveButton on a field focus must not toggle")
	}
}
