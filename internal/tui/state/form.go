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
	// FormImageRemove / FormContainerRemove / FormVolumeRemove replace
	// the 2-option ChoiceDialog with explicit Force + child-flag forms.
	// FormSpec marks them Dangerous so the dialog opens on Cancel.
	FormImageRemove
	FormContainerRemove
	FormVolumeRemove
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
	// DisplayOptions holds the localized label shown to the user for each
	// Options entry; rendered alongside the dropdown marker and in popup
	// rows. When non-empty and aligned with Options, the renderer uses it
	// instead of the raw key. Empty/nil falls back to Options (legacy).
	DisplayOptions []string
	Index          int             // selected variant for FormSelect
	Selected       map[string]bool // selected variants for FormMultiSelect
	Toggle         bool

	Required bool
	Error    string

	// HelperText is a short secondary label rendered below the input
	// (units, hints). Optional; empty string hides it.
	HelperText string `json:"helperText,omitempty"  yaml:"helperText,omitempty"`
	// Unit is a tiny suffix glyph shown next to numeric inputs
	// (e.g. "MB", "%"). Optional; empty string hides it.
	Unit string `json:"unit,omitempty"        yaml:"unit,omitempty"`
	// Min is the inclusive lower bound for numeric fields. nil = unbounded.
	// Use FormFieldMin(v) to build the pointer in field literals.
	Min *float64 `json:"min,omitempty"         yaml:"min,omitempty"`
	// Max is the inclusive upper bound for numeric fields. nil = unbounded.
	// Use FormFieldMax(v) to build the pointer in field literals.
	Max *float64 `json:"max,omitempty"         yaml:"max,omitempty"`
	// Placeholder is a hint shown when the input is empty (e.g. "1024").
	// Optional; empty string hides it.
	Placeholder string `json:"placeholder,omitempty" yaml:"placeholder,omitempty"`

	PathSource  PathSource
	PathMode    PathMode
	Suggestions []PathEntry
	PathLoading bool
	// PathTabInput records a completed Tab request that could not extend the
	// common prefix. Repeating Tab with the same input opens the candidate list.
	PathTabInput string

	Touched bool // true once the user manually edited the value

	// DependsOn names another field whose current value gates this field's
	// visibility. When DependsOn is empty, the field is always shown.
	DependsOn string
	// DependsEq is the bool value the dependency must hold for this field
	// to remain visible. Only meaningful when the dependency is a FormBool.
	DependsEq bool
	// Hidden is set by RecomputeVisibility and read by the renderer and
	// MoveField to skip this field. Mutating it directly is unsupported.
	Hidden bool

	// ShowHidden toggles dotfile visibility in path completion popups
	// (Ctrl+H inside the field).
	ShowHidden bool

	// PathError holds the last completion error message (e.g. permission
	// denied). The popup renders this verbatim when Suggestions is empty
	// and PathLoading is false.
	PathError string
}

// FormFieldMin returns a pointer to v, suitable for FormField.Min literals.
// Provided so call sites can write FormField{..., Min: FormFieldMin(1024)}
// without an intermediate local variable.
func FormFieldMin(v float64) *float64 { return &v }

// FormFieldMax returns a pointer to v, suitable for FormField.Max literals.
// Provided so call sites can write FormField{..., Max: FormFieldMax(8192)}
// without an intermediate local variable.
func FormFieldMax(v float64) *float64 { return &v }

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
	Kind         FormKind
	Title        string
	TargetID     string
	TargetName   string
	Fields       []FormField
	CWD          string
	ConfirmLabel string
	CancelLabel  string
	// Dangerous opens the form with the Cancel row focused by default
	// (instead of the first field), to reduce accidental confirmation
	// of destructive actions such as Remove / Force Delete. Open() honors
	// this when computing the initial FieldFocus.
	Dangerous bool
}

// ContainerUpdateConfigLoaded carries the current limits fetched for an open
// Update form. ContainerID prevents applying stale inspect results.
type ContainerUpdateConfigLoaded struct {
	ContainerID string
	Detail      *runtimeapi.ContainerDetail
	Error       error
}

// FormState owns the active container-action form: its fields, the focused
// row, the Confirm / Cancel slots that follow the fields, and any popup.
//
// FieldFocus is a single linear index over every focusable row
// (BR-043 §3.3 scheme B): values 0..len(Fields)-1 select fields,
// len(Fields) selects the Confirm row, len(Fields)+1 selects the Cancel row.
// Up/Down navigation wraps across the whole list.
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

// Open resets the form to a fresh state from a spec. The initial focus is the
// Cancel row for Dangerous forms (anti-misclick on destructive actions) and
// the first field otherwise (BR-043 §3.3 + UI-improvements A decision).
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
		initialFocus = 1 // Cancel row when there are no fields to focus first
	}
	if spec.Dangerous {
		initialFocus = len(spec.Fields) + 1 // Cancel row
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

// Reset clears every field's editable state (Input, Toggle, Selected,
// Touched, Error, Suggestions, Path loading/tab state) back to its declared
// default. The field list itself, Focus, Title, and Targets are preserved so
// the caller can decide whether to close the form. UI-improvements V decision
// — used by the Cancel handler so the next Open() sees clean state.
func (s *FormState) Reset() {
	for i := range s.Fields {
		f := &s.Fields[i]
		f.Input = QueryInputState{}
		f.Toggle = false
		f.Selected = nil
		f.Touched = false
		f.Error = ""
		f.Suggestions = nil
		f.PathLoading = false
		f.PathTabInput = ""
		f.ShowHidden = false
		f.PathError = ""
	}
	s.Popup = FormPopupState{}
	s.Loading = false
}

// Field returns the currently focused field, or nil when the focus is on the
// Confirm / Cancel row, the index is out of range, or the focused field is
// currently hidden (DependsOn not satisfied).
func (s *FormState) Field() *FormField {
	if s.FieldFocus < 0 || s.FieldFocus >= len(s.Fields) {
		return nil
	}
	if s.Fields[s.FieldFocus].Hidden {
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

// SlotCount is the total number of focusable rows: every field plus Confirm
// and Cancel (BR-043 §3.3).
func (s *FormState) SlotCount() int { return len(s.Fields) + 2 }

// ConfirmSlot returns the focus index of the Confirm row.
func (s *FormState) ConfirmSlot() int { return len(s.Fields) }

// CancelSlot returns the focus index of the Cancel row.
func (s *FormState) CancelSlot() int { return len(s.Fields) + 1 }

// Close clears the form back to its zero value.
func (s *FormState) Close() { *s = FormState{} }

// RecomputeVisibility updates each field's Hidden flag from its DependsOn /
// DependsEq configuration. If the currently focused field becomes hidden as
// a result, focus is moved to the next visible field; if no visible field
// exists, focus falls back to the Cancel row. Call this after Open and
// after any change to a field that other fields depend on (BR-043 §3.5).
func (s *FormState) RecomputeVisibility() {
	for i := range s.Fields {
		s.Fields[i].Hidden = s.fieldHidden(&s.Fields[i])
	}
	if s.FieldFocus < 0 || s.FieldFocus >= len(s.Fields) {
		return
	}
	if !s.Fields[s.FieldFocus].Hidden {
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
			if !s.Fields[cur].Hidden {
				s.FieldFocus = cur
				return
			}
		}
	}
	s.FieldFocus = s.CancelSlot()
}

// fieldHidden reports whether f should be hidden based on its DependsOn
// configuration. Unsupported dependency kinds default to visible.
func (s *FormState) fieldHidden(f *FormField) bool {
	if f.DependsOn == "" {
		return false
	}
	dep := s.Get(f.DependsOn)
	if dep == nil {
		return false
	}
	if dep.Kind == FormBool {
		return dep.Toggle != f.DependsEq
	}
	return false
}

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

// FocusedButton reports which button row currently owns the focus, or ""
// when focus is on a field row (BR-043 §3.3).
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

// MoveField moves focus up or down across every row with wrap-around
// (BR-043 §3.3 scheme B): fields 0..len(Fields)-1, then Confirm
// (len(Fields)), then Cancel (len(Fields)+1). Hidden fields are skipped
// (BR-043 §3.5); if the next non-hidden row would be a button row, focus
// lands directly on it. A negative FieldFocus (zero-value form) is
// interpreted as the Cancel row.
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
			if s.Fields[next].Hidden {
				cur = next
				continue
			}
		}
		s.FieldFocus = next
		return
	}
}

// MoveButton is a no-op kept for backward compatibility; Cancel/Confirm
// navigation is now handled by MoveField's linear row model (BR-043 §3.3).
func (s *FormState) MoveButton(delta int) { _ = delta }

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
