package dialog

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// PlaceDialog centers dialogBox over content WITHOUT a full-screen scrim.
// Only the dialog box itself carries the dark background (set by DialogBox).
// Content outside the dialog area remains completely visible.
func PlaceDialog(content string, dialogBox string, termW, termH int, _ string, cfg dialogConfig) string {
	contentLines := strings.Split(content, "\n")
	dialogLines := strings.Split(dialogBox, "\n")

	dlgH := len(dialogLines)
	if dlgH == 0 {
		return content
	}

	dlgW := 0
	for _, line := range dialogLines {
		w := lipgloss.Width(line)
		if w > dlgW {
			dlgW = w
		}
	}
	startX, startY := dialogPosition(termW, termH, dlgW, dlgH, cfg)

	for dy := 0; dy < dlgH && startY+dy < len(contentLines); dy++ {
		dl := dialogLines[dy]
		visW := lipgloss.Width(dl)

		leftPad := startX
		padded := strings.Repeat(" ", leftPad) + dl
		rightPad := termW - startX - visW
		if rightPad > 0 {
			padded += strings.Repeat(" ", rightPad)
		}
		contentLines[startY+dy] = padded
	}

	return strings.Join(contentLines, "\n")
}
