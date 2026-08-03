package state

import (
	"strings"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// FormKind identifies the parameter form behind a complex container action.
// Forms are presented in a dedicated overlay (ModeContainerForm) and hold a
// focused field plus the Confirm / Cancel buttons in one Tab loop.
type FormKind int

const (
	FormNone FormKind = iota
	FormContainerCopy
	FormContainerUpdate
	FormContainerExport
	FormContainerCommit
	FormImageSave
	FormImageLoad
)

// FormFieldKind selects how a form field is edited and rendered.
type FormFieldKind int

const (
	// FormText is a free-text field edited with editQueryInput.
	FormText FormFieldKind = iota
	// FormInt is a free-text field restricted to digits on submit.
	FormInt
	// FormSelect is a one-of-many choice opened into a single-column popup.
	FormSelect
	// FormBool is a checkbox toggled with Space.
	FormBool
	// FormPath is a path field edited with Tab / Ctrl+Space completion
	// (BR-041 §7.2). PathSource selects the candidate provider.
	FormPath
	// FormMultiSelect is a many-of-many choice opened into a popup where
	// Space toggles each item (BR-041 §8.2).
	FormMultiSelect
)

// PopupKind identifies the sub-layer that is temporarily open above a form.
type PopupKind int

const (
	PopupNone PopupKind = iota
	// PopupSelect is a single-select popup for a FormSelect field.
	PopupSelect
	// PopupMulti is a multi-select popup for a FormMultiSelect field.
	PopupMulti
	// PopupPath lists path candidates for a FormPath field.
	PopupPath
)

// FormPopupState owns the transient popup opened above a form. It carries the
// popup type, which field owns it, and the cursor row inside the popup.
type FormPopupState struct {
	Kind            PopupKind
	Field           int // index into FormState.Fields
	Cursor          int // selected row within the popup options
	Open            bool
	PendingSelected map[string]bool
}

// FormPopupVisibleRows is shared by keyboard paging and popup rendering.
const FormPopupVisibleRows = 8

// FormField is a single row in a container-action form.
type FormField struct {
	Key      string
	Label    string
	Kind     FormFieldKind
	Input    QueryInputState
	Options  []string
	Index    int             // selected variant for FormSelect
	Selected map[string]bool // selected variants for FormMultiSelect
	Toggle   bool

	Required bool
	Error    string

	PathSource  PathSource
	PathMode    PathMode
	Suggestions []PathEntry
	PathLoading bool
	// PathTabInput records a completed Tab request that could not extend the
	// common prefix. Repeating Tab with the same input opens the candidate list.
	PathTabInput string

	Touched bool // true once the user manually edited the value
}

// Text returns the trimmed value of a text/Int field.
func (f *FormField) Text() string {
	return strings.TrimSpace(f.Input.Text)
}

// Option returns the currently selected variant of a Select field.
func (f *FormField) Option() string {
	if f.Index < 0 || f.Index >= len(f.Options) {
		return ""
	}
	return f.Options[f.Index]
}

// StepSelect moves a Select field to the next/previous option (wrap-around).
func (f *FormField) StepSelect(delta int) {
	if len(f.Options) == 0 {
		return
	}
	f.Index = (f.Index + delta%len(f.Options) + len(f.Options)) % len(f.Options)
}

// FormSpec is the atom used to open a form dialog with a known field layout.
type FormSpec struct {
	Kind       FormKind
	Title      string
	TargetID   string
	TargetName string
	Fields     []FormField
	CWD        string
}

// ContainerUpdateConfigLoaded carries the current limits fetched for an open
// Update form. ContainerID prevents applying stale inspect results.
type ContainerUpdateConfigLoaded struct {
	ContainerID string
	Detail      *runtimeapi.ContainerDetail
	Error       error
}

// FormState owns the active container-action form: its fields, the focused
// field, the Confirm / Cancel slots that follow the fields, and any popup.
type FormState struct {
	Kind       FormKind
	Title      string
	TargetID   string
	TargetName string
	Fields     []FormField
	FieldFocus int  // index into Fields; -1 selects the Confirm/Cancel slot
	OnConfirm  bool // when FieldFocus == -1, true selects Confirm, false Cancel
	Popup      FormPopupState
	CWD        string
	Loading    bool
}

// Open resets the form to a fresh state from a spec. The initial focus is the
// Cancel slot (OnConfirm=false) so the safest action is always selected.
func (s *FormState) Open(spec FormSpec) {
	*s = FormState{
		Kind:       spec.Kind,
		Title:      spec.Title,
		TargetID:   spec.TargetID,
		TargetName: spec.TargetName,
		Fields:     spec.Fields,
		FieldFocus: -1,
		OnConfirm:  false,
		CWD:        spec.CWD,
	}
}

// Field returns the currently focused field, or nil when the focus is on the
// Confirm/Cancel slot or there are no fields.
func (s *FormState) Field() *FormField {
	if s.FieldFocus < 0 || s.FieldFocus >= len(s.Fields) {
		return nil
	}
	return &s.Fields[s.FieldFocus]
}

// Get returns the field with the given key, or nil when no such field exists.
func (s *FormState) Get(key string) *FormField {
	for i := range s.Fields {
		if s.Fields[i].Key == key {
			return &s.Fields[i]
		}
	}
	return nil
}

// SlotCount is the total number of Tab stops: every field plus Confirm and
// Cancel.
func (s *FormState) SlotCount() int { return len(s.Fields) + 2 }

// Slot returns the focus index for Confirm and Cancel respectively.
func (s *FormState) ConfirmSlot() int { return len(s.Fields) }
func (s *FormState) CancelSlot() int  { return len(s.Fields) + 1 }

// MoveSlot advances focus by delta across fields + Confirm + Cancel, wrapping
// around. The result stays within [0, SlotCount-1] and is exposed as
// FieldFocus (-1 = focus not on a field) plus OnConfirm.
func (s *FormState) MoveSlot(delta int) {
	total := s.SlotCount()
	if total <= 0 {
		return
	}
	cur := 0
	if s.FieldFocus >= 0 {
		cur = s.FieldFocus
	} else if s.OnConfirm {
		cur = s.ConfirmSlot()
	} else {
		cur = s.CancelSlot()
	}
	next := (cur + delta%total + total) % total
	s.FieldFocus = -1
	s.OnConfirm = false
	if next < len(s.Fields) {
		s.FieldFocus = next
	} else if next == s.ConfirmSlot() {
		s.OnConfirm = true
	}
}

// Close clears the form back to its zero value.
func (s *FormState) Close() { *s = FormState{} }

// IsMulti reports whether a FormMultiSelect field has any selected option.
func (f *FormField) IsMulti() bool { return len(f.Selected) > 0 }

// ToggleMulti flips the given option in a FormMultiSelect field.
func (f *FormField) ToggleMulti(option string) {
	if f.Selected == nil {
		f.Selected = make(map[string]bool)
	}
	if f.Selected[option] {
		delete(f.Selected, option)
	} else {
		f.Selected[option] = true
	}
}

// ToggleBool is the single activation path for FormBool fields. Keyboard
// handlers call it only for Space so every Form shares the same Bool rule.
func (f *FormField) ToggleBool() bool {
	if f == nil || f.Kind != FormBool {
		return false
	}
	f.Toggle = !f.Toggle
	f.Touched = true
	return true
}

// OpenPopup opens a popup for the focused field. The cursor starts on the
// currently selected option (or the first row otherwise).
func (s *FormState) OpenPopup() {
	f := s.Field()
	if f == nil {
		return
	}
	switch f.Kind {
	case FormSelect:
		s.Popup = FormPopupState{Kind: PopupSelect, Field: s.FieldFocus, Cursor: f.Index, Open: true}
	case FormMultiSelect:
		pending := make(map[string]bool, len(f.Selected))
		for option, selected := range f.Selected {
			if selected {
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
func (s *FormState) PopupField() *FormField {
	if !s.Popup.Open {
		return nil
	}
	if s.Popup.Field < 0 || s.Popup.Field >= len(s.Fields) {
		return nil
	}
	return &s.Fields[s.Popup.Field]
}

// PopupCount is the number of rows a popup may navigate over.
func (s *FormState) PopupCount() int {
	f := s.PopupField()
	if f == nil {
		return 0
	}
	switch f.Kind {
	case FormSelect:
		return len(f.Options)
	case FormMultiSelect:
		return len(f.Options)
	case FormPath:
		return len(f.Suggestions)
	default:
		return 0
	}
}

// PopupCursor moves the popup cursor by delta, wrapping within the rows.
func (s *FormState) PopupCursor(delta int) {
	n := s.PopupCount()
	if n <= 0 {
		return
	}
	total := n
	s.Popup.Cursor = (s.Popup.Cursor + delta%total + total) % total
}

// PopupCursorHome moves the popup cursor to the first row.
func (s *FormState) PopupCursorHome() {
	if n := s.PopupCount(); n > 0 {
		s.Popup.Cursor = 0
	}
}

// PopupCursorEnd moves the popup cursor to the last row.
func (s *FormState) PopupCursorEnd() {
	if n := s.PopupCount(); n > 0 {
		s.Popup.Cursor = n - 1
	}
}

// PopupCursorPage moves the popup cursor by pageSize rows in the delta
// direction and clamps to the valid range (PgUp/PgDn semantics).
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

// ClosePopup closes the active popup but leaves the form open.
func (s *FormState) ClosePopup() { s.Popup = FormPopupState{} }

// FocusedButton reports which button slot currently owns the focus, or ""
// when focus is on a field. The default Cancel slot returns "cancel".
func (s *FormState) FocusedButton() string {
	if s.FieldFocus >= 0 {
		return ""
	}
	if s.OnConfirm {
		return "confirm"
	}
	return "cancel"
}

// MoveField moves field focus up or down with the rules from BR-041 §3.1:
// Down from the last field enters the button area at Cancel; Up from the
// first field enters the button area at Cancel; Up from the button area
// returns to the last field; Down in the button area is a no-op (does not
// cycle or execute). Entering the button area never auto-selects Confirm.
func (s *FormState) MoveField(delta int) {
	if len(s.Fields) == 0 {
		return
	}
	if s.FieldFocus < 0 {
		if delta < 0 {
			s.FieldFocus = len(s.Fields) - 1
			s.OnConfirm = false
		}
		return
	}
	if delta > 0 {
		if s.FieldFocus == len(s.Fields)-1 {
			s.FieldFocus = -1
			s.OnConfirm = false
			return
		}
		s.FieldFocus++
		return
	}
	if s.FieldFocus == 0 {
		s.FieldFocus = -1
		s.OnConfirm = false
		return
	}
	s.FieldFocus--
}

// MoveButton follows the spatial order Cancel (left), Confirm (right).
// Repeated movement at either edge stays on that edge.
func (s *FormState) MoveButton(delta int) {
	if s.FieldFocus >= 0 {
		return
	}
	if delta == 0 {
		return
	}
	s.OnConfirm = delta > 0
}

// TogglePopupMulti updates the popup's working selection. The field itself is
// unchanged until CommitPopupMulti, allowing Esc to cancel cleanly.
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
	f.Selected = make(map[string]bool, len(s.Popup.PendingSelected))
	for option, selected := range s.Popup.PendingSelected {
		if selected {
			f.Selected[option] = true
		}
	}
	s.ClosePopup()
}
