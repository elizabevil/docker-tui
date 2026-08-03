package dialog

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// execShellOptions are the default shell choices in the exec dialog.
var execShellOptions = []string{"/bin/sh", "/bin/bash", "/bin/ash"}

// ExecDialog renders the container exec dialog with shell options + custom input.
// Focus and input cursor are owned by DialogState.
//
// When bodyW > 0 and bodyH > 0, dialog sizing follows the active panel body
// (BR-043 §3.2: width = bodyW*3/4, height = bodyH, clamped to cfg). Otherwise
// it falls back to viewport-percentage sizing.
func ExecDialog(m *state.AppModel, overlayColor string, cfg dialogConfig, bodyW, bodyH int) string {
	var dialogW, dialogH int
	if bodyW > 0 && bodyH > 0 {
		dialogW = panelDialogWidth(bodyW, cfg)
		dialogH = panelDialogHeight(bodyH, cfg)
	} else {
		dialogW = dialogWidth(m.Viewport.Width, cfg)
		dialogH = dialogHeight(m.Viewport.Height, cfg)
	}
	if overlayColor == "" {
		overlayColor = "#0d1117cc"
	}

	enterKey := i18n.T("key.sym_enter")
	escKey := i18n.T("key.sym_esc")
	confirmLabel := i18n.T("key.confirm")
	cancelLabel := i18n.T("key.cancel")

	// Shell option buttons
	var optionBtns []string
	for i, opt := range execShellOptions {
		if m.Dialog.Focus == i {
			optionBtns = append(optionBtns, lipgloss.NewStyle().Foreground(style.Colors.Green).Bold(true).Render("\u25b6 "+opt))
		} else {
			optionBtns = append(optionBtns, lipgloss.NewStyle().Foreground(style.Colors.Gray).Render(opt))
		}
	}
	optionsRow := lipgloss.NewStyle().Width(dialogW - 4).Align(lipgloss.Center).Render(
		strings.Join(optionBtns, "   "),
	)

	// Custom input field with cursor
	inputText := m.Dialog.Input.Text
	if inputText == "" {
		inputText = "/bin/sh"
	}
	inputRunes := []rune(inputText)
	cursor := m.Dialog.Input.Cursor
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(inputRunes) {
		cursor = len(inputRunes)
	}
	inputDisplay := component.GetStyle("dim").Render(i18n.T("inspect.shell")+": ") + string(inputRunes[:cursor])
	if m.Dialog.Focus == execFocusInput && !m.CursorBlinkHidden {
		inputDisplay += "\u2588" // block cursor when focused
	} else {
		inputDisplay += " " // space when not focused
	}
	if cursor < len(inputRunes) {
		inputDisplay += string(inputRunes[cursor:])
	}

	// Confirm / Cancel buttons with shortcut hints
	var confirmBtn, cancelBtn string
	switch {
	case m.Dialog.Focus == 4:
		confirmBtn = lipgloss.NewStyle().Foreground(style.Colors.Green).Bold(true).Render(enterKey + " \u25b6 " + confirmLabel)
		cancelBtn = lipgloss.NewStyle().Foreground(style.Colors.Gray).Render(escKey + " " + cancelLabel)
	case m.Dialog.Focus == 5:
		confirmBtn = lipgloss.NewStyle().Foreground(style.Colors.Gray).Render(enterKey + " " + confirmLabel)
		cancelBtn = lipgloss.NewStyle().Foreground(style.Colors.Green).Bold(true).Render(escKey + " \u25b6 " + cancelLabel)
	default:
		confirmBtn = lipgloss.NewStyle().Foreground(style.Colors.Gray).Render(enterKey + " " + confirmLabel)
		cancelBtn = lipgloss.NewStyle().Foreground(style.Colors.Gray).Render(escKey + " " + cancelLabel)
	}
	buttons := lipgloss.NewStyle().Width(dialogW - 4).Align(lipgloss.Center).Render(
		confirmBtn + "   " + cancelBtn,
	)

	hint := i18n.T("hint.tab_switch")

	var parts []string
	parts = append(parts, component.GetStyle("panelTitle").Render(i18n.T("hint.enter_shell")))
	parts = append(parts, "")
	parts = append(parts, component.GetStyle("dim").Render("Select shell:"))
	parts = append(parts, optionsRow)
	parts = append(parts, "")
	parts = append(parts, inputDisplay)
	parts = append(parts, "")
	parts = append(parts, buttons)
	parts = append(parts, "")
	parts = append(parts, component.GetStyle("dim").Render(hint))

	return DialogBox(DialogStyle{Width: dialogW, Height: dialogH, TitleColor: style.Colors.Green, OverlayColor: overlayColor}, parts...)
}

// execFocusInput is the focus position for the custom input field.
// Must match keyboard/exec_dialog.go's execFocusInput.
const execFocusInput = 3
