package dialog

import (
	"image/color"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
)

// SelectionDialog renders a dialog with Tab-focusable Confirm/Cancel buttons.
// focus == 0 highlights Confirm; focus == 1 highlights Cancel.
// Buttons show Enter (Confirm) and Esc (Cancel) shortcut hints.
//
// When bodyW > 0 and bodyH > 0, dialog sizing follows the active panel body
// (BR-043 §3.2: width = bodyW*3/4, height = bodyH, clamped to cfg). Otherwise
// it falls back to termW/termH-percentage sizing.
func SelectionDialog(title, body, preview, action string, focus int, termW, termH int, bodyW, bodyH int, titleColor color.Color, overlayColor string, cfg DialogConfig) string {
	if titleColor == nil {
		titleColor = component.GetStyle("panelTitle").GetForeground()
	}
	if overlayColor == "" {
		overlayColor = OverlayColor(nil)
	}
	var dialogW, dialogH int
	if bodyW > 0 && bodyH > 0 {
		dialogW = panelDialogWidth(bodyW, cfg)
		dialogH = panelDialogHeight(bodyH, cfg)
	} else {
		dialogW = dialogWidth(termW, cfg)
		dialogH = dialogHeight(termH, cfg)
	}

	confirmLabel := i18n.T("key.confirm")
	cancelLabel := i18n.T("key.cancel")
	enterKey := i18n.T("key.sym_enter")
	escKey := i18n.T("key.sym_esc")

	var confirmBtn, cancelBtn string
	if focus == 0 {
		confirmBtn = lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render(enterKey + " " + component.ButtonIndicator + " " + confirmLabel)
		cancelBtn = lipgloss.NewStyle().Foreground(component.GetStyle("dim").GetForeground()).Render(escKey + " " + cancelLabel)
	} else {
		confirmBtn = lipgloss.NewStyle().Foreground(component.GetStyle("dim").GetForeground()).Render(enterKey + " " + confirmLabel)
		cancelBtn = lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render(escKey + " " + component.ButtonIndicator + " " + cancelLabel)
	}
	buttons := lipgloss.NewStyle().Width(dialogW - 4).Align(lipgloss.Center).Render(
		confirmBtn + "   " + cancelBtn,
	)

	hint := i18n.T("hint.tab_switch")

	var parts []string
	parts = append(parts, component.GetStyle("panelTitle").Render(title))
	parts = append(parts, "")
	parts = append(parts, body)
	if preview != "" {
		parts = append(parts, "")
		parts = append(parts, component.GetStyle("dim").Render("Preview:"))
		parts = append(parts, preview)
	}
	parts = append(parts, "")
	parts = append(parts, buttons)
	parts = append(parts, "")
	parts = append(parts, component.GetStyle("dim").Render(hint))

	return DialogBox(DialogStyle{Width: dialogW, Height: dialogH, TitleColor: titleColor, OverlayColor: overlayColor}, parts...)
}
