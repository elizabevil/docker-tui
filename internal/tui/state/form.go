package state

import (
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// FormKind identifies the parameter form behind a complex container action.
type FormKind int

const (
	FormNone FormKind = iota
	FormContainerCopy
	FormContainerUpdate
	FormContainerExport
	FormContainerCommit
	FormImageSave
	FormImageLoad
	FormImageRemove
	FormContainerRemove
	FormVolumeRemove
)

// FormFieldKind selects how a form field is edited and rendered.
type FormFieldKind int

const (
	FormText FormFieldKind = iota
	FormInt
	FormSelect
	FormBool
	FormPath
	FormMultiSelect
	FormRadioGroup
	FormTextMultiLine
	FormTextPassword
)

// PopupKind identifies the sub-layer that is temporarily open above a form.
type PopupKind int

const (
	PopupNone PopupKind = iota
	PopupSelect
	PopupMulti
	PopupPath
)

// FormPopupState owns the transient popup opened above a form.
type FormPopupState struct {
	Kind            PopupKind
	Field           int
	Cursor          int
	Open            bool
	PendingSelected map[string]bool
}

const FormPopupVisibleRows = 8

// FormField is the lean form-widget interface shared by every form kind. Each
// concrete impl (TextField, IntField, BoolField, SelectField, MultiSelectField,
// PathField, RadioGroupField, TextMultiLineField, TextPasswordField) carries
// only the fields it actually needs. PathField is a SPECIAL form — its
// path-specific methods (PathSource / Suggestions / PathLoading / PathError /
// PathTabInput / ShowHidden) live on *PathField directly, NOT on this
// interface, so non-path kinds don't pay for them.
type FormField interface {
	// Identity.
	Kind() FormFieldKind
	Key() string

	// Derived.
	Hidden() bool
	SetHidden(bool)

	// Mutable.
	Error() string
	SetError(string)
	Touched() bool
	SetTouched(bool)

	// Submit value.
	Value() any

	// Lifecycle.
	Reset()

	// Text-like (Text/Int/Path/MultiLine/Password). Empty for non-text kinds.
	Text() string
	SetText(string)
	Cursor() int
	SetCursor(int)

	// Bool.
	Toggle() bool
	SetToggle(bool)

	// Select / radio.
	Options() []string
	DisplayOptions() []string
	Index() int
	SetIndex(int)

	// Multi-select.
	Selected() map[string]bool
	SetSelected(map[string]bool)
}

// FormSpec is the atom used to open a form dialog. Fields is a list of impls
// (constructed by the caller via dialog.NewXxxField).
type FormSpec struct {
	Kind         FormKind
	Title        string
	TargetID     string
	TargetName   string
	Fields       []FormField
	CWD          string
	ConfirmLabel string
	CancelLabel  string
	Dangerous    bool
}

// ContainerUpdateConfigLoaded carries the current limits fetched for an open
// Update form.
type ContainerUpdateConfigLoaded struct {
	ContainerID string
	Detail      *runtimeapi.ContainerDetail
	Error       error
}

// FormState owns the active form: its fields (impls), focused row, Confirm /
// Cancel slots, and any popup.
type FormState struct {
	Kind         FormKind
	Title        string
	TargetID     string
	TargetName   string
	Fields       []FormField
	FieldFocus   int
	ConfirmLabel string
	CancelLabel  string
	Popup        FormPopupState
	CWD          string
	Loading      bool
}

// Open resets the form to a fresh state from a spec.
func (s *FormState) Open(spec FormSpec) {
	confirm := spec.ConfirmLabel
	if confirm == "" {
		confirm = "Confirm"
	}
	cancel := spec.CancelLabel
	if cancel == "" {
		cancel = "Cancel"
	}
	initialFocus := 0
	if len(spec.Fields) == 0 {
		initialFocus = 1
	}
	if spec.Dangerous {
		initialFocus = len(spec.Fields) + 1
	}
	*s = FormState{
		Kind:         spec.Kind,
		Title:        spec.Title,
		TargetID:     spec.TargetID,
		TargetName:   spec.TargetName,
		Fields:       spec.Fields,
		ConfirmLabel: confirm,
		CancelLabel:  cancel,
		FieldFocus:   initialFocus,
		CWD:          spec.CWD,
	}
	s.RecomputeVisibility()
}

// Reset clears every field's editable state back to its declared default.
func (s *FormState) Reset() {
	for i := range s.Fields {
		s.Fields[i].Reset()
	}
	s.Popup = FormPopupState{}
	s.Loading = false
}

// Field returns the currently focused field, or nil when focus is on a
// button row or the field is hidden.
func (s *FormState) Field() FormField {
	if s.FieldFocus < 0 || s.FieldFocus >= len(s.Fields) {
		return nil
	}
	if s.Fields[s.FieldFocus].Hidden() {
		return nil
	}
	return s.Fields[s.FieldFocus]
}

// Get returns the field with the given key.
func (s *FormState) Get(key string) FormField {
	for i := range s.Fields {
		if s.Fields[i].Key() == key {
			return s.Fields[i]
		}
	}
	return nil
}

// SlotCount is the total number of focusable rows.
func (s *FormState) SlotCount() int { return len(s.Fields) + 2 }

func (s *FormState) ConfirmSlot() int { return len(s.Fields) }
func (s *FormState) CancelSlot() int  { return len(s.Fields) + 1 }

func (s *FormState) Close() { *s = FormState{} }

// RecomputeVisibility updates each field's Hidden flag from its
// DependsOn / DependsEq configuration. PathField exposes DependsOn /
// DependsEq via type assertion (it satisfies state.FormField for the common
// methods and has its own dependency config).
func (s *FormState) RecomputeVisibility() {
	for i := range s.Fields {
		s.Fields[i].SetHidden(s.fieldHidden(s.Fields[i]))
	}
	if s.FieldFocus < 0 || s.FieldFocus >= len(s.Fields) {
		return
	}
	if !s.Fields[s.FieldFocus].Hidden() {
		return
	}
	if n := len(s.Fields); n > 0 {
		cur := s.FieldFocus
		for i := 0; i < n; i++ {
			cur++
			if cur >= n {
				s.FieldFocus = s.CancelSlot()
				return
			}
			if !s.Fields[cur].Hidden() {
				s.FieldFocus = cur
				return
			}
		}
	}
	s.FieldFocus = s.CancelSlot()
}

// fieldHidden reports whether f should be hidden based on its DependsOn /
// DependsEq configuration. Non-PathField kinds do not support dependencies
// in this revision.
func (s *FormState) fieldHidden(f FormField) bool {
	type depCarrier interface {
		DependsOn() string
		DependsEq() bool
	}
	type toggleCarrier interface {
		Toggle() bool
	}
	d, ok := f.(depCarrier)
	if !ok || d.DependsOn() == "" {
		return false
	}
	dep := s.Get(d.DependsOn())
	if dep == nil {
		return false
	}
	if t, ok := dep.(toggleCarrier); ok {
		return t.Toggle() != d.DependsEq()
	}
	return false
}

// MoveField moves focus up or down across every row with wrap-around.
func (s *FormState) MoveField(delta int) {
	total := s.SlotCount()
	if total <= 0 {
		return
	}
	cur := s.FieldFocus
	if cur < 0 {
		cur = s.CancelSlot()
	}
	for i := 0; i < total; i++ {
		next := (cur + delta + total) % total
		if next < len(s.Fields) {
			if s.Fields[next].Hidden() {
				cur = next
				continue
			}
		}
		s.FieldFocus = next
		return
	}
}

func (s *FormState) MoveButton(delta int) { _ = delta }

// FocusedButton reports which button row currently owns the focus.
func (s *FormState) FocusedButton() string {
	switch s.FieldFocus {
	case s.ConfirmSlot():
		return "confirm"
	case s.CancelSlot():
		return "cancel"
	default:
		return ""
	}
}

// OpenPopup opens a popup for the focused field.
func (s *FormState) OpenPopup() {
	f := s.Field()
	if f == nil {
		return
	}
	switch f.Kind() {
	case FormSelect:
		s.Popup = FormPopupState{Kind: PopupSelect, Field: s.FieldFocus, Cursor: f.Index(), Open: true}
	case FormMultiSelect:
		selected := f.Selected()
		pending := make(map[string]bool, len(selected))
		for option, sel := range selected {
			if sel {
				pending[option] = true
			}
		}
		s.Popup = FormPopupState{Kind: PopupMulti, Field: s.FieldFocus, Open: true, PendingSelected: pending}
	case FormPath:
		s.Popup = FormPopupState{Kind: PopupPath, Field: s.FieldFocus, Open: true}
	default:
		s.Popup = FormPopupState{}
	}
}

// PopupField returns the field owning the active popup, or nil.
func (s *FormState) PopupField() FormField {
	if !s.Popup.Open {
		return nil
	}
	if s.Popup.Field < 0 || s.Popup.Field >= len(s.Fields) {
		return nil
	}
	return s.Fields[s.Popup.Field]
}

// PopupCount is the number of rows a popup may navigate over.
func (s *FormState) PopupCount() int {
	f := s.PopupField()
	if f == nil {
		return 0
	}
	switch f.Kind() {
	case FormSelect:
		return len(f.Options())
	case FormMultiSelect:
		return len(f.Options())
	case FormPath:
		type sugCarrier interface {
			Suggestions() []PathEntry
		}
		if s, ok := f.(sugCarrier); ok {
			return len(s.Suggestions())
		}
		return 0
	default:
		return 0
	}
}

func (s *FormState) PopupCursor(delta int) {
	n := s.PopupCount()
	if n <= 0 {
		return
	}
	s.Popup.Cursor = (s.Popup.Cursor + delta%n + n) % n
}

func (s *FormState) PopupCursorHome() {
	if n := s.PopupCount(); n > 0 {
		s.Popup.Cursor = 0
	}
}

func (s *FormState) PopupCursorEnd() {
	if n := s.PopupCount(); n > 0 {
		s.Popup.Cursor = n - 1
	}
}

func (s *FormState) PopupCursorPage(delta, pageSize int) {
	if pageSize < 1 {
		pageSize = 1
	}
	n := s.PopupCount()
	if n <= 0 {
		return
	}
	next := s.Popup.Cursor + delta*pageSize
	if next < 0 {
		next = 0
	} else if next >= n {
		next = n - 1
	}
	s.Popup.Cursor = next
}

func (s *FormState) ClosePopup() { s.Popup = FormPopupState{} }

// TogglePopupMulti updates the popup's working selection.
func (s *FormState) TogglePopupMulti(option string) {
	if !s.Popup.Open || s.Popup.Kind != PopupMulti {
		return
	}
	if s.Popup.PendingSelected == nil {
		s.Popup.PendingSelected = make(map[string]bool)
	}
	if s.Popup.PendingSelected[option] {
		delete(s.Popup.PendingSelected, option)
	} else {
		s.Popup.PendingSelected[option] = true
	}
}

// CommitPopupMulti applies the working selection and closes the popup.
func (s *FormState) CommitPopupMulti() {
	f := s.PopupField()
	if f == nil || s.Popup.Kind != PopupMulti {
		s.ClosePopup()
		return
	}
	pending := make(map[string]bool, len(s.Popup.PendingSelected))
	for option, sel := range s.Popup.PendingSelected {
		if sel {
			pending[option] = true
		}
	}
	f.SetSelected(pending)
	s.ClosePopup()
}

// FormMin / FormMax are pointer factories for numeric Min / Max fields.
func FormMin(v float64) *float64 { return &v }
func FormMax(v float64) *float64 { return &v }