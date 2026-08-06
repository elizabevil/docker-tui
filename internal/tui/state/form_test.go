package state

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/utils"
)

// fakeField is a minimal FormField implementation for FormState tests. It
// holds a few mutable fields and lets the tests drive state through the
// public FormState API.
type fakeField struct {
	key, label, helperText, unit, placeholder string
	required                                bool
	hidden                                  bool
	touched, toggle                         bool
	index                                   int
	selected                                map[string]bool
	dependsOn                               string
	dependsEq                               bool
	text                                    string
	cursor                                  int
	errMsg                                  string
	kind                                    FormFieldKind
	options                                 []string
	dispOpts                                []string
	suggestions                             []PathEntry
	pathLoading                             bool
	showHidden                              bool
	pathError                               string
	pathTabInput                            string
	min, max                                *float64
}

func (f *fakeField) Kind() FormFieldKind      { return f.kind }
func (f *fakeField) Key() string              { return f.key }
func (f *fakeField) Hidden() bool             { return f.hidden }
func (f *fakeField) SetHidden(v bool)         { f.hidden = v }
func (f *fakeField) Error() string            { return f.errMsg }
func (f *fakeField) SetError(v string)        { f.errMsg = v }
func (f *fakeField) Touched() bool            { return f.touched }
func (f *fakeField) SetTouched(v bool)        { f.touched = v }
func (f *fakeField) Value() any               { return f.text }
func (f *fakeField) Reset() {
	f.text = ""
	f.cursor = 0
	f.touched = false
	f.errMsg = ""
	f.toggle = false
	f.index = 0
	f.selected = nil
}
func (f *fakeField) Text() string             { return f.text }
func (f *fakeField) SetText(s string)         { f.text = s }
func (f *fakeField) Cursor() int              { return f.cursor }
func (f *fakeField) SetCursor(c int)          { f.cursor = c }
func (f *fakeField) Toggle() bool             { return f.toggle }
func (f *fakeField) SetToggle(v bool)         { f.toggle = v }
func (f *fakeField) Options() []string        { return f.options }
func (f *fakeField) DisplayOptions() []string { return f.dispOpts }
func (f *fakeField) Index() int               { return f.index }
func (f *fakeField) SetIndex(i int)           { f.index = i }
func (f *fakeField) Selected() map[string]bool { return f.selected }
func (f *fakeField) SetSelected(v map[string]bool) { f.selected = v }
func (f *fakeField) DependsOn() string        { return f.dependsOn }
func (f *fakeField) DependsEq() bool          { return f.dependsEq }

func fakeText(key string) *fakeField {
	return &fakeField{kind: FormText, key: key, label: key}
}
func fakeBool(key string) *fakeField {
	return &fakeField{kind: FormBool, key: key, label: key}
}
func fakePath(key string) *fakeField {
	return &fakeField{kind: FormPath, key: key, label: key}
}
func fakeSelect(key string, opts ...string) *fakeField {
	return &fakeField{kind: FormSelect, key: key, label: key, options: opts}
}
func fakeMultiSelect(key string, opts ...string) *fakeField {
	return &fakeField{kind: FormMultiSelect, key: key, label: key, options: opts}
}

func TestOpenPopupByKind(t *testing.T) {
	s := &FormState{}
	s.Open(FormSpec{
		Kind: FormContainerUpdate,
		Fields: []FormField{
			fakeText("mem"),
			func() FormField {
				f := fakeSelect("restart", "a", "b", "c")
				f.index = 1
				return f
			}(),
			func() FormField {
				f := fakePath("dir")
				f.suggestions = []PathEntry{{Name: "x", Path: "/x"}}
				return f
			}(),
		},
	})
	if len(s.Fields) != 3 {
		t.Fatalf("Fields = %d, want 3", len(s.Fields))
	}
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
	s := &FormState{Fields: []FormField{fakeSelect("s", "a", "b", "c")}}
	s.Popup = FormPopupState{Kind: PopupSelect, Field: 0, Open: true}
	if got := s.PopupCount(); got != 3 {
		t.Fatalf("PopupCount = %d, want 3", got)
	}
	s.PopupCursor(1)
	if s.Popup.Cursor != 1 {
		t.Fatalf("cursor after +1 = %d, want 1", s.Popup.Cursor)
	}
	s.PopupCursor(-3)
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
	s := &FormState{Fields: []FormField{
		func() FormField {
			f := fakeMultiSelect("m", "a", "b")
			f.selected = map[string]bool{"a": true}
			return f
		}(),
	}, FieldFocus: 0}
	s.OpenPopup()
	s.TogglePopupMulti("b")
	if s.Fields[0].Selected()["b"] {
		t.Fatal("popup edits must not mutate the field before commit")
	}
	s.ClosePopup()
	if !s.Fields[0].Selected()["a"] || s.Fields[0].Selected()["b"] {
		t.Fatalf("cancel changed selection: %#v", s.Fields[0].Selected())
	}
	s.OpenPopup()
	s.TogglePopupMulti("b")
	s.CommitPopupMulti()
	if !s.Fields[0].Selected()["a"] || !s.Fields[0].Selected()["b"] {
		t.Fatalf("commit selection = %#v", s.Fields[0].Selected())
	}
}

func TestPopupCursorHomeEndPage(t *testing.T) {
	s := &FormState{Fields: []FormField{fakeSelect("s", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j")}}
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
	s := &FormState{Fields: []FormField{fakeText("x")}}
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
			fakeBool("force"),
			func() FormField {
				f := fakeBool("removeVolumes")
				f.dependsOn = "force"
				f.dependsEq = true
				return f
			}(),
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
			fakePath("source"),
			fakePath("destination"),
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
			fakeBool("force"),
			func() FormField {
				f := fakeBool("removeVolumes")
				f.dependsOn = "force"
				f.dependsEq = true
				return f
			}(),
		},
	})
	force := s.Get("force")
	force.SetToggle(true)
	force.SetTouched(true)
	volumes := s.Get("removeVolumes")
	volumes.SetToggle(true)
	volumes.SetTouched(true)
	volumes.SetError("previous error")
	s.Popup = FormPopupState{Kind: PopupSelect, Field: 0, Open: true}
	s.Loading = true
	s.RecomputeVisibility()
	if volumes.Hidden() {
		t.Fatal("setup: RemoveVolumes must be visible after Force toggle")
	}

	s.Reset()

	if force.Toggle() || force.Touched() {
		t.Fatalf("Reset must clear Force toggle/touched: toggle=%v touched=%v", force.Toggle(), force.Touched())
	}
	if volumes.Toggle() || volumes.Touched() || volumes.Error() != "" {
		t.Fatalf("Reset must clear RemoveVolumes state: toggle=%v touched=%v error=%q", volumes.Toggle(), volumes.Touched(), volumes.Error())
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
		fakeText("a"),
		fakeText("b"),
		fakeText("c"),
	}}
	s.FieldFocus = s.CancelSlot()
	s.MoveField(-1)
	if s.FocusedButton() != "confirm" {
		t.Fatalf("Up from Cancel = %s, want confirm", s.FocusedButton())
	}
	s.MoveField(-1)
	if s.FieldFocus != 2 {
		t.Fatalf("Up from Confirm = %d, want 2", s.FieldFocus)
	}
	s.MoveField(1)
	if s.FocusedButton() != "confirm" {
		t.Fatalf("Down from last = %s, want confirm", s.FocusedButton())
	}
	s.MoveField(1)
	if s.FocusedButton() != "cancel" {
		t.Fatalf("Down from Confirm = %s, want cancel", s.FocusedButton())
	}
	s.MoveField(1)
	if s.FieldFocus != 0 {
		t.Fatalf("Down from Cancel wraps = %d, want 0", s.FieldFocus)
	}
	s.MoveField(-1)
	if s.FocusedButton() != "cancel" {
		t.Fatalf("Up from first wraps = %s, want cancel", s.FocusedButton())
	}
}

func TestMoveButtonIsNoOp(t *testing.T) {
	s := &FormState{Fields: []FormField{fakeText("a")}}
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

func TestRecomputeVisibilityHidesFieldUntilDependencyMatches(t *testing.T) {
	s := &FormState{}
	s.Open(FormSpec{
		Kind: FormContainerCommit,
		Fields: []FormField{
			fakeText("repo"),
			fakeBool("tar"),
			func() FormField {
				f := fakePath("archive")
				f.dependsOn = "tar"
				f.dependsEq = true
				return f
			}(),
		},
	})
	if !s.Fields[2].Hidden() {
		t.Fatal("archive must start hidden when tar=false")
	}
	s.Get("tar").SetToggle(true)
	s.RecomputeVisibility()
	if s.Fields[2].Hidden() {
		t.Fatal("archive must become visible when tar=true")
	}
	s.Get("tar").SetToggle(false)
	s.RecomputeVisibility()
	if !s.Fields[2].Hidden() {
		t.Fatal("archive must hide again when tar=false")
	}
}

func TestRecomputeVisibilityMovesFocusWhenFocusedFieldHides(t *testing.T) {
	s := &FormState{}
	s.Open(FormSpec{
		Kind: FormContainerCommit,
		Fields: []FormField{
			fakeText("repo"),
			func() FormField {
				f := fakeBool("tar")
				f.toggle = true
				return f
			}(),
			func() FormField {
				f := fakePath("archive")
				f.dependsOn = "tar"
				f.dependsEq = true
				return f
			}(),
		},
	})
	s.FieldFocus = 2
	if s.Fields[2].Hidden() {
		t.Fatal("setup: archive must be visible")
	}
	s.Get("tar").SetToggle(false)
	s.RecomputeVisibility()
	if s.FieldFocus == 2 {
		t.Fatal("focus must move off archive after it hides")
	}
	if s.FieldFocus >= 0 && s.FieldFocus < len(s.Fields) && s.Fields[s.FieldFocus].Hidden() {
		t.Fatalf("focus landed on hidden field %d", s.FieldFocus)
	}
}

func TestMoveFieldSkipsHiddenFields(t *testing.T) {
	s := &FormState{Fields: []FormField{
		fakeText("a"),
		func() FormField {
			f := fakeText("hidden")
			f.SetHidden(true)
			return f
		}(),
		fakeText("c"),
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

var _ = json.Marshal
var _ = strings.Contains
var _ = utils.MarshalJSONCStd