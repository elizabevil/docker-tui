package dialog

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

type ChoiceOption struct {
	Label       string
	Description string
	Disabled    bool
}

func ChoiceDialog(title, body string, options []ChoiceOption, focus, termW, termH int, titleColor color.Color, overlayColor string, cfg DialogConfig) string {
	dialogW, dialogH := choiceDialogSize(termW, termH, cfg)
	if titleColor == nil {
		titleColor = component.GetStyle("panelTitle").GetForeground()
	}
	innerW := max(1, dialogW-6)
	body = lipgloss.NewStyle().Width(innerW).Render(body)
	parts := []string{component.GetStyle("panelTitle").Render(title), "", body, ""}
	optionLabels := make([]string, 0, len(options))
	for i, option := range options {
		label := option.Label
		if option.Description != "" {
			label += "  " + option.Description
		}
		optionStyle := lipgloss.NewStyle().Foreground(component.GetStyle("dim").GetForeground())
		if option.Disabled {
			optionStyle = optionStyle.Faint(true)
		} else if i == focus {
			optionStyle = lipgloss.NewStyle().Foreground(titleColor).Bold(true)
			label = "[ ▶ " + label + " ]"
		} else {
			label = "[   " + label + " ]"
		}
		optionLabels = append(optionLabels, optionStyle.Render(label))
	}
	parts = append(parts, lipgloss.NewStyle().Width(innerW).Align(lipgloss.Center).Render(strings.Join(optionLabels, "   ")))
	parts = append(parts, "", component.GetStyle("dim").Render("Tab/Shift+Tab switch  Enter select  Esc cancel"))
	return DialogBox(DialogStyle{Width: dialogW, Height: dialogH, TitleColor: titleColor, OverlayColor: overlayColor}, parts...)
}

func choiceDialogSize(termW, termH int, cfg DialogConfig) (int, int) {
	if termW <= 0 || termH <= 0 {
		return dialogWidth(termW, cfg), dialogHeight(termH, cfg)
	}
	return max(20, termW/2), max(8, termH/2)
}
