package dialog

import (
	"image/color"

	sty "github.com/elizabevil/docker-tui/internal/tui/ui/style"

	"charm.land/lipgloss/v2"
)

// DialogStyle groups visual parameters for a dialog box.
type DialogStyle struct {
	Width        int
	Height       int
	TitleColor   color.Color
	OverlayColor string
}

// DialogBox renders a bordered dialog container.
// parts are joined vertically (title, body, buttons, hint, etc.).
func DialogBox(style DialogStyle, parts ...string) string {
	if style.Width <= 0 {
		style.Width = 50
	}
	titleColor := style.TitleColor
	if titleColor == nil {
		titleColor = sty.Colors.Cyan
	}

	inner := lipgloss.JoinVertical(lipgloss.Top, parts...)

	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Foreground(titleColor).
		Padding(1, 2).
		Width(style.Width)

	if style.Height > 0 {
		s = s.Height(style.Height).MaxHeight(style.Height)
	}

	return s.Align(lipgloss.Center).Render(inner)
}
