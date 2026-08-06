package dialog

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"

	"charm.land/lipgloss/v2"
)

// FormField is the dialog-side form widget. It extends state.FormField with
// the label/HelperText config getters used by the form-row renderer, plus
// rendering, popup, and help-text. Path-specific methods (PathSource,
// Suggestions, …) live on *PathField directly, not on this interface.
type FormField interface {
	state.FormField
	Label() string
	HelperText() string
	Render(focused bool, width int, cursorVisible ...bool) string
	HandleKey(key string) (handled, updated bool)
	OpenPopup() (FormPopup, bool)
	HelpText() string
}

// FormPopup is the per-kind sub-overlay that a FormField can spawn.
type FormPopup interface {
	Title() string
	Size() (int, int)
	Render(width, height int) string
	HandleKey(key string) (handled, selected bool, selectedValue any)
	Highlighted() int
}

const passwordBullet = "•"

// applyEditKey maps a terminal key to text/cursor mutations.
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

func fitValue(value string, width int) string {
	return utils.FitVisible(value, max(1, width))
}

func appendUnitSuffix(value, unit string, width int) string {
	if unit == "" {
		// renderTextCell already fitted the value (truncated single-line for
		// text kinds, multi-line wrapped for paths). Re-truncating here would
		// collapse wrapped output — return it unchanged.
		return value
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

// --- TextField ------------------------------------------------------------

// TextField is a free-text input.
type TextField struct {
	key, label, helperText, unit, placeholder string
	required                                  bool
	hidden                                    bool
	errMsg                                    string
	text                                      string
	cursor                                    int
	touched                                   bool
}

// TextFieldConfig builds a TextField.
type TextFieldConfig struct {
	Key, Label, HelperText, Unit, Placeholder string
	Required                                  bool
	Text                                      string
	Cursor                                    int
}

// NewTextField constructs a TextField impl.
func NewTextField(cfg TextFieldConfig) state.FormField {
	return &TextField{
		key:         cfg.Key,
		label:       cfg.Label,
		helperText:  cfg.HelperText,
		required:    cfg.Required,
		unit:        cfg.Unit,
		placeholder: cfg.Placeholder,
		text:        cfg.Text,
		cursor:      cfg.Cursor,
	}
}

func (f *TextField) Kind() state.FormFieldKind      { return state.FormText }
func (f *TextField) Key() string                   { return f.key }
func (f *TextField) Hidden() bool                  { return f.hidden }
func (f *TextField) SetHidden(v bool)              { f.hidden = v }
func (f *TextField) Error() string                 { return f.errMsg }
func (f *TextField) SetError(v string)             { f.errMsg = v }
func (f *TextField) Touched() bool                 { return f.touched }
func (f *TextField) SetTouched(v bool)             { f.touched = v }
func (f *TextField) Value() any                    { return strings.TrimSpace(f.text) }
func (f *TextField) Reset()                        { f.text = ""; f.cursor = 0; f.touched = false; f.errMsg = "" }
func (f *TextField) Text() string                  { return f.text }
func (f *TextField) SetText(s string)              { f.text = s }
func (f *TextField) Cursor() int                   { return f.cursor }
func (f *TextField) SetCursor(c int)               { f.cursor = c }
func (f *TextField) Toggle() bool                  { return false }
func (f *TextField) SetToggle(_ bool)              {}
func (f *TextField) Options() []string            { return nil }
func (f *TextField) DisplayOptions() []string     { return nil }
func (f *TextField) Index() int                   { return 0 }
func (f *TextField) SetIndex(_ int)               {}
func (f *TextField) Selected() map[string]bool    { return nil }
func (f *TextField) SetSelected(_ map[string]bool) {}

func (f *TextField) Label() string   { return f.label }
func (f *TextField) HelperText() string { return f.helperText }
func (f *TextField) HelpText() string { return f.helperText }

func (f *TextField) Render(focused bool, width int, cursorVisible ...bool) string {
	value := renderTextCell(f.text, f.cursor, state.FormText, focused, width, cursorVisible...)
	return appendUnitSuffix(value, f.unit, width)
}

func (f *TextField) HandleKey(key string) (bool, bool) {
	h, c := applyEditKey(key, &f.text, &f.cursor)
	if c {
		f.touched = true
	}
	return h, c
}

func (f *TextField) OpenPopup() (FormPopup, bool) { return nil, false }

// --- IntField -------------------------------------------------------------

// IntField is a free-text field restricted to digits + dec point on submit.
type IntField struct {
	key, label, helperText, unit, placeholder string
	required                                  bool
	hidden                                    bool
	errMsg                                    string
	text                                      string
	cursor                                    int
	touched                                   bool
	min, max                                  *float64
}

// IntFieldConfig builds an IntField.
type IntFieldConfig struct {
	Key, Label, HelperText, Unit, Placeholder string
	Required                                  bool
	Min, Max                                  *float64
	Text                                      string
	Cursor                                    int
}

func NewIntField(cfg IntFieldConfig) state.FormField {
	return &IntField{
		key:         cfg.Key,
		label:       cfg.Label,
		helperText:  cfg.HelperText,
		required:    cfg.Required,
		unit:        cfg.Unit,
		placeholder: cfg.Placeholder,
		text:        cfg.Text,
		cursor:      cfg.Cursor,
		min:         cfg.Min,
		max:         cfg.Max,
	}
}

func (f *IntField) Kind() state.FormFieldKind      { return state.FormInt }
func (f *IntField) Key() string                   { return f.key }
func (f *IntField) Hidden() bool                  { return f.hidden }
func (f *IntField) SetHidden(v bool)              { f.hidden = v }
func (f *IntField) Error() string                 { return f.errMsg }
func (f *IntField) SetError(v string)             { f.errMsg = v }
func (f *IntField) Touched() bool                 { return f.touched }
func (f *IntField) SetTouched(v bool)             { f.touched = v }
func (f *IntField) Value() any                    { return strings.TrimSpace(f.text) }
func (f *IntField) Reset()                        { f.text = ""; f.cursor = 0; f.touched = false; f.errMsg = "" }
func (f *IntField) Text() string                  { return f.text }
func (f *IntField) SetText(s string)              { f.text = s }
func (f *IntField) Cursor() int                   { return f.cursor }
func (f *IntField) SetCursor(c int)               { f.cursor = c }
func (f *IntField) Toggle() bool                  { return false }
func (f *IntField) SetToggle(_ bool)              {}
func (f *IntField) Options() []string            { return nil }
func (f *IntField) DisplayOptions() []string     { return nil }
func (f *IntField) Index() int                   { return 0 }
func (f *IntField) SetIndex(_ int)               {}
func (f *IntField) Selected() map[string]bool    { return nil }
func (f *IntField) SetSelected(_ map[string]bool) {}
func (f *IntField) Min() *float64                 { return f.min }
func (f *IntField) Max() *float64                 { return f.max }

func (f *IntField) Label() string   { return f.label }
func (f *IntField) HelperText() string { return f.helperText }
func (f *IntField) HelpText() string { return f.helperText }

func (f *IntField) Render(focused bool, width int, cursorVisible ...bool) string {
	value := renderTextCell(f.text, f.cursor, state.FormInt, focused, width, cursorVisible...)
	return appendUnitSuffix(value, f.unit, width)
}

func (f *IntField) HandleKey(key string) (bool, bool) {
	h, c := applyEditKey(key, &f.text, &f.cursor)
	if c {
		f.touched = true
	}
	return h, c
}

func (f *IntField) OpenPopup() (FormPopup, bool) { return nil, false }

// Validate returns a parse / range error. Not on the FormField interface;
// submit handlers call directly when needed.
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

// --- PathField ------------------------------------------------------------

// PathField is a SPECIAL form for path inputs. It carries the common
// text-editing state plus path-specific state. The path-specific methods
// (PathSource, Suggestions, PathLoading, …) live on *PathField directly —
// they are NOT on the FormField interface. Callers type-assert to
// *PathField to access them.
type PathField struct {
	key, label, helperText string
	required               bool
	hidden                 bool
	errMsg                 string

	text        string
	cursor      int
	touched     bool
	showHidden  bool
	suggestions []state.PathEntry
	loading     bool
	tabInput    string
	pathError   string
	pathSource  state.PathSource
	pathMode    state.PathMode
	dependsOn   string
	dependsEq   bool
}

// PathFieldConfig builds a PathField.
type PathFieldConfig struct {
	Key, Label, HelperText string
	Required               bool
	PathSource             state.PathSource
	PathMode               state.PathMode
	Text                   string
	Cursor                 int
	InitialSuggestions     []state.PathEntry
	InitialLoading         bool
	InitialShowHidden      bool
	InitialTabInput        string
	InitialPathError       string
	DependsOn              string
	DependsEq              bool
}

func NewPathField(cfg PathFieldConfig) state.FormField {
	return &PathField{
		key:         cfg.Key,
		label:       cfg.Label,
		helperText:  cfg.HelperText,
		required:    cfg.Required,
		text:        cfg.Text,
		cursor:      cfg.Cursor,
		pathSource:  cfg.PathSource,
		pathMode:    cfg.PathMode,
		showHidden:  cfg.InitialShowHidden,
		suggestions: cfg.InitialSuggestions,
		loading:     cfg.InitialLoading,
		tabInput:    cfg.InitialTabInput,
		pathError:   cfg.InitialPathError,
		dependsOn:   cfg.DependsOn,
		dependsEq:   cfg.DependsEq,
	}
}

func (f *PathField) Kind() state.FormFieldKind { return state.FormPath }
func (f *PathField) Key() string              { return f.key }
func (f *PathField) Hidden() bool             { return f.hidden }
func (f *PathField) SetHidden(v bool)         { f.hidden = v }
func (f *PathField) Error() string            { return f.errMsg }
func (f *PathField) SetError(v string)        { f.errMsg = v }
func (f *PathField) Touched() bool            { return f.touched }
func (f *PathField) SetTouched(v bool)        { f.touched = v }
func (f *PathField) Value() any               { return strings.TrimSpace(f.text) }
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
func (f *PathField) Text() string     { return f.text }
func (f *PathField) SetText(s string) { f.text = s }
func (f *PathField) Cursor() int      { return f.cursor }
func (f *PathField) SetCursor(c int)  { f.cursor = c }
func (f *PathField) Toggle() bool     { return false }
func (f *PathField) SetToggle(_ bool) {}
func (f *PathField) Options() []string { return nil }
func (f *PathField) DisplayOptions() []string { return nil }
func (f *PathField) Index() int { return 0 }
func (f *PathField) SetIndex(_ int) {}
func (f *PathField) Selected() map[string]bool { return nil }
func (f *PathField) SetSelected(_ map[string]bool) {}

// DependsOn / DependsEq: PathField-specific. The state-level field-hidden
// walker in state.FormState.RecomputeVisibility uses type-assertion to
// access these.
func (f *PathField) DependsOn() string { return f.dependsOn }
func (f *PathField) DependsEq() bool   { return f.dependsEq }

func (f *PathField) Label() string    { return f.label }
func (f *PathField) HelperText() string { return f.helperText }
func (f *PathField) HelpText() string { return "path with Tab completion" }

func (f *PathField) Render(focused bool, width int, cursorVisible ...bool) string {
	value := renderTextCell(f.text, f.cursor, state.FormPath, focused, width, cursorVisible...)
	return appendUnitSuffix(value, "", width)
}

func (f *PathField) HandleKey(key string) (bool, bool) {
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

// PathField-specific methods (not on FormField).
func (f *PathField) PathSource() state.PathSource     { return f.pathSource }
func (f *PathField) PathMode() state.PathMode         { return f.pathMode }
func (f *PathField) ShowHidden() bool                 { return f.showHidden }
func (f *PathField) SetShowHidden(v bool)             { f.showHidden = v }
func (f *PathField) Suggestions() []state.PathEntry   { return f.suggestions }
func (f *PathField) SetSuggestions(v []state.PathEntry) { f.suggestions = v }
func (f *PathField) PathLoading() bool                { return f.loading }
func (f *PathField) SetPathLoading(v bool)            { f.loading = v }
func (f *PathField) PathTabInput() string             { return f.tabInput }
func (f *PathField) SetPathTabInput(v string)         { f.tabInput = v }
func (f *PathField) PathError() string                { return f.pathError }
func (f *PathField) SetPathError(v string)            { f.pathError = v }
func (f *PathField) TextRaw() string                 { return f.text }

// --- SelectField ----------------------------------------------------------

// SelectField is a one-of-many choice with a dropdown marker.
type SelectField struct {
	key, label, helperText string
	required               bool
	hidden                 bool
	errMsg                 string

	options        []string
	displayOptions []string
	index          int
	touched        bool
}

// SelectFieldConfig builds a SelectField.
type SelectFieldConfig struct {
	Key, Label, HelperText string
	Required               bool
	Options, DisplayOptions []string
	Index                   int
}

func NewSelectField(cfg SelectFieldConfig) state.FormField {
	return &SelectField{
		key:            cfg.Key,
		label:          cfg.Label,
		helperText:     cfg.HelperText,
		required:       cfg.Required,
		options:        cfg.Options,
		displayOptions: cfg.DisplayOptions,
		index:          cfg.Index,
	}
}

func (f *SelectField) Kind() state.FormFieldKind { return state.FormSelect }
func (f *SelectField) Key() string              { return f.key }
func (f *SelectField) Hidden() bool             { return f.hidden }
func (f *SelectField) SetHidden(v bool)         { f.hidden = v }
func (f *SelectField) Error() string            { return f.errMsg }
func (f *SelectField) SetError(v string)        { f.errMsg = v }
func (f *SelectField) Touched() bool            { return f.touched }
func (f *SelectField) SetTouched(v bool)        { f.touched = v }

func (f *SelectField) Value() any {
	if f.index < 0 || f.index >= len(f.options) {
		return ""
	}
	return f.options[f.index]
}
func (f *SelectField) Reset() { f.index = 0; f.touched = false; f.errMsg = "" }
func (f *SelectField) Text() string     { return "" }
func (f *SelectField) SetText(_ string) {}
func (f *SelectField) Cursor() int      { return 0 }
func (f *SelectField) SetCursor(_ int)  {}
func (f *SelectField) Toggle() bool     { return false }
func (f *SelectField) SetToggle(_ bool) {}
func (f *SelectField) Options() []string        { return f.options }
func (f *SelectField) DisplayOptions() []string { return f.displayOptions }
func (f *SelectField) Index() int               { return f.index }
func (f *SelectField) SetIndex(i int)           { f.index = i }
func (f *SelectField) Selected() map[string]bool { return nil }
func (f *SelectField) SetSelected(_ map[string]bool) {}

func (f *SelectField) Label() string   { return f.label }
func (f *SelectField) HelperText() string { return f.helperText }
func (f *SelectField) HelpText() string { return "single choice dropdown" }

func (f *SelectField) Render(focused bool, width int, _ ...bool) string {
	value, marker := renderSelectCellValue(f.options, f.displayOptions, f.index, nil, state.FormSelect, focused, width)
	return appendDropdownMarker(value, marker, width)
}

func (f *SelectField) HandleKey(string) (bool, bool) { return false, false }

func (f *SelectField) OpenPopup() (FormPopup, bool) {
	if len(f.options) == 0 {
		return nil, false
	}
	return &SelectPopup{options: f.options, multi: false}, true
}

// --- MultiSelectField -----------------------------------------------------

// MultiSelectField is a many-of-many choice where Space toggles each item.
type MultiSelectField struct {
	key, label, helperText string
	required               bool
	hidden                 bool
	errMsg                 string

	options        []string
	displayOptions []string
	selected       map[string]bool
	touched        bool
}

// MultiSelectFieldConfig builds a MultiSelectField.
type MultiSelectFieldConfig struct {
	Key, Label, HelperText string
	Required               bool
	Options, DisplayOptions []string
	Selected               map[string]bool
}

func NewMultiSelectField(cfg MultiSelectFieldConfig) state.FormField {
	return &MultiSelectField{
		key:            cfg.Key,
		label:          cfg.Label,
		helperText:     cfg.HelperText,
		required:       cfg.Required,
		options:        cfg.Options,
		displayOptions: cfg.DisplayOptions,
		selected:       cfg.Selected,
	}
}

func (f *MultiSelectField) Kind() state.FormFieldKind { return state.FormMultiSelect }
func (f *MultiSelectField) Key() string              { return f.key }
func (f *MultiSelectField) Hidden() bool             { return f.hidden }
func (f *MultiSelectField) SetHidden(v bool)         { f.hidden = v }
func (f *MultiSelectField) Error() string            { return f.errMsg }
func (f *MultiSelectField) SetError(v string)        { f.errMsg = v }
func (f *MultiSelectField) Touched() bool            { return f.touched }
func (f *MultiSelectField) SetTouched(v bool)        { f.touched = v }

func (f *MultiSelectField) Value() any {
	if f.selected == nil {
		return map[string]bool{}
	}
	return f.selected
}
func (f *MultiSelectField) Reset() { f.selected = nil; f.touched = false; f.errMsg = "" }
func (f *MultiSelectField) Text() string     { return "" }
func (f *MultiSelectField) SetText(_ string) {}
func (f *MultiSelectField) Cursor() int      { return 0 }
func (f *MultiSelectField) SetCursor(_ int)  {}
func (f *MultiSelectField) Toggle() bool     { return false }
func (f *MultiSelectField) SetToggle(_ bool) {}
func (f *MultiSelectField) Options() []string        { return f.options }
func (f *MultiSelectField) DisplayOptions() []string { return f.displayOptions }
func (f *MultiSelectField) Index() int               { return 0 }
func (f *MultiSelectField) SetIndex(_ int)           {}
func (f *MultiSelectField) Selected() map[string]bool { return f.selected }
func (f *MultiSelectField) SetSelected(v map[string]bool) { f.selected = v }

func (f *MultiSelectField) Label() string   { return f.label }
func (f *MultiSelectField) HelperText() string { return f.helperText }
func (f *MultiSelectField) HelpText() string { return "multi-select dropdown" }

func (f *MultiSelectField) Render(focused bool, width int, _ ...bool) string {
	value, marker := renderSelectCellValue(f.options, f.displayOptions, 0, f.selected, state.FormMultiSelect, focused, width)
	return appendDropdownMarker(value, marker, width)
}

func (f *MultiSelectField) HandleKey(string) (bool, bool) { return false, false }

func (f *MultiSelectField) OpenPopup() (FormPopup, bool) {
	if len(f.options) == 0 {
		return nil, false
	}
	return &SelectPopup{options: f.options, selected: f.selected, multi: true}, true
}

// --- BoolField ------------------------------------------------------------

// BoolField is a checkbox toggled with Space. Recomputes the dependency
// graph so child fields (DependsOn) show/hide correctly.
type BoolField struct {
	key, label, helperText string
	required               bool
	hidden                 bool
	errMsg                 string
	dependsOn              string
	dependsEq              bool

	toggle  bool
	touched bool
}

// BoolFieldConfig builds a BoolField.
type BoolFieldConfig struct {
	Key, Label, HelperText string
	Required               bool
	Toggle                 bool
	DependsOn              string
	DependsEq              bool
}

func NewBoolField(cfg BoolFieldConfig) state.FormField {
	return &BoolField{
		key:        cfg.Key,
		label:      cfg.Label,
		helperText: cfg.HelperText,
		required:   cfg.Required,
		toggle:     cfg.Toggle,
		dependsOn:  cfg.DependsOn,
		dependsEq:  cfg.DependsEq,
	}
}

func (f *BoolField) Kind() state.FormFieldKind { return state.FormBool }
func (f *BoolField) Key() string              { return f.key }
func (f *BoolField) Hidden() bool             { return f.hidden }
func (f *BoolField) SetHidden(v bool)         { f.hidden = v }
func (f *BoolField) Error() string            { return f.errMsg }
func (f *BoolField) SetError(v string)        { f.errMsg = v }
func (f *BoolField) Touched() bool            { return f.touched }
func (f *BoolField) SetTouched(v bool)        { f.touched = v }
func (f *BoolField) Value() any               { return f.toggle }
func (f *BoolField) Reset()                   { f.toggle = false; f.touched = false; f.errMsg = "" }
func (f *BoolField) Text() string             { return "" }
func (f *BoolField) SetText(_ string)         {}
func (f *BoolField) Cursor() int              { return 0 }
func (f *BoolField) SetCursor(_ int)          {}
func (f *BoolField) Toggle() bool             { return f.toggle }
func (f *BoolField) SetToggle(v bool)         { f.toggle = v }
func (f *BoolField) Options() []string        { return nil }
func (f *BoolField) DisplayOptions() []string { return nil }
func (f *BoolField) Index() int               { return 0 }
func (f *BoolField) SetIndex(_ int)           {}
func (f *BoolField) Selected() map[string]bool { return nil }
func (f *BoolField) SetSelected(_ map[string]bool) {}

// DependsOn / DependsEq: BoolField-specific.
func (f *BoolField) DependsOn() string { return f.dependsOn }
func (f *BoolField) DependsEq() bool   { return f.dependsEq }

func (f *BoolField) Label() string   { return f.label }
func (f *BoolField) HelperText() string { return f.helperText }
func (f *BoolField) HelpText() string { return "boolean checkbox" }

func (f *BoolField) Render(focused bool, width int, _ ...bool) string {
	return fitValue(renderBoolCell(f.toggle, focused), width)
}

func (f *BoolField) HandleKey(key string) (bool, bool) {
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

// --- RadioGroupField ------------------------------------------------------

// RadioGroupField is an inline single-choice alternative to SelectField.
type RadioGroupField struct {
	key, label, helperText string
	required               bool
	hidden                 bool
	errMsg                 string

	options        []string
	displayOptions []string
	index          int
	touched        bool
}

// RadioGroupFieldConfig builds a RadioGroupField.
type RadioGroupFieldConfig struct {
	Key, Label, HelperText string
	Required               bool
	Options, DisplayOptions []string
	Index                   int
}

func NewRadioGroupField(cfg RadioGroupFieldConfig) state.FormField {
	return &RadioGroupField{
		key:            cfg.Key,
		label:          cfg.Label,
		helperText:     cfg.HelperText,
		required:       cfg.Required,
		options:        cfg.Options,
		displayOptions: cfg.DisplayOptions,
		index:          cfg.Index,
	}
}

func (f *RadioGroupField) Kind() state.FormFieldKind { return state.FormRadioGroup }
func (f *RadioGroupField) Key() string              { return f.key }
func (f *RadioGroupField) Hidden() bool             { return f.hidden }
func (f *RadioGroupField) SetHidden(v bool)         { f.hidden = v }
func (f *RadioGroupField) Error() string            { return f.errMsg }
func (f *RadioGroupField) SetError(v string)        { f.errMsg = v }
func (f *RadioGroupField) Touched() bool            { return f.touched }
func (f *RadioGroupField) SetTouched(v bool)        { f.touched = v }

func (f *RadioGroupField) Value() any {
	if f.index < 0 || f.index >= len(f.options) {
		return ""
	}
	return f.options[f.index]
}
func (f *RadioGroupField) Reset() { f.index = 0; f.touched = false; f.errMsg = "" }
func (f *RadioGroupField) Text() string     { return "" }
func (f *RadioGroupField) SetText(_ string) {}
func (f *RadioGroupField) Cursor() int      { return 0 }
func (f *RadioGroupField) SetCursor(_ int)  {}
func (f *RadioGroupField) Toggle() bool     { return false }
func (f *RadioGroupField) SetToggle(_ bool) {}
func (f *RadioGroupField) Options() []string        { return f.options }
func (f *RadioGroupField) DisplayOptions() []string { return f.displayOptions }
func (f *RadioGroupField) Index() int               { return f.index }
func (f *RadioGroupField) SetIndex(i int)           { f.index = i }
func (f *RadioGroupField) Selected() map[string]bool { return nil }
func (f *RadioGroupField) SetSelected(_ map[string]bool) {}

func (f *RadioGroupField) Label() string   { return f.label }
func (f *RadioGroupField) HelperText() string { return f.helperText }
func (f *RadioGroupField) HelpText() string { return "radio group (single choice)" }

func (f *RadioGroupField) Render(focused bool, width int, _ ...bool) string {
	label := fieldValueLabel(f.options, f.displayOptions, f.index)
	value := utils.TruncateVisible(label, max(1, width-4))
	style := component.GetStyle(component.StyleDim).GetForeground()
	if focused {
		style = component.GetStyle(component.StylePanelTitle).GetForeground()
	}
	styled := lipgloss.NewStyle().Foreground(style).Bold(focused).Render(value)
	return fitValue(styled, max(1, width-4)) + " (" + component.MarkCheck + ")"
}

func (f *RadioGroupField) HandleKey(string) (bool, bool) { return false, false }

func (f *RadioGroupField) OpenPopup() (FormPopup, bool) { return nil, false }

// --- TextMultiLineField ---------------------------------------------------

// TextMultiLineField is a long-text input. Renders the first line only.
type TextMultiLineField struct {
	key, label, helperText string
	required               bool
	placeholder            string
	hidden                 bool
	errMsg                 string

	text   string
	cursor int
	touched bool
}

// TextMultiLineFieldConfig builds a TextMultiLineField.
type TextMultiLineFieldConfig struct {
	Key, Label, HelperText string
	Required               bool
	Placeholder            string
	Text                   string
	Cursor                 int
}

func NewTextMultiLineField(cfg TextMultiLineFieldConfig) state.FormField {
	return &TextMultiLineField{
		key:         cfg.Key,
		label:       cfg.Label,
		helperText:  cfg.HelperText,
		required:    cfg.Required,
		placeholder: cfg.Placeholder,
		text:        cfg.Text,
		cursor:      cfg.Cursor,
	}
}

func (f *TextMultiLineField) Kind() state.FormFieldKind { return state.FormTextMultiLine }
func (f *TextMultiLineField) Key() string              { return f.key }
func (f *TextMultiLineField) Hidden() bool             { return f.hidden }
func (f *TextMultiLineField) SetHidden(v bool)         { f.hidden = v }
func (f *TextMultiLineField) Error() string            { return f.errMsg }
func (f *TextMultiLineField) SetError(v string)        { f.errMsg = v }
func (f *TextMultiLineField) Touched() bool            { return f.touched }
func (f *TextMultiLineField) SetTouched(v bool)        { f.touched = v }
func (f *TextMultiLineField) Value() any               { return strings.TrimSpace(f.text) }
func (f *TextMultiLineField) Reset()                   { f.text = ""; f.cursor = 0; f.touched = false; f.errMsg = "" }
func (f *TextMultiLineField) Text() string             { return f.text }
func (f *TextMultiLineField) SetText(s string)         { f.text = s }
func (f *TextMultiLineField) Cursor() int              { return f.cursor }
func (f *TextMultiLineField) SetCursor(c int)          { f.cursor = c }
func (f *TextMultiLineField) Toggle() bool             { return false }
func (f *TextMultiLineField) SetToggle(_ bool)         {}
func (f *TextMultiLineField) Options() []string        { return nil }
func (f *TextMultiLineField) DisplayOptions() []string { return nil }
func (f *TextMultiLineField) Index() int               { return 0 }
func (f *TextMultiLineField) SetIndex(_ int)           {}
func (f *TextMultiLineField) Selected() map[string]bool { return nil }
func (f *TextMultiLineField) SetSelected(_ map[string]bool) {}

func (f *TextMultiLineField) Label() string   { return f.label }
func (f *TextMultiLineField) HelperText() string { return f.helperText }
func (f *TextMultiLineField) HelpText() string { return "multi-line text (first line shown)" }

func (f *TextMultiLineField) Render(focused bool, width int, cursorVisible ...bool) string {
	first, _, _ := strings.Cut(f.text, "\n")
	cursor := f.cursor
	if r := len([]rune(first)); cursor > r {
		cursor = r
	}
	value := renderTextCell(first, cursor, state.FormTextMultiLine, focused, width, cursorVisible...)
	return appendUnitSuffix(value, "", width)
}

func (f *TextMultiLineField) HandleKey(key string) (bool, bool) {
	h, c := applyEditKey(key, &f.text, &f.cursor)
	if c {
		f.touched = true
	}
	return h, c
}

func (f *TextMultiLineField) OpenPopup() (FormPopup, bool) { return nil, false }

// --- TextPasswordField ----------------------------------------------------

// TextPasswordField masks the value with U+2022 BULLET.
type TextPasswordField struct {
	key, label, helperText string
	required               bool
	placeholder            string
	hidden                 bool
	errMsg                 string

	text   string
	cursor int
	touched bool
}

// TextPasswordFieldConfig builds a TextPasswordField.
type TextPasswordFieldConfig struct {
	Key, Label, HelperText string
	Required               bool
	Placeholder            string
	Text                   string
	Cursor                 int
}

func NewTextPasswordField(cfg TextPasswordFieldConfig) state.FormField {
	return &TextPasswordField{
		key:         cfg.Key,
		label:       cfg.Label,
		helperText:  cfg.HelperText,
		required:    cfg.Required,
		placeholder: cfg.Placeholder,
		text:        cfg.Text,
		cursor:      cfg.Cursor,
	}
}

func (f *TextPasswordField) Kind() state.FormFieldKind { return state.FormTextPassword }
func (f *TextPasswordField) Key() string              { return f.key }
func (f *TextPasswordField) Hidden() bool             { return f.hidden }
func (f *TextPasswordField) SetHidden(v bool)         { f.hidden = v }
func (f *TextPasswordField) Error() string            { return f.errMsg }
func (f *TextPasswordField) SetError(v string)        { f.errMsg = v }
func (f *TextPasswordField) Touched() bool            { return f.touched }
func (f *TextPasswordField) SetTouched(v bool)        { f.touched = v }
func (f *TextPasswordField) Value() any               { return strings.TrimSpace(f.text) }
func (f *TextPasswordField) Reset()                   { f.text = ""; f.cursor = 0; f.touched = false; f.errMsg = "" }
func (f *TextPasswordField) Text() string             { return f.text }
func (f *TextPasswordField) SetText(s string)         { f.text = s }
func (f *TextPasswordField) Cursor() int              { return f.cursor }
func (f *TextPasswordField) SetCursor(c int)          { f.cursor = c }
func (f *TextPasswordField) Toggle() bool             { return false }
func (f *TextPasswordField) SetToggle(_ bool)         {}
func (f *TextPasswordField) Options() []string        { return nil }
func (f *TextPasswordField) DisplayOptions() []string { return nil }
func (f *TextPasswordField) Index() int               { return 0 }
func (f *TextPasswordField) SetIndex(_ int)           {}
func (f *TextPasswordField) Selected() map[string]bool { return nil }
func (f *TextPasswordField) SetSelected(_ map[string]bool) {}

func (f *TextPasswordField) Label() string   { return f.label }
func (f *TextPasswordField) HelperText() string { return f.helperText }
func (f *TextPasswordField) HelpText() string { return "password (masked)" }

func (f *TextPasswordField) Render(focused bool, width int, cursorVisible ...bool) string {
	masked := strings.Repeat(passwordBullet, len([]rune(f.text)))
	value := renderTextCell(masked, f.cursor, state.FormTextPassword, focused, width, cursorVisible...)
	return appendUnitSuffix(value, "", width)
}

func (f *TextPasswordField) HandleKey(key string) (bool, bool) {
	h, c := applyEditKey(key, &f.text, &f.cursor)
	if c {
		f.touched = true
	}
	return h, c
}

func (f *TextPasswordField) OpenPopup() (FormPopup, bool) { return nil, false }

// --- Popup impls ----------------------------------------------------------

// SelectPopup adapts a Select or MultiSelect field's options into a popup
// renderer. The hot path uses FormPopupState + handleFormPopupKey.
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
func (p *SelectPopup) Render(width, height int) string { return "" }

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

// PathPopup adapts a PathField's suggestions into a popup renderer.
type PathPopup struct {
	owner  *PathField
	cursor int
}

func (p *PathPopup) Title() string     { return "" }
func (p *PathPopup) Highlighted() int { return p.cursor }
func (p *PathPopup) Size() (int, int) { return 100, state.FormPopupVisibleRows + 5 }
func (p *PathPopup) Render(width, height int) string { return "" }

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

// --- Render helpers (primitive-input) --------------------------------------

func renderTextCell(text string, cursor int, kind state.FormFieldKind, focused bool, valueWidth int, cursorVisible ...bool) string {
	runes := []rune(text)
	cursor = clampCursor(cursor, len(runes))
	start := 0
	cursorWidth := 1
	if cursor < len(runes) {
		cursorWidth = max(1, utils.DisplayWidth(string(runes[cursor])))
	}
	for start < cursor && utils.DisplayWidth(string(runes[start:cursor]))+cursorWidth > valueWidth {
		start++
	}
	end := cursor
	if cursor < len(runes) {
		end++
	}
	for end < len(runes) && utils.DisplayWidth(string(runes[start:end+1])) <= valueWidth {
		end++
	}
	base := lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleDim).GetForeground())
	if kind == state.FormPath && !focused {
		return wrapVisiblePath(valueWidth, text)
	}
	if !focused {
		visible := base.Render(text)
		return utils.TruncateVisible(visible, valueWidth)
	}

	before := lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleHelpDescription).GetForeground()).Render(string(runes[start:cursor]))
	visible := len(cursorVisible) == 0 || cursorVisible[0]
	caret := lipgloss.NewStyle().Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).Bold(true).Underline(true)
	current := component.NarrowCursor
	if cursor < len(runes) {
		current = string(runes[cursor])
	}
	if !visible {
		caret = lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleHelpDescription).GetForeground())
		if cursor == len(runes) {
			current = " "
		}
	}
	afterStart := cursor
	if cursor < len(runes) {
		afterStart++
	}
	after := lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleHelpDescription).GetForeground()).Render(string(runes[afterStart:end]))
	truncated := utils.TruncateVisible(before+caret.Render(current)+after, valueWidth)
	return component.GetStyle(component.StyleFormInput).Render(truncated)
}

func renderSelectCellValue(options []string, displayOptions []string, index int, selected map[string]bool, kind state.FormFieldKind, focused bool, valueWidth int) (value, marker string) {
	marker = component.TriangleDownSmall
	displayed := func(idx int) string {
		if idx >= 0 && idx < len(options) && idx < len(displayOptions) && displayOptions[idx] != "" {
			return displayOptions[idx]
		}
		if idx >= 0 && idx < len(options) {
			return options[idx]
		}
		return ""
	}
	var text string
	switch kind {
	case state.FormMultiSelect:
		if len(selected) > 0 {
			keys := make([]string, 0, len(selected))
			for k := range selected {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			keyIndex := make(map[string]int, len(options))
			for i, opt := range options {
				keyIndex[opt] = i
			}
			labels := make([]string, 0, len(keys))
			for _, k := range keys {
				if idx, ok := keyIndex[k]; ok {
					labels = append(labels, displayed(idx))
				} else {
					labels = append(labels, k)
				}
			}
			text = strings.Join(labels, ", ")
		}
	default:
		text = displayed(index)
	}
	value = utils.TruncateVisible(text, valueWidth)
	if focused {
		value = lipgloss.NewStyle().Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).Bold(true).Render(value)
	} else {
		value = lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleDim).GetForeground()).Render(value)
	}
	return value, marker
}

func renderBoolCell(toggle, focused bool) string {
	mark := " "
	if toggle {
		mark = component.MarkCheck
	}
	cell := "[" + mark + "]"
	if focused {
		return lipgloss.NewStyle().
			Foreground(component.GetStyle(component.StyleDialogConfirm).GetForeground()).
			Bold(true).
			Render(cell)
	}
	return lipgloss.NewStyle().
		Foreground(component.GetStyle(component.StyleDim).GetForeground()).
		Render(cell)
}

func wrapVisiblePath(valueWidth int, value string) string {
	if valueWidth <= 0 || value == "" {
		return value
	}
	var lines []string
	line := ""
	lineWidth := 0
	for _, r := range []rune(value) {
		width := utils.DisplayWidth(string(r))
		if line != "" && lineWidth+width > valueWidth {
			lines = append(lines, line)
			line = ""
			lineWidth = 0
		}
		line += string(r)
		lineWidth += width
	}
	if line != "" || len(lines) == 0 {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func clampCursor(cursor, length int) int {
	if cursor < 0 {
		return 0
	}
	if cursor > length {
		return length
	}
	return cursor
}

func fieldValueLabel(options []string, displayOptions []string, index int) string {
	if index >= 0 && index < len(options) {
		if index < len(displayOptions) && displayOptions[index] != "" {
			return displayOptions[index]
		}
		return options[index]
	}
	return ""
}