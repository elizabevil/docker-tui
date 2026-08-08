package dialog

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// PlaceDialog centers dialogBox over content WITHOUT a full-screen scrim.
// Only the dialog box itself carries the dark background (set by DialogBox).
// Content outside the dialog area remains completely visible.
func PlaceDialog(content string, dialogBox string, termW, termH int, _ string, cfg DialogConfig) string {
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
		if visW < dlgW {
			dl += strings.Repeat(" ", dlgW-visW)
		}

		base := contentLines[startY+dy]
		left := ansi.TruncateWc(base, startX, "")
		left += strings.Repeat(" ", max(0, startX-ansi.StringWidth(left)))
		rightStart := startX + dlgW
		right := ansi.TruncateLeftWc(ansi.TruncateWc(base, termW, ""), rightStart, "")
		right += strings.Repeat(" ", max(0, termW-rightStart-ansi.StringWidth(right)))
		contentLines[startY+dy] = left + ansiReset + dl + ansiReset + right
	}

	return strings.Join(contentLines, "\n")
}
