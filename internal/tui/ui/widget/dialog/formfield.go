package dialog

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"

	"charm.land/lipgloss/v2"
)

// FormField is the per-kind interaction policy for a form field. Each kind
// has a concrete impl that owns its lean data (text + cursor + flags), the
// rendering, key handling, value access, and validation.
//
// UI-improvements §587-820 commit-4 architecture. Each impl declares only the
// fields it actually needs — the previous back-reference to *state.FormField
// was removed because state.FormField carries 25+ fields the impl never
// touches (Placeholder, DependsOn, PathTabInput, …) and the back-reference
// made every impl type structurally identical.
//
// Submit / form-state pipeline still reads from *state.FormField, so each impl
// exposes SyncTo to write its lean data back after every HandleKey.
type FormField interface {
	Kind() state.FormFieldKind
	Value() any
	SetValue(v any)
	Default() any
	Validate() error
	Reset()
	Render(focused bool, width int, cursorVisible ...bool) string
	HandleKey(key string, m *state.AppModel) (handled bool, updated bool)
	OpenPopup() (FormPopup, bool)
	HelpText() string
	SyncTo(field *state.FormField)
}

// FormPopup is the per-kind sub-overlay that a FormField can spawn.
type FormPopup interface {
	Title() string
	Size() (int, int)
	Render(width, height int) string
	HandleKey(key string) (handled bool, selected bool, selectedValue any)
	Highlighted() int
}

// Registry maps every FormFieldKind to a constructor that copies the
// *state.FormField source into a fresh lean impl. Add a kind here when you
// introduce a new FormFieldKind constant in state/form.go.
var Registry = map[state.FormFieldKind]func(*state.FormField) FormField{
	state.FormText:          func(src *state.FormField) FormField { return newTextField(src) },
	state.FormInt:           func(src *state.FormField) FormField { return newIntField(src) },
	state.FormSelect:        func(src *state.FormField) FormField { return newSelectField(src) },
	state.FormBool:          func(src *state.FormField) FormField { return newBoolField(src) },
	state.FormPath:          func(src *state.FormField) FormField { return newPathField(src) },
	state.FormMultiSelect:   func(src *state.FormField) FormField { return newMultiSelectField(src) },
	state.FormRadioGroup:    func(src *state.FormField) FormField { return newRadioGroupField(src) },
	state.FormTextMultiLine: func(src *state.FormField) FormField { return newTextMultiLineField(src) },
	state.FormTextPassword:  func(src *state.FormField) FormField { return newTextPasswordField(src) },
}

// FormFieldFor returns a fresh impl for f. Unknown kinds fall back to
// TextField so a stale form spec never crashes the dialog.
func FormFieldFor(f *state.FormField) FormField {
	if f == nil {
		return newTextField(nil)
	}
	if ctor, ok := Registry[f.Kind]; ok && ctor != nil {
		return ctor(f)
	}
	return newTextField(f)
}

func fitValue(value string, width int) string {
	return utils.FitVisible(value, max(1, width))
}

func appendUnitSuffix(value, unit string, width int) string {
	if unit == "" {
		return fitValue(value, width)
	}
	suffix := " " + unit
	return fitValue(value, max(1, width-utils.DisplayWidth(suffix))) + suffix
}

func appendDropdownMarker(value, marker string, width int) string {
	if marker == "" {
		return value
	}
	return fitValue(value, max(1, width-2)) + " " + marker
}

// applyEditKey maps a terminal key string to primitive text/cursor mutations.
// Mirrors keyboard/editQueryInput so dialog-side impls process keys identically.
// Syncs the transient state back on handled=true (cursor-only keys like Left
// don't change text but still advance the impl's cursor).
func applyEditKey(key string, text *string, cursor *int) (bool, bool) {
	if text == nil {
		return false, false
	}
	input := state.QueryInputState{Text: *text, Cursor: *cursor}
	handled, changed := applyEditKeyRaw(key, &input)
	if handled {
		*text = input.Text
		*cursor = input.Cursor
	}
	return handled, changed
}

func applyEditKeyRaw(key string, input *state.QueryInputState) (bool, bool) {
	if input == nil {
		return false, false
	}
	input.Clamp()
	switch key {
	case keys.KeyLeft, "ctrl+b":
		input.Move(-1)
		return true, false
	case keys.KeyRight, "ctrl+f":
		input.Move(1)
		return true, false
	case "alt+b":
		input.MoveWordBackward()
		return true, false
	case "alt+f":
		input.MoveWordForward()
		return true, false
	case "ctrl+a", keys.KeyHome:
		input.MoveHome()
		return true, false
	case "ctrl+e", keys.KeyEnd:
		input.MoveEnd()
		return true, false
	case keys.KeyBackspace, "ctrl+h":
		return true, input.DeleteBackward()
	case keys.KeyDelete:
		return true, input.DeleteForward()
	case "ctrl+u":
		return true, input.DeleteToStart()
	case "ctrl+k":
		return true, input.DeleteToEnd()
	case "ctrl+w":
		return true, input.DeleteWordBackward()
	default:
		if len([]rune(key)) != 1 {
			return false, false
		}
		return true, input.Insert(key)
	}
}

// --- TextField ------------------------------------------------------------

// TextField is a free-text input edited by the shared key map.
type TextField struct {
	text        string
	cursor      int
	touched     bool
	required    bool
	unit        string
	helperText  string
	placeholder string
	errMsg      string
}

func newTextField(src *state.FormField) *TextField {
	if src == nil {
		return &TextField{}
	}
	return &TextField{
		text:        src.Input.Text,
		cursor:      src.Input.Cursor,
		required:    src.Required,
		unit:        src.Unit,
		helperText:  src.HelperText,
		placeholder: src.Placeholder,
		errMsg:      src.Error,
	}
}

func (f *TextField) Kind() state.FormFieldKind { return state.FormText }
func (f *TextField) Value() any                { return f.text }
func (f *TextField) Default() any              { return "" }
func (f *TextField) HelpText() string          { return f.helperText }

func (f *TextField) SetValue(v any) {
	if s, ok := v.(string); ok {
		f.text = s
		f.cursor = len([]rune(s))
	}
}

func (f *TextField) Validate() error {
	if f.required && strings.TrimSpace(f.text) == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func (f *TextField) Reset() {
	f.text = ""
	f.cursor = 0
	f.touched = false
	f.errMsg = ""
}

func (f *TextField) Render(focused bool, width int, cursorVisible ...bool) string {
	value := renderTextCell(f.text, f.cursor, state.FormText, focused, width, cursorVisible...)
	return appendUnitSuffix(value, f.unit, width)
}

func (f *TextField) HandleKey(key string, _ *state.AppModel) (bool, bool) {
	h, c := applyEditKey(key, &f.text, &f.cursor)
	if c {
		f.touched = true
	}
	return h, c
}

func (f *TextField) OpenPopup() (FormPopup, bool) { return nil, false }

func (f *TextField) SyncTo(field *state.FormField) {
	field.Input.Text = f.text
	field.Input.Cursor = f.cursor
	field.Touched = f.touched
	field.Error = f.errMsg
}

// --- IntField -------------------------------------------------------------

// IntField is a free-text field restricted to digits + dec point on submit.
type IntField struct {
	text       string
	cursor     int
	touched    bool
	required   bool
	unit       string
	helperText string
	errMsg      string
	min        *float64
	max        *float64
}

func newIntField(src *state.FormField) *IntField {
	if src == nil {
		return &IntField{}
	}
	return &IntField{
		text:       src.Input.Text,
		cursor:     src.Input.Cursor,
		required:   src.Required,
		unit:       src.Unit,
		helperText: src.HelperText,
		errMsg:     src.Error,
		min:        src.Min,
		max:        src.Max,
	}
}

func (f *IntField) Kind() state.FormFieldKind { return state.FormInt }
func (f *IntField) Value() any                { return f.text }
func (f *IntField) Default() any              { return "" }
func (f *IntField) HelpText() string          { return f.helperText }

func (f *IntField) SetValue(v any) {
	if s, ok := v.(string); ok {
		f.text = s
		f.cursor = len([]rune(s))
	}
}

func (f *IntField) Validate() error {
	raw := strings.TrimSpace(f.text)
	if raw == "" {
		if f.required {
			return fmt.Errorf("value is required")
		}
		return nil
	}
	x, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fmt.Errorf("not a number: %q", raw)
	}
	if f.min != nil && x < *f.min {
		return fmt.Errorf("value below minimum %v", *f.min)
	}
	if f.max != nil && x > *f.max {
		return fmt.Errorf("value above maximum %v", *f.max)
	}
	return nil
}

func (f *IntField) Reset() {
	f.text = ""
	f.cursor = 0
	f.touched = false
	f.errMsg = ""
}

func (f *IntField) Render(focused bool, width int, cursorVisible ...bool) string {
	value := renderTextCell(f.text, f.cursor, state.FormInt, focused, width, cursorVisible...)
	return appendUnitSuffix(value, f.unit, width)
}

func (f *IntField) HandleKey(key string, _ *state.AppModel) (bool, bool) {
	h, c := applyEditKey(key, &f.text, &f.cursor)
	if c {
		f.touched = true
	}
	return h, c
}

func (f *IntField) OpenPopup() (FormPopup, bool) { return nil, false }

func (f *IntField) SyncTo(field *state.FormField) {
	field.Input.Text = f.text
	field.Input.Cursor = f.cursor
	field.Touched = f.touched
	field.Error = f.errMsg
}

// --- PathField ------------------------------------------------------------

// PathField is a path input with Tab completion (BR-041 §3.3) and Ctrl+H
// hidden-file toggle.
type PathField struct {
	text        string
	cursor      int
	touched     bool
	required    bool
	helperText  string
	errMsg      string
	showHidden  bool
	suggestions []state.PathEntry
	loading     bool
	tabInput    string
	pathError   string
	pathSource  state.PathSource
	pathMode    state.PathMode
}

func newPathField(src *state.FormField) *PathField {
	if src == nil {
		return &PathField{}
	}
	return &PathField{
		text:        src.Input.Text,
		cursor:      src.Input.Cursor,
		required:    src.Required,
		helperText:  src.HelperText,
		errMsg:      src.Error,
		touched:     src.Touched,
		showHidden:  src.ShowHidden,
		suggestions: src.Suggestions,
		loading:     src.PathLoading,
		tabInput:    src.PathTabInput,
		pathError:   src.PathError,
		pathSource:  src.PathSource,
		pathMode:    src.PathMode,
	}
}

func (f *PathField) Kind() state.FormFieldKind { return state.FormPath }
func (f *PathField) Value() any                { return f.text }
func (f *PathField) Default() any              { return "" }
func (f *PathField) HelpText() string          { return "path with Tab completion" }

func (f *PathField) SetValue(v any) {
	if s, ok := v.(string); ok {
		f.text = s
		f.cursor = len([]rune(s))
	}
}

func (f *PathField) Validate() error {
	if f.required && strings.TrimSpace(f.text) == "" {
		return fmt.Errorf("path is required")
	}
	return nil
}

func (f *PathField) Reset() {
	f.text = ""
	f.cursor = 0
	f.touched = false
	f.errMsg = ""
	f.suggestions = nil
	f.loading = false
	f.tabInput = ""
	f.showHidden = false
	f.pathError = ""
}

func (f *PathField) Render(focused bool, width int, cursorVisible ...bool) string {
	value := renderTextCell(f.text, f.cursor, state.FormPath, focused, width, cursorVisible...)
	return appendUnitSuffix(value, "", width)
}

func (f *PathField) HandleKey(key string, _ *state.AppModel) (bool, bool) {
	if key == keys.KeyCtrlH {
		f.showHidden = !f.showHidden
		f.suggestions = nil
		f.pathError = ""
		f.loading = false
		return true, false
	}
	h, c := applyEditKey(key, &f.text, &f.cursor)
	if c {
		f.touched = true
	}
	return h, c
}

func (f *PathField) OpenPopup() (FormPopup, bool) {
	if len(f.suggestions) == 0 && !f.loading {
		return nil, false
	}
	return &PathPopup{owner: f}, true
}

func (f *PathField) SyncTo(field *state.FormField) {
	field.Input.Text = f.text
	field.Input.Cursor = f.cursor
	field.Touched = f.touched
	field.Error = f.errMsg
	field.ShowHidden = f.showHidden
	field.Suggestions = f.suggestions
	field.PathLoading = f.loading
	field.PathTabInput = f.tabInput
	field.PathError = f.pathError
	field.PathSource = f.pathSource
	field.PathMode = f.pathMode
}

// --- SelectField ----------------------------------------------------------

// SelectField is a one-of-many choice opened into a single-column popup.
type SelectField struct {
	options        []string
	displayOptions []string
	index          int
	touched        bool
	required       bool
	helperText     string
	errMsg         string
}

func newSelectField(src *state.FormField) *SelectField {
	if src == nil {
		return &SelectField{}
	}
	return &SelectField{
		options:        src.Options,
		displayOptions: src.DisplayOptions,
		index:          src.Index,
		required:       src.Required,
		helperText:     src.HelperText,
		errMsg:         src.Error,
	}
}

func (f *SelectField) Kind() state.FormFieldKind { return state.FormSelect }
func (f *SelectField) Value() any                { return selectValue(f.options, f.index) }
func (f *SelectField) Default() any              { return selectDefault(f.options) }
func (f *SelectField) HelpText() string          { return "single choice dropdown" }

func (f *SelectField) SetValue(v any) {
	if s, ok := v.(string); ok {
		for i, opt := range f.options {
			if opt == s {
				f.index = i
				return
			}
		}
	}
}

func (f *SelectField) Validate() error {
	if (f.required && f.index < 0) || f.index >= len(f.options) {
		return fmt.Errorf("selection required")
	}
	return nil
}

func (f *SelectField) Reset() {
	f.index = 0
	f.touched = false
	f.errMsg = ""
}

func (f *SelectField) Render(focused bool, width int, _ ...bool) string {
	value, marker := renderSelectCellValue(f.options, f.displayOptions, f.index, nil, state.FormSelect, focused, width)
	return appendDropdownMarker(value, marker, width)
}

func (f *SelectField) HandleKey(string, *state.AppModel) (bool, bool) {
	return false, false
}

func (f *SelectField) OpenPopup() (FormPopup, bool) {
	if len(f.options) == 0 {
		return nil, false
	}
	return &SelectPopup{options: f.options, multi: false}, true
}

func (f *SelectField) SyncTo(field *state.FormField) {
	field.Index = f.index
	field.Touched = f.touched
	field.Error = f.errMsg
}

// --- MultiSelectField -----------------------------------------------------

// MultiSelectField is a many-of-many choice opened into a popup where Space
// toggles each item (BR-041 §8.2).
type MultiSelectField struct {
	options        []string
	displayOptions []string
	selected       map[string]bool
	touched        bool
	required       bool
	helperText     string
	errMsg         string
}

func newMultiSelectField(src *state.FormField) *MultiSelectField {
	if src == nil {
		return &MultiSelectField{}
	}
	return &MultiSelectField{
		options:        src.Options,
		displayOptions: src.DisplayOptions,
		selected:       src.Selected,
		required:       src.Required,
		helperText:     src.HelperText,
		errMsg:         src.Error,
	}
}

func (f *MultiSelectField) Kind() state.FormFieldKind { return state.FormMultiSelect }
func (f *MultiSelectField) Value() any {
	if f.selected == nil {
		return map[string]bool{}
	}
	return f.selected
}
func (f *MultiSelectField) Default() any     { return map[string]bool{} }
func (f *MultiSelectField) HelpText() string { return "multi-select dropdown" }

func (f *MultiSelectField) SetValue(v any) {
	if m, ok := v.(map[string]bool); ok {
		out := make(map[string]bool, len(m))
		for k, sel := range m {
			out[k] = sel
		}
		f.selected = out
	}
}

func (f *MultiSelectField) Validate() error { return nil }

func (f *MultiSelectField) Reset() {
	f.selected = nil
	f.touched = false
	f.errMsg = ""
}

func (f *MultiSelectField) Render(focused bool, width int, _ ...bool) string {
	value, marker := renderSelectCellValue(f.options, f.displayOptions, 0, f.selected, state.FormMultiSelect, focused, width)
	return appendDropdownMarker(value, marker, width)
}

func (f *MultiSelectField) HandleKey(string, *state.AppModel) (bool, bool) {
	return false, false
}

func (f *MultiSelectField) OpenPopup() (FormPopup, bool) {
	if len(f.options) == 0 {
		return nil, false
	}
	return &SelectPopup{options: f.options, selected: f.selected, multi: true}, true
}

func (f *MultiSelectField) SyncTo(field *state.FormField) {
	field.Selected = f.selected
	field.Touched = f.touched
	field.Error = f.errMsg
}

// --- BoolField ------------------------------------------------------------

// BoolField is a checkbox toggled with Space. Recomputes the dependency
// graph so child fields (DependsOn) show/hide correctly.
type BoolField struct {
	toggle     bool
	touched    bool
	required   bool
	helperText string
	errMsg     string
}

func newBoolField(src *state.FormField) *BoolField {
	if src == nil {
		return &BoolField{}
	}
	return &BoolField{
		toggle:     src.Toggle,
		required:   src.Required,
		helperText: src.HelperText,
		errMsg:     src.Error,
	}
}

func (f *BoolField) Kind() state.FormFieldKind { return state.FormBool }
func (f *BoolField) Value() any                { return f.toggle }
func (f *BoolField) Default() any              { return false }
func (f *BoolField) HelpText() string          { return "boolean checkbox" }

func (f *BoolField) SetValue(v any) {
	if b, ok := v.(bool); ok {
		f.toggle = b
		f.touched = true
	}
}

func (f *BoolField) Validate() error { return nil }

func (f *BoolField) Reset() {
	f.toggle = false
	f.touched = false
	f.errMsg = ""
}

func (f *BoolField) Render(focused bool, width int, _ ...bool) string {
	return fitValue(renderBoolCell(f.toggle, focused), width)
}

func (f *BoolField) HandleKey(key string, _ *state.AppModel) (bool, bool) {
	if key == keys.KeySpace || key == "space" {
		f.toggle = !f.toggle
		f.touched = true
		return true, true
	}
	if key == keys.KeyLeft || key == keys.KeyRight || len([]rune(key)) == 1 {
		return true, false
	}
	return false, false
}

func (f *BoolField) OpenPopup() (FormPopup, bool) { return nil, false }

func (f *BoolField) SyncTo(field *state.FormField) {
	field.Toggle = f.toggle
	field.Touched = f.touched
	field.Error = f.errMsg
}

// --- RadioGroupField ------------------------------------------------------

// RadioGroupField is an inline single-choice alternative to SelectField.
type RadioGroupField struct {
	options        []string
	displayOptions []string
	index          int
	touched        bool
	required       bool
	helperText     string
	errMsg         string
}

func newRadioGroupField(src *state.FormField) *RadioGroupField {
	if src == nil {
		return &RadioGroupField{}
	}
	return &RadioGroupField{
		options:        src.Options,
		displayOptions: src.DisplayOptions,
		index:          src.Index,
		required:       src.Required,
		helperText:     src.HelperText,
		errMsg:         src.Error,
	}
}

func (f *RadioGroupField) Kind() state.FormFieldKind { return state.FormRadioGroup }
func (f *RadioGroupField) Value() any                { return selectValue(f.options, f.index) }
func (f *RadioGroupField) Default() any              { return selectDefault(f.options) }
func (f *RadioGroupField) HelpText() string          { return "radio group (single choice)" }

func (f *RadioGroupField) SetValue(v any) {
	if s, ok := v.(string); ok {
		for i, opt := range f.options {
			if opt == s {
				f.index = i
				return
			}
		}
	}
}

func (f *RadioGroupField) Validate() error { return nil }

func (f *RadioGroupField) Reset() {
	f.index = 0
	f.touched = false
	f.errMsg = ""
}

func (f *RadioGroupField) Render(focused bool, width int, _ ...bool) string {
	label := selectLabel(f.options, f.displayOptions, f.index)
	value := utils.TruncateVisible(label, max(1, width-4))
	style := component.GetStyle(component.StyleDim).GetForeground()
	if focused {
		style = component.GetStyle(component.StylePanelTitle).GetForeground()
	}
	styled := lipgloss.NewStyle().Foreground(style).Bold(focused).Render(value)
	return fitValue(styled, max(1, width-4)) + " (" + component.MarkCheck + ")"
}

func (f *RadioGroupField) HandleKey(string, *state.AppModel) (bool, bool) {
	return false, false
}

func (f *RadioGroupField) OpenPopup() (FormPopup, bool) { return nil, false }

func (f *RadioGroupField) SyncTo(field *state.FormField) {
	field.Index = f.index
	field.Touched = f.touched
	field.Error = f.errMsg
}

// --- TextMultiLineField ---------------------------------------------------

// TextMultiLineField is a long-text input. The value cell renders the first
// line only — multi-line rendering is deferred until a form actually needs
// FormTextArea (UI-improvements P4).
type TextMultiLineField struct {
	text        string
	cursor      int
	touched     bool
	required    bool
	helperText  string
	placeholder string
	errMsg      string
}

func newTextMultiLineField(src *state.FormField) *TextMultiLineField {
	if src == nil {
		return &TextMultiLineField{}
	}
	return &TextMultiLineField{
		text:        src.Input.Text,
		cursor:      src.Input.Cursor,
		required:    src.Required,
		helperText:  src.HelperText,
		placeholder: src.Placeholder,
		errMsg:      src.Error,
	}
}

func (f *TextMultiLineField) Kind() state.FormFieldKind { return state.FormTextMultiLine }
func (f *TextMultiLineField) Value() any                { return f.text }
func (f *TextMultiLineField) Default() any              { return "" }
func (f *TextMultiLineField) HelpText() string          { return "multi-line text (first line shown)" }

func (f *TextMultiLineField) SetValue(v any) {
	if s, ok := v.(string); ok {
		f.text = s
		f.cursor = len([]rune(s))
	}
}

func (f *TextMultiLineField) Validate() error {
	if f.required && strings.TrimSpace(f.text) == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func (f *TextMultiLineField) Reset() {
	f.text = ""
	f.cursor = 0
	f.touched = false
	f.errMsg = ""
}

func (f *TextMultiLineField) Render(focused bool, width int, cursorVisible ...bool) string {
	first, _, _ := strings.Cut(f.text, "\n")
	cursor := f.cursor
	if r := len([]rune(first)); cursor > r {
		cursor = r
	}
	value := renderTextCell(first, cursor, state.FormTextMultiLine, focused, width, cursorVisible...)
	return appendUnitSuffix(value, "", width)
}

func (f *TextMultiLineField) HandleKey(key string, _ *state.AppModel) (bool, bool) {
	h, c := applyEditKey(key, &f.text, &f.cursor)
	if c {
		f.touched = true
	}
	return h, c
}

func (f *TextMultiLineField) OpenPopup() (FormPopup, bool) { return nil, false }

func (f *TextMultiLineField) SyncTo(field *state.FormField) {
	field.Input.Text = f.text
	field.Input.Cursor = f.cursor
	field.Touched = f.touched
	field.Error = f.errMsg
}

// --- TextPasswordField ----------------------------------------------------

// TextPasswordField masks the value with U+2022 BULLET. The raw value is
// preserved on the field; only the visual cell is masked.
type TextPasswordField struct {
	text        string
	cursor      int
	touched     bool
	required    bool
	helperText  string
	placeholder string
	errMsg      string
}

func newTextPasswordField(src *state.FormField) *TextPasswordField {
	if src == nil {
		return &TextPasswordField{}
	}
	return &TextPasswordField{
		text:        src.Input.Text,
		cursor:      src.Input.Cursor,
		required:    src.Required,
		helperText:  src.HelperText,
		placeholder: src.Placeholder,
		errMsg:      src.Error,
	}
}

func (f *TextPasswordField) Kind() state.FormFieldKind { return state.FormTextPassword }
func (f *TextPasswordField) Value() any                { return f.text }
func (f *TextPasswordField) Default() any              { return "" }
func (f *TextPasswordField) HelpText() string          { return "password (masked)" }

func (f *TextPasswordField) SetValue(v any) {
	if s, ok := v.(string); ok {
		f.text = s
		f.cursor = len([]rune(s))
	}
}

func (f *TextPasswordField) Validate() error {
	if f.required && f.text == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func (f *TextPasswordField) Reset() {
	f.text = ""
	f.cursor = 0
	f.touched = false
	f.errMsg = ""
}

func (f *TextPasswordField) Render(focused bool, width int, cursorVisible ...bool) string {
	masked := strings.Repeat(passwordBullet, len([]rune(f.text)))
	value := renderTextCell(masked, f.cursor, state.FormTextPassword, focused, width, cursorVisible...)
	return appendUnitSuffix(value, "", width)
}

func (f *TextPasswordField) HandleKey(key string, _ *state.AppModel) (bool, bool) {
	h, c := applyEditKey(key, &f.text, &f.cursor)
	if c {
		f.touched = true
	}
	return h, c
}

func (f *TextPasswordField) OpenPopup() (FormPopup, bool) { return nil, false }

func (f *TextPasswordField) SyncTo(field *state.FormField) {
	field.Input.Text = f.text
	field.Input.Cursor = f.cursor
	field.Touched = f.touched
	field.Error = f.errMsg
}

// passwordBullet is the U+2022 BULLET used to mask password input.
const passwordBullet = "•"

// --- Popup impls ----------------------------------------------------------

// SelectPopup adapts a Select or MultiSelect field's options / selected state
// into a popup renderer. The hot path uses FormPopupState + handleFormPopupKey;
// this adapter exists for future extensions.
type SelectPopup struct {
	options  []string
	selected map[string]bool
	multi    bool
	cursor   int
}

func (p *SelectPopup) Title() string { return "" }
func (p *SelectPopup) Size() (int, int) {
	rows := len(p.options)
	if rows > 8 {
		rows = 8
	}
	return 60, rows + 2
}
func (p *SelectPopup) Highlighted() int { return p.cursor }
func (p *SelectPopup) Render(width, height int) string {
	return ""
}

func (p *SelectPopup) HandleKey(key string) (bool, bool, any) {
	n := len(p.options)
	if n == 0 {
		return false, false, nil
	}
	switch key {
	case keys.KeyUp:
		p.cursor = (p.cursor - 1 + n) % n
		return true, false, nil
	case keys.KeyDown:
		p.cursor = (p.cursor + 1) % n
		return true, false, nil
	case keys.KeyHome:
		p.cursor = 0
		return true, false, nil
	case keys.KeyEnd:
		p.cursor = n - 1
		return true, false, nil
	case keys.KeyEnter, keys.KeySpace:
		if p.multi {
			return true, false, nil
		}
		if p.cursor < 0 || p.cursor >= n {
			return true, false, nil
		}
		return true, true, p.options[p.cursor]
	}
	return false, false, nil
}

// PathPopup adapts a PathField's suggestions into a popup renderer. The hot
// path uses FormPopupState + renderPathPopup; this adapter exists for future
// extensions.
type PathPopup struct {
	owner  *PathField
	cursor int
}

func (p *PathPopup) Title() string     { return "" }
func (p *PathPopup) Highlighted() int { return p.cursor }
func (p *PathPopup) Size() (int, int) { return 100, state.FormPopupVisibleRows + 5 }
func (p *PathPopup) Render(width, height int) string {
	return ""
}

func (p *PathPopup) HandleKey(key string) (bool, bool, any) {
	n := len(p.owner.suggestions)
	if n == 0 {
		return false, false, nil
	}
	switch key {
	case keys.KeyUp:
		p.cursor = (p.cursor - 1 + n) % n
		return true, false, nil
	case keys.KeyDown:
		p.cursor = (p.cursor + 1) % n
		return true, false, nil
	case keys.KeyHome:
		p.cursor = 0
		return true, false, nil
	case keys.KeyEnd:
		p.cursor = n - 1
		return true, false, nil
	case keys.KeyEnter:
		if p.cursor < 0 || p.cursor >= n {
			return true, false, nil
		}
		return true, true, p.owner.suggestions[p.cursor]
	}
	return false, false, nil
}

// --- Small helpers ---------------------------------------------------------

func selectValue(options []string, index int) string {
	if index < 0 || index >= len(options) {
		return ""
	}
	return options[index]
}

func selectDefault(options []string) string {
	if len(options) > 0 {
		return options[0]
	}
	return ""
}

func selectLabel(options []string, displayOptions []string, index int) string {
	if index < 0 || index >= len(options) {
		return ""
	}
	if index < len(displayOptions) && displayOptions[index] != "" {
		return displayOptions[index]
	}
	return options[index]
}