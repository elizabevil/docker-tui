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
	// popupInnerW is the width to use when rendering inline dropdown rows
	// beneath a focused select field. Width = label column + value column + 1
	// (separator) so the dropdown lines line up with the field column.
	popupInnerW := labelWidth + valueWidth + 1
	parts = appendFormLoading(parts, form)
	parts = appendFormFields(parts, form, labelWidth, valueWidth, popupInnerW, !m.CursorBlinkHidden)
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
		MaxWidth:     cfg.PanelSize.MaxWidth,
		TitleColor:   component.GetStyle(component.StylePanelTitle).GetForeground(),
		OverlayColor: overlayColor,
		LeftAligned:  true,
	}, parts...)
	if form.Popup.Open && form.Popup.Kind == state.PopupPath {
		// Path popup is a multi-line directory/file picker; show as a
		// separate dialog because the inline form layout cannot host the
		// breadcrumb + column header + detail box at once.
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
	// Compute the label column width so we can count how many rows each
	// HelperText wrap will add (design P: HelperText wraps, never truncates).
	var labelW int
	if innerW := formInnerWidth(dialogW); innerW > 0 && len(m.Form.Fields) > 0 {
		labelW, _ = formLayout(m.Form.Fields, innerW)
	}
	requiredH := 10
	for i := range m.Form.Fields {
		if m.Form.Fields[i].Hidden {
			continue
		}
		requiredH++
		// HelperText wrap continuation lines: the description is wrapped to
		// the label column width (rune/display-width aware via
		// utils.WrapCells); extra rows live beneath the value column.
		if labelW > 0 && m.Form.Fields[i].HelperText != "" {
			full := m.Form.Fields[i].Label + " (" + m.Form.Fields[i].HelperText + ")"
			if extra := len(utils.WrapCells(full, labelW)) - 1; extra > 0 {
				requiredH += extra
			}
		}
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

func appendFormFields(parts []string, form state.FormState, labelWidth, valueWidth, popupInnerW int, cursorVisible bool) []string {
	for i := range form.Fields {
		if form.Fields[i].Hidden {
			continue
		}
		parts = append(parts, renderFormField(form, &form.Fields[i], i, labelWidth, valueWidth, cursorVisible))
		// If this is the field owning an open popup, splice the option list
		// INTO the form parts so the user sees an inline dropdown (no
		// separate bordered dialog replacing the form). The rows sit directly
		// beneath the focused field inside the same DialogBox.
		if form.Popup.Open && form.Popup.Field == i && form.Popup.Kind != state.PopupPath {
			parts = appendFormPopupRows(parts, form, &form.Fields[i], popupInnerW)
		}
	}
	return parts
}

// appendFormPopupRows emits the rows of an open FormSelect / FormMultiSelect
// popup as additional form parts so the dropdown is rendered inline beneath
// the focused field. The path popup uses the separate renderPathPopup
// renderer because it's a multi-line directory/file picker that does not fit
// the inline form layout.
func appendFormPopupRows(parts []string, form state.FormState, field *state.FormField, innerW int) []string {
	rows := popupRows(field)
	if len(rows) == 0 {
		return parts
	}
	lines := visiblePopupRows(form, field, rows, innerW, state.FormPopupVisibleRows)
	for len(lines) < state.FormPopupVisibleRows {
		lines = append(lines, component.FormRow("", 0, innerW, ""))
	}
	return append(parts, lines...)
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
// the value column (BR-041 §9.8). HelperText is wrapped (never truncated) via
// utils.WrapCells — long descriptions flow onto rows beneath the value column
// (design P). All runes are preserved.
func renderFormField(form state.FormState, f *state.FormField, index, labelWidth, valueWidth int, cursorVisible ...bool) string {
	focused := form.FieldFocus == index

	labelText := f.Label
	var wrapLines []string
	if f.HelperText != "" {
		full := labelText + " (" + f.HelperText + ")"
		lines := utils.WrapCells(full, labelWidth)
		labelText = lines[0]
		if len(lines) > 1 {
			wrapLines = lines[1:]
		}
	}
	label := component.FormRow(labelText, labelWidth, -1, "")
	if focused {
		label = lipgloss.NewStyle().Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).Bold(true).Render(label)
	}

	impl := FormFieldFor(f)
	value := impl.Render(focused, valueWidth, cursorVisible...)
	row := label + " " + value
	if len(wrapLines) > 0 {
		indent := utils.PadVisible("", labelWidth+1)
		for _, extra := range wrapLines {
			row += "\n" + indent + extra
		}
	}
	if f.Error != "" {
		pad := utils.PadVisible("", labelWidth+1)
		row += "\n" + pad + lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleDialogError).GetForeground()).Render(f.Error)
	}
	return row
}

// renderFormValue renders the primitive value of a text/int/path/bool field.
// Thin wrapper that dispatches on Kind; the Bool branch delegates to
// renderBoolCell and the text branch to renderEditableValue. Lean impls call
// those helpers directly.
func renderFormValue(f *state.FormField, focused bool, valueWidth int, cursorVisible ...bool) string {
	switch f.Kind {
	case state.FormBool:
		return renderBoolCell(f.Toggle, focused)
	default: // FormText, FormInt, FormPath
		return renderEditableValue(f, focused, valueWidth, cursorVisible...)
	}
}

// renderEditableValue draws a text-like field. Inside existing text the
// cursor styles the current rune without consuming another terminal cell, so
// Left/Right never shifts the path. At end-of-input a one-cell caret is shown.
func renderEditableValue(f *state.FormField, focused bool, valueWidth int, cursorVisible ...bool) string {
	return renderTextCell(f.Input.Text, f.Input.Cursor, f.Kind, focused, valueWidth, cursorVisible...)
}

// renderTextCell draws a text-like value cell given primitive inputs. kind
// determines path-specific rendering (FormPath wraps unfocused long paths).
// This is the canonical implementation that lean FormField impls call.
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
		// When the user is not editing the field (typical for the default
		// Local destination (tar) value at form-open) the path is rendered
		// from the start with a tail ellipsis. This shows the directory the
		// file will land in, which is the meaningful context, instead of a
		// window anchored at the cursor that hides the directory part.
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

// renderBoolCell renders the styled checkbox for a Bool field.
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

// renderSelectCell renders the collapsed value plus a dropdown marker for a
// single- or multi-select field (BR-041 §8.1, §8.2). Thin wrapper around
// renderSelectCellValue; kept for tests that target *state.FormField.
func renderSelectCell(f *state.FormField, focused bool, valueWidth int) (value, marker string) {
	return renderSelectCellValue(f.Options, f.DisplayOptions, f.Index, f.Selected, f.Kind, focused, valueWidth)
}

// renderSelectCellValue draws the collapsed single- or multi-select cell
// given primitive inputs. This is the canonical implementation that lean
// FormField impls call.
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

func clampCursor(cursor, length int) int {
	if cursor < 0 {
		return 0
	}
	if cursor > length {
		return length
	}
	return cursor
}
