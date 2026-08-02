package dialog

import (
	"sort"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
	"github.com/elizabevil/docker-tui/internal/utils"

	"charm.land/lipgloss/v2"
)

// FormDialog renders the container-action form as a two-column layout: a
// right-aligned label column and a shared value column (BR-041 §9). The
// Confirm / Cancel pair stays in the same Tab loop as the fields.
func FormDialog(m *state.AppModel, overlayColor string, cfg dialogConfig) string {
	form := m.Form
	dialogW := dialogWidth(m.Viewport.Width, cfg)
	dialogH := dialogHeight(m.Viewport.Height, cfg)
	if overlayColor == "" {
		overlayColor = "#0d1117cc"
	}
	innerWidth := formInnerWidth(dialogW)

	enterKey := i18n.T("key.sym_enter")
	escKey := i18n.T("key.sym_esc")
	confirmLabel := i18n.T("key.confirm")
	cancelLabel := i18n.T("key.cancel")

	parts := []string{
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Left).Render(component.GetStyle("panelTitle").Render(form.Title)),
		"",
	}
	labelWidth, valueWidth := formLayout(form.Fields, innerWidth)
	for i := range form.Fields {
		parts = append(parts, renderFormField(form, &form.Fields[i], i, labelWidth, valueWidth))
	}
	parts = append(parts, "")

	onConfirm := form.OnConfirm && form.FieldFocus < 0
	var confirmBtn, cancelBtn string
	if onConfirm {
		confirmBtn = lipgloss.NewStyle().Foreground(style.Colors.Green).Bold(true).Render(enterKey + " \u25b6 " + confirmLabel)
		cancelBtn = lipgloss.NewStyle().Foreground(style.Colors.Gray).Render(escKey + " " + cancelLabel)
	} else {
		confirmBtn = lipgloss.NewStyle().Foreground(style.Colors.Gray).Render(enterKey + " " + confirmLabel)
		cancelBtn = lipgloss.NewStyle().Foreground(style.Colors.Green).Bold(true).Render(escKey + " \u25b6 " + cancelLabel)
	}
	// Cancel on the left, Confirm on the right; the group is centered.
	buttons := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
		cancelBtn + "   " + confirmBtn,
	)

	parts = append(parts, buttons)
	parts = append(parts, "")
	parts = append(parts, component.GetStyle("dim").Render(i18n.T("form.hint.navigation")))

	box := DialogBox(DialogStyle{Width: dialogW, Height: dialogH, TitleColor: style.Colors.Cyan, OverlayColor: overlayColor, LeftAligned: true}, parts...)
	if form.Popup.Open {
		box = renderFormPopup(form, box, dialogW, dialogH)
	}
	return box
}

// formInnerWidth is the width available to a row after border and padding are
// removed from the dialog.
func formInnerWidth(dialogW int) int {
	if dialogW <= 4 {
		return 1
	}
	return dialogW - 4
}

// formLayout computes the label and value column widths for the given fields
// from the dialog's inner width, outside-in (BR-041 §9).
func formLayout(fields []state.FormField, innerWidth int) (labelWidth, valueWidth int) {
	labels := make([]string, 0, len(fields))
	for i := range fields {
		labels = append(labels, fields[i].Label)
	}
	labelWidth = component.FormLabelColumnWidth(labels, innerWidth)
	valueWidth = innerWidth - labelWidth - 1
	if valueWidth < 1 {
		valueWidth = 1
	}
	return labelWidth, valueWidth
}

// renderFormField draws one two-column row for a form field. The focused field
// is highlighted; any validation error is drawn on a following line aligned to
// the value column (BR-041 §9.8).
func renderFormField(form state.FormState, f *state.FormField, index, labelWidth, valueWidth int) string {
	focused := form.FieldFocus == index

	label := component.FormRow(f.Label, labelWidth, -1, "")
	if focused {
		label = lipgloss.NewStyle().Foreground(style.Colors.Cyan).Bold(true).Render(label)
	}

	var value, marker string
	if f.Kind == state.FormSelect || f.Kind == state.FormMultiSelect {
		value, marker = renderSelectCell(f, focused, valueWidth)
	} else {
		value = renderFormValue(f, focused, valueWidth)
	}

	if marker != "" {
		value = utils.FitVisible(value, max(1, valueWidth-2)) + " " + marker
	} else {
		value = utils.FitVisible(value, valueWidth)
	}
	row := label + " " + value
	if f.Error != "" {
		pad := utils.PadVisible("", labelWidth+1)
		row += "\n" + pad + lipgloss.NewStyle().Foreground(style.Colors.Red).Render(f.Error)
	}
	return row
}

// renderFormValue renders the primitive value of a text/int/path/bool field.
func renderFormValue(f *state.FormField, focused bool, valueWidth int) string {
	switch f.Kind {
	case state.FormBool:
		mark := " "
		if f.Toggle {
			mark = "\u2713"
		}
		cell := "[" + mark + "]"
		if focused {
			return lipgloss.NewStyle().Foreground(style.Colors.Green).Bold(true).Render(cell)
		}
		return lipgloss.NewStyle().Foreground(style.Colors.Gray).Render(cell)
	default: // FormText, FormInt, FormPath
		return renderEditableValue(f, focused, valueWidth)
	}
}

// renderEditableValue draws a text-like field. Inside existing text the
// cursor styles the current rune without consuming another terminal cell, so
// Left/Right never shifts the path. At end-of-input a one-cell caret is shown.
func renderEditableValue(f *state.FormField, focused bool, valueWidth int) string {
	runes := []rune(f.Input.Text)
	cursor := clampCursor(f.Input.Cursor, len(runes))
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
	base := lipgloss.NewStyle().Foreground(style.Colors.Gray)
	if !focused {
		return utils.TruncateVisible(base.Render(string(runes[start:end])), valueWidth)
	}

	before := lipgloss.NewStyle().Foreground(style.Colors.White).Render(string(runes[start:cursor]))
	caret := lipgloss.NewStyle().Foreground(style.Colors.Cyan).Bold(true).Underline(true)
	current := "\u258f"
	if cursor < len(runes) {
		current = string(runes[cursor])
	}
	afterStart := cursor
	if cursor < len(runes) {
		afterStart++
	}
	after := lipgloss.NewStyle().Foreground(style.Colors.White).Render(string(runes[afterStart:end]))
	return utils.TruncateVisible(before+caret.Render(current)+after, valueWidth)
}

// renderSelectCell renders the collapsed value plus a dropdown marker for a
// single- or multi-select field (BR-041 §8.1, §8.2).
func renderSelectCell(f *state.FormField, focused bool, valueWidth int) (value, marker string) {
	marker = "\u25be"
	var text string
	switch f.Kind {
	case state.FormMultiSelect:
		if len(f.Selected) > 0 {
			keys := make([]string, 0, len(f.Selected))
			for k := range f.Selected {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			text = strings.Join(keys, ", ")
		}
	default:
		if f.Index >= 0 && f.Index < len(f.Options) {
			text = f.Options[f.Index]
		}
	}
	value = utils.TruncateVisible(text, valueWidth)
	if focused {
		value = lipgloss.NewStyle().Foreground(style.Colors.Cyan).Bold(true).Render(value)
	} else {
		value = lipgloss.NewStyle().Foreground(style.Colors.Gray).Render(value)
	}
	return value, marker
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
