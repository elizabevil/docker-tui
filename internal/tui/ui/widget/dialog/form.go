package dialog

import (
	"fmt"
	"image/color"
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
func FormDialog(m *state.AppModel, overlayColor string, cfg DialogConfig, bodyW, bodyH int) string {
	form := m.Form
	dialogW, dialogH := formDialogSize(m, cfg, bodyW, bodyH)
	if overlayColor == "" {
		overlayColor = OverlayColor(nil)
	}
	innerWidth := formInnerWidth(dialogW)

	enterKey := i18n.T("key.sym_enter")
	escKey := i18n.T("key.sym_esc")

	opStyles := containerOperationStyles(form.Kind)
	titleColor := component.GetStyle(component.StylePanelTitle).GetForeground()

	parts := []string{
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Left).
			Render(component.GetStyle(component.StylePanelTitle).Render(form.Title)),
	}
	parts = appendFormHeader(parts, form)
	parts = append(parts, "")
	labelWidth, valueWidth := formLayout(form.Fields, innerWidth)
	popupInnerW := labelWidth + valueWidth + 1
	parts = appendFormLoading(parts, form)
	parts = appendFormFields(parts, form, labelWidth, valueWidth, popupInnerW, !m.CursorBlinkHidden, opStyles)
	parts = append(parts, "")

	cancelBtn := NewButton(escKey, form.CancelLabel)
	cancelBtn.Foreground = opStyles.Cancel.Foreground
	cancelBtn.Background = opStyles.Cancel.Background
	confirmBtn := NewButton(enterKey, form.ConfirmLabel)
	confirmBtn.Foreground = opStyles.Confirm.Foreground
	confirmBtn.Background = opStyles.Confirm.Background
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

	dialogStyle := DialogStyle{
		Width:        dialogW,
		Height:       dialogH,
		MaxWidth:     cfg.PanelSize.MaxWidth,
		TitleColor:   titleColor,
		OverlayColor: overlayColor,
		LeftAligned:  true,
	}
	if opStyles.Window != (opStylesWindow{}) {
		dialogStyle.Background = opStyles.Window.Background
		dialogStyle.BorderColor = opStyles.Window.Border
	}
	box := DialogBox(dialogStyle, parts...)
	if form.Popup.Open && form.Popup.Kind == state.PopupPath {
		box = renderFormPopup(form, box, dialogW, dialogH)
	}
	return box
}

// opStylesWindow bundles a single container-operation window's background
// + border so the dialog can be drawn with one consistent chrome. The zero
// value is a sentinel meaning "use the existing default dialog chrome".
type opStylesWindow struct {
	Background color.Color
	Border     color.Color
}

// opStylesButton bundles the foreground + background colour for a single
// action button (Confirm or Cancel).
type opStylesButton struct {
	Foreground color.Color
	Background color.Color
}

// formOperationStyles is the resolved style bundle for one of the eight
// Action Bar container actions (or the zero-value fallback for forms that
// aren't part of the set).
type formOperationStyles struct {
	Window  opStylesWindow
	Confirm opStylesButton
	Cancel  opStylesButton
}

// containerOperationStyles resolves the per-form-kind style bundle. The four
// form-based container actions (Copy / Update / Export / Commit) consume the
// theme.action.container slots; any other form falls back to the zero value
// (no overrides), so the existing default dialog chrome remains unchanged.
//
// This dispatch is the container-scope branch of the broader Operation
// theme contract (see R06-01); when image-scope forms are added later,
// a parallel branch selecting theme.action.image is the one-line extension.
func containerOperationStyles(kind state.FormKind) formOperationStyles {
	if !isContainerOperationForm(kind) {
		return formOperationStyles{}
	}
	windowStyle := component.GetStyle(component.StyleActionContainerWindow)
	borderStyle := component.GetStyle(component.StyleActionContainerWindowBorder)
	confirmStyle := component.GetStyle(component.StyleActionContainerConfirm)
	cancelStyle := component.GetStyle(component.StyleActionContainerCancel)
	return formOperationStyles{
		Window: opStylesWindow{
			Background: windowStyle.GetBackground(),
			Border:     borderStyle.GetForeground(),
		},
		Confirm: opStylesButton{
			Foreground: confirmStyle.GetForeground(),
			Background: confirmStyle.GetBackground(),
		},
		Cancel: opStylesButton{
			Foreground: cancelStyle.GetForeground(),
			Background: cancelStyle.GetBackground(),
		},
	}
}

// isContainerOperationForm reports whether the given FormKind belongs to the
// four form-based container-scope Operations (Copy / Update / Export /
// Commit). It gates the application of theme.action.container.* slots in
// FormDialog.
func isContainerOperationForm(kind state.FormKind) bool {
	switch kind {
	case state.FormContainerCopy, state.FormContainerUpdate,
		state.FormContainerExport, state.FormContainerCommit:
		return true
	}
	return false
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

func appendFormFields(parts []string, form state.FormState, labelWidth, valueWidth, popupInnerW int, cursorVisible bool, opStyles formOperationStyles) []string {
	for i := range form.Fields {
		f := form.Fields[i]
		if f.Hidden() {
			continue
		}
		parts = append(parts, renderFormField(form, f, i, labelWidth, valueWidth, cursorVisible, opStyles))
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
func renderFormField(form state.FormState, f state.FormField, index, labelWidth, valueWidth int, cursorVisible bool, opStyles formOperationStyles) string {
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
		value = df.Render(focused, valueWidth, cursorVisible)
		value = wrapFormInputValue(value, valueWidth, opStyles)
	}
	var row strings.Builder
	row.WriteString(label + " " + value)
	if len(wrapLines) > 0 {
		indent := utils.PadVisible("", labelWidth+1)
		for _, extra := range wrapLines {
			row.WriteString("\n" + indent + extra)
		}
	}
	if f.Error() != "" {
		pad := utils.PadVisible("", labelWidth+1)
		row.WriteString("\n" + pad + lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleDialogError).GetForeground()).Render(f.Error()))
	}
	return row.String()
}

// wrapFormInputValue applies the formInput style (foreground / background)
// to a rendered value cell. The zero-value opStyles is a no-op so existing
// forms keep their original look; the container-operation forms override the
// default FormInput slot with their own StyleContainerOperationFormInput
// so the value cell matches the surrounding window chrome.
func wrapFormInputValue(value string, width int, opStyles formOperationStyles) string {
	if value == "" || width <= 0 {
		return value
	}
	style := component.GetStyle(component.StyleFormInput)
	fg := style.GetForeground()
	bg := style.GetBackground()
	if isContainerOperationFormFromStyles(opStyles) {
		formInput := component.GetStyle(component.StyleActionContainerFormInput)
		fg = formInput.GetForeground()
		bg = formInput.GetBackground()
	}
	s := lipgloss.NewStyle()
	if fg != nil {
		s = s.Foreground(fg)
	}
	if bg != nil {
		s = s.Background(bg)
	}
	return s.Width(width).Render(value)
}

// isContainerOperationFormFromStyles mirrors isContainerOperationForm but
// for the resolved style bundle: any non-zero Confirm field means the form
// is one of the four container-operation forms (Copy / Update / Export /
// Commit) and should consume the containerOperation.* theme slots.
func isContainerOperationFormFromStyles(opStyles formOperationStyles) bool {
	return opStyles.Confirm != (opStylesButton{})
}
