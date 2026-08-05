package dialog

import (
	"image/color"
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	sty "github.com/elizabevil/docker-tui/internal/tui/ui/style"

	"charm.land/lipgloss/v2"
)

// DialogStyle groups visual parameters for a dialog box.
type DialogStyle struct {
	Width        int
	Height       int
	TitleColor   color.Color
	OverlayColor string
	LeftAligned  bool
}

// DialogBox renders a bordered dialog container.
// parts are joined vertically (title, body, buttons, hint, etc.).
func DialogBox(style DialogStyle, parts ...string) string {
	if style.Width <= 0 {
		style.Width = 50
	}
	titleColor := style.TitleColor
	if titleColor == nil {
		titleColor = sty.Colors.Primary
	}

	inner := lipgloss.JoinVertical(lipgloss.Top, parts...)

	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Foreground(titleColor).
		Padding(1, 2).
		Width(style.Width)
	bodyBg := component.GetStyle(component.StyleDialogBodyBackground).GetBackground()
	if bodyBg != nil {
		s = s.Background(bodyBg)
	}

	if style.Height > 0 {
		s = s.Height(style.Height).MaxHeight(style.Height)
	}

	if style.LeftAligned {
		return s.Align(lipgloss.Left).Render(inner)
	}
	return s.Align(lipgloss.Center).Render(inner)
}

// opaqueDialogColor converts the configured overlay/scrim colour into the
// solid dialog surface colour supported by terminals. Alpha is meaningful to
// compositors, not terminal SGR background sequences.
func opaqueDialogColor(value string) string {
	value = strings.TrimSpace(value)
	if len(value) == 9 && strings.HasPrefix(value, "#") {
		return value[:7]
	}
	return value
}
