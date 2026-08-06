package dialog

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"

	"charm.land/lipgloss/v2"
)

// FormDialog renders the container-action form as a two-column layout: a
// right-aligned label column and a shared value column (BR-041 §9). The
// Confirm / Cancel pair stays in the same Tab loop as the fields.
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
	popupInnerW := labelWidth + valueWidth + 1
	parts = appendFormLoading(parts, form)
	parts = appendFormFields(parts, form, labelWidth, valueWidth, popupInnerW, !m.CursorBlinkHidden)
	parts = append(parts, "")

	cancelBtn := NewButton(escKey, form.CancelLabel)
	confirmBtn := NewButton(enterKey, form.ConfirmLabel)
	if form.FieldFocus == form.CancelSlot() && !form.Popup.Open {
		cancelBtn.State = ButtonFocused
	}
	if form.FieldFocus == form.ConfirmSlot() && !form.Popup.Open {
		confirmBtn.State = ButtonFocused
	}
	buttons := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(
		cancelBtn.Render() + "   " + confirmBtn.Render(),
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
	var labelW int
	if innerW := formInnerWidth(dialogW); innerW > 0 && len(m.Form.Fields) > 0 {
		labelW, _ = formLayout(m.Form.Fields, innerW)
	}
	requiredH := 10
	for i := range m.Form.Fields {
		f := m.Form.Fields[i]
		if f.Hidden() {
			continue
		}
		requiredH++
		df, ok := f.(FormField)
		if !ok {
			continue
		}
		if labelW > 0 && df.HelperText() != "" {
			full := df.Label() + " (" + df.HelperText() + ")"
			if extra := len(utils.WrapCells(full, labelW)) - 1; extra > 0 {
				requiredH += extra
			}
		}
		if f.Error() != "" {
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
		f := form.Fields[i]
		if f.Hidden() {
			continue
		}
		parts = append(parts, renderFormField(form, f, i, labelWidth, valueWidth, cursorVisible))
		if form.Popup.Open && form.Popup.Field == i && form.Popup.Kind != state.PopupPath {
			parts = appendFormPopupRows(parts, form, f, popupInnerW)
		}
	}
	return parts
}

func appendFormPopupRows(parts []string, form state.FormState, field state.FormField, innerW int) []string {
	df, _ := field.(FormField)
	if df == nil {
		return parts
	}
	rows := popupRows(df)
	if len(rows) == 0 {
		return parts
	}
	lines := visiblePopupRows(form, df, rows, innerW, state.FormPopupVisibleRows)
	for len(lines) < state.FormPopupVisibleRows {
		lines = append(lines, component.FormRow("", 0, innerW, ""))
	}
	return append(parts, lines...)
}

func formInnerWidth(dialogW int) int {
	if dialogW <= 4 {
		return 1
	}
	return dialogW - 4
}

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
		df, ok := fields[i].(FormField)
		if !ok {
			labels = append(labels, "")
			continue
		}
		labels = append(labels, df.Label())
	}
	labelWidth = component.FormLabelColumnWidth(labels, innerWidth)
	valueWidth = innerWidth - labelWidth - 1
	if valueWidth < 1 {
		valueWidth = 1
	}
	return labelWidth, valueWidth
}

// renderFormField draws one two-column row for a form field.
func renderFormField(form state.FormState, f state.FormField, index, labelWidth, valueWidth int, cursorVisible ...bool) string {
	focused := form.FieldFocus == index
	df, _ := f.(FormField)

	labelText := ""
	helperText := ""
	if df != nil {
		labelText = df.Label()
		helperText = df.HelperText()
	}

	var wrapLines []string
	if helperText != "" {
		full := labelText + " (" + helperText + ")"
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

	value := ""
	if df != nil {
		value = df.Render(focused, valueWidth, cursorVisible...)
	}
	row := label + " " + value
	if len(wrapLines) > 0 {
		indent := utils.PadVisible("", labelWidth+1)
		for _, extra := range wrapLines {
			row += "\n" + indent + extra
		}
	}
	if f.Error() != "" {
		pad := utils.PadVisible("", labelWidth+1)
		row += "\n" + pad + lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleDialogError).GetForeground()).Render(f.Error())
	}
	return row
}