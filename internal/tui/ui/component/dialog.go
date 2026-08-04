package component

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/constants"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
)

func RenderShellDialogBox(shell string, termW int, overlayColor string, cursorVisible ...bool) string {
	dialogW := termW * 25 / 100
	if dialogW < 40 {
		dialogW = 40
	}
	if dialogW > 60 {
		dialogW = 60
	}
	if shell == "" {
		shell = "/bin/sh"
	}
	cursor := BlockCursor
	if len(cursorVisible) > 0 && !cursorVisible[0] {
		cursor = " "
	}
	inner := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Foreground(style.Color(constants.ColorWarning)).Render("Enter container shell"),
		"",
		lipgloss.NewStyle().Faint(true).Render("  Shell: "+shell+cursor),
		"",
		lipgloss.NewStyle().Faint(true).Render("  [Enter] Confirm  [Esc] Cancel"),
	)
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Foreground(style.Color(constants.ColorWarning)).
		Background(style.Color(overlayColor)).
		Padding(1, 2).
		Width(dialogW).
		Align(lipgloss.Center).
		Render(inner)
	return dialog
}

func RenderTextInputBox(title, value string, cursor, termW int, overlayColor string, cursorVisible ...bool) string {
	runes := []rune(value)
	cursor = min(max(0, cursor), len(runes))
	mark := BlockCursor
	if len(cursorVisible) > 0 && !cursorVisible[0] {
		mark = " "
	}
	input := string(runes[:cursor]) + mark + string(runes[cursor:])
	dialogW := min(60, max(40, termW*25/100))
	inner := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Foreground(style.Color(constants.ColorPrimary)).Render(title),
		"",
		lipgloss.NewStyle().Render(input),
		"",
		lipgloss.NewStyle().Faint(true).Render("[Enter] Confirm  [Esc] Cancel"),
	)
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(dialogW).Render(inner)
}

func RenderProgressDialogBox(title, target, status string, current, total int64, termW int, overlayColor string) string {
	progress := status
	if total > 0 {
		progress = fmt.Sprintf("%s  %d / %d bytes", status, current, total)
	} else if current > 0 {
		progress = fmt.Sprintf("%s  %d bytes", status, current)
	}
	inner := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Foreground(style.Color(constants.ColorPrimary)).Render(title),
		"", target, "", progress, "",
		lipgloss.NewStyle().Faint(true).Render("[Esc] "+i18n.T("key.cancel")),
	)
	dialogW := min(70, max(40, termW*35/100))
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(dialogW).Render(inner)
}

// PlaceOverlay centers a dialog over the terminal area WITHOUT a full-screen
// backdrop. The dialog's own background provides contrast; surrounding content
// remains visible through the terminal background.
func PlaceOverlay(termW, termH int, dialog string, overlayColor string) string {
	return lipgloss.Place(termW, termH,
		lipgloss.Center, lipgloss.Center, dialog,
	)
}
