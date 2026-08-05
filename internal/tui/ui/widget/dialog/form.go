package dialog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"

	"charm.land/lipgloss/v2"
)

// FormDialog renders the container-action form as a two-column layout: a
// right-aligned label column and a shared value column (BR-041 §9). The
// Confirm / Cancel pair stays in the same Tab loop as the fields.
//
// When bodyW > 0 and bodyH > 0, dialog sizing follows the active panel body
// (BR-043 §3.2: width = bodyW*3/4, height = bodyH, clamped to cfg). Otherwise
// it falls back to viewport-percentage sizing (dialogWidth/dialogHeight).
func FormDialog(m *state.AppModel, overlayColor string, cfg DialogConfig, bodyW, bodyH int) string {
	form := m.Form
	dialogW, dialogH := formDialogSize(m, cfg, bodyW, bodyH)
	if overlayColor == "" {
		overlayColor = OverlayColor(nil)
	}
	innerWidth := formInnerWidth(dialogW)

	enterKey := i18n.T("key.sym_enter")
	escKey := i18n.T("key.sym_esc")

	parts := []string{
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Left).
			Render(component.GetStyle(component.StylePanelTitle).Render(form.Title)),
	}
	parts = appendFormHeader(parts, form)
	parts = append(parts, "")
	labelWidth, valueWidth := formLayout(form.Fields, innerWidth)
	parts = appendFormLoading(parts, form)
	parts = appendFormFields(parts, form, labelWidth, valueWidth, !m.CursorBlinkHidden)
	parts = append(parts, "")

	// Confirm/Cancel on the same line, left-right (BR-043 §3.3 scheme B +
	// height 3/4 revision). Cancel on the left (Esc), Confirm on the right
	// (Enter). Focus stays linear: Cancel = n+1, Confirm = n.
	cancelBtn := renderFormButton(escKey, form.CancelLabel, form.FieldFocus == form.CancelSlot() && !form.Popup.Open)
	confirmBtn := renderFormButton(enterKey, form.ConfirmLabel, form.FieldFocus == form.ConfirmSlot() && !form.Popup.Open)
	buttons := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
		cancelBtn + "   " + confirmBtn,
	)
	hintKey := "form.hint.navigation"
	if form.Kind == state.FormContainerUpdate {
		hintKey = "form.hint.update_navigation"
	}
	parts = append(parts, buttons, "", component.GetStyle(component.StyleDim).Render(i18n.T(hintKey)))

	box := DialogBox(DialogStyle{
		Width:        dialogW,
		Height:       dialogH,
		TitleColor:   component.GetStyle(component.StylePanelTitle).GetForeground(),
		OverlayColor: overlayColor,
		LeftAligned:  true,
	}, parts...)
	if form.Popup.Open {
		box = renderFormPopup(form, box, dialogW, dialogH)
	}
	return box
}

func formDialogSize(m *state.AppModel, cfg DialogConfig, bodyW, bodyH int) (dialogW, dialogH int) {
	if bodyW > 0 && bodyH > 0 {
		dialogW = panelDialogWidth(bodyW, cfg)
		dialogH = panelDialogHeight(bodyH, cfg)
	} else {
		dialogW = dialogWidth(m.Viewport.Width, cfg)
		dialogH = dialogHeight(m.Viewport.Height, cfg)
	}
	requiredH := 10
	for i := range m.Form.Fields {
		if m.Form.Fields[i].Hidden {
			continue
		}
		requiredH++
		if m.Form.Fields[i].Error != "" {
			requiredH++
		}
	}
	if m.Form.Loading {
		requiredH++
	}
	if requiredH > dialogH {
		dialogH = requiredH
	}
	if maxH := m.Viewport.Height - 2; maxH > 0 && dialogH > maxH {
		dialogH = maxH
	}
	return dialogW, dialogH
}

func appendFormHeader(parts []string, form state.FormState) []string {
	if form.TargetName == "" {
		return parts
	}
	target := form.TargetName
	if form.TargetID != "" && form.TargetID != target {
		target = fmt.Sprintf("%s (%s)", target, form.TargetID)
	}
	line := fmt.Sprintf("%s: %s", targetLabelForForm(form.Kind), target)
	return append(parts, component.GetStyle(component.StyleDim).Render(line))
}

func appendFormLoading(parts []string, form state.FormState) []string {
	if !form.Loading {
		return parts
	}
	return append(parts, component.GetStyle(component.StyleDim).Render(i18n.T("container.update.form.loading")))
}

func appendFormFields(parts []string, form state.FormState, labelWidth, valueWidth int, cursorVisible bool) []string {
	for i := range form.Fields {
		if form.Fields[i].Hidden {
			continue
		}
		parts = append(parts, renderFormField(form, &form.Fields[i], i, labelWidth, valueWidth, cursorVisible))
	}
	return parts
}

func renderFormButton(key, label string, focused bool) string {
	if focused {
		return lipgloss.NewStyle().
			Foreground(component.GetStyle(component.StyleDialogConfirm).GetForeground()).
			Bold(true).
			Render(key + " " + component.ButtonIndicator + " " + label)
	}
	return lipgloss.NewStyle().
		Foreground(component.GetStyle(component.StyleDim).GetForeground()).
		Render(key + " " + label)
}

// formInnerWidth is the width available to a row after border and padding are
// removed from the dialog.
func formInnerWidth(dialogW int) int {
	if dialogW <= 4 {
		return 1
	}
	return dialogW - 4
}

// targetLabelForForm returns the human-readable target label for the form kind
// (BR-043 §8.3: dialog header must show which container/image the operation
// targets).
func targetLabelForForm(kind state.FormKind) string {
	switch kind {
	case state.FormContainerCopy, state.FormContainerUpdate,
		state.FormContainerExport, state.FormContainerCommit:
		return "Container"
	case state.FormImageSave, state.FormImageLoad:
		return "Image"
	default:
		return ""
	}
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
func renderFormField(form state.FormState, f *state.FormField, index, labelWidth, valueWidth int, cursorVisible ...bool) string {
	focused := form.FieldFocus == index

	label := component.FormRow(f.Label, labelWidth, -1, "")
	if focused {
		label = lipgloss.NewStyle().Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).Bold(true).Render(label)
	}

	var value, marker string
	if f.Kind == state.FormSelect || f.Kind == state.FormMultiSelect {
		value, marker = renderSelectCell(f, focused, valueWidth)
	} else {
		value = renderFormValue(f, focused, valueWidth, cursorVisible...)
	}

	if marker != "" {
		value = utils.FitVisible(value, max(1, valueWidth-2)) + " " + marker
	} else {
		value = utils.FitVisible(value, valueWidth)
	}
	row := label + " " + value
	if f.Error != "" {
		pad := utils.PadVisible("", labelWidth+1)
		row += "\n" + pad + lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleDialogError).GetForeground()).Render(f.Error)
	}
	return row
}

// renderFormValue renders the primitive value of a text/int/path/bool field.
func renderFormValue(f *state.FormField, focused bool, valueWidth int, cursorVisible ...bool) string {
	switch f.Kind {
	case state.FormBool:
		mark := " "
		if f.Toggle {
			mark = component.MarkCheck
		}
		cell := "[" + mark + "]"
		if focused {
			return lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleDialogConfirm).GetForeground()).Bold(true).Render(cell)
		}
		return lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleDim).GetForeground()).Render(cell)
	default: // FormText, FormInt, FormPath
		return renderEditableValue(f, focused, valueWidth, cursorVisible...)
	}
}

// renderEditableValue draws a text-like field. Inside existing text the
// cursor styles the current rune without consuming another terminal cell, so
// Left/Right never shifts the path. At end-of-input a one-cell caret is shown.
func renderEditableValue(f *state.FormField, focused bool, valueWidth int, cursorVisible ...bool) string {
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
	base := lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleDim).GetForeground())
	if !focused {
		return utils.TruncateVisible(base.Render(string(runes[start:end])), valueWidth)
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

// renderSelectCell renders the collapsed value plus a dropdown marker for a
// single- or multi-select field (BR-041 §8.1, §8.2).
func renderSelectCell(f *state.FormField, focused bool, valueWidth int) (value, marker string) {
	marker = component.TriangleDownSmall
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
		value = lipgloss.NewStyle().Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).Bold(true).Render(value)
	} else {
		value = lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleDim).GetForeground()).Render(value)
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
