package dialog

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const ansiReset = "\x1b[0m"

// PanelBody describes the rectangular region of a panel body in
// terminal coordinates where a dialog may be centered.
//
// Per BR-040, dialogs are centered within the body of the active
// panel rather than the entire terminal. LayoutReport.Panel carries
// the same geometry (panelTop / bodyTop / bodyRows / bodyLeft / bodyWidth)
// and is converted to PanelBody at the call site.
type PanelBody struct {
	Left  int // X of the first column inside the panel body
	Top   int // Y of the first row inside the panel body
	Width int // inner width excluding left + right border
	Rows  int // inner height excluding top + bottom border
}

// CenterOnPanel places a rendered dialogBox into the content string,
// centering it within body. Outside the dialog region the content
// remains untouched, achieving the BR-040 "transparent border + center
// window" visual contract.
//
// Behavior (per docs/task/br-040-implementation-decisions.md):
//   - dialog content larger than body: clamped to body size; wide
//     dialog lines are ANSI-truncated; tall dialogs drop trailing rows.
//   - dialog never covers terminal header / footer rails (limited by
//     body's Top + Rows).
//   - dialog content outside body is silently dropped (no overflow
//     into adjacent panels or rails).
//
// termW, termH bound the spliced content so that lines wider than
// the terminal are not produced. The dialog is centered within body,
// not within the full terminal.
func CenterOnPanel(content, dialogBox string, body PanelBody, termW, termH int, cfg dialogConfig) string {
	if dialogBox == "" {
		return content
	}

	contentLines := strings.Split(content, "\n")
	dialogLines := strings.Split(dialogBox, "\n")

	if len(dialogLines) == 0 {
		return content
	}

	// Limit dialog to fit within both body and terminal.
	maxW := body.Width
	if avail := termW - body.Left; avail < maxW {
		maxW = avail
	}
	if maxW < 1 {
		maxW = 1
	}
	maxH := body.Rows
	if avail := termH - body.Top; avail < maxH {
		maxH = avail
	}
	if maxH < 1 {
		maxH = 1
	}

	dlgW := 0
	for _, line := range dialogLines {
		if w := lipgloss.Width(line); w > dlgW {
			dlgW = w
		}
	}
	effW := dlgW
	if effW > maxW {
		effW = maxW
	}
	effH := len(dialogLines)
	if effH > maxH {
		effH = maxH
	}

	startX := body.Left + (maxW-effW)/2
	if startX < body.Left {
		startX = body.Left
	}
	if startX > termW {
		startX = termW
	}
	startY := body.Top + (maxH-effH)/2
	if startY < body.Top {
		startY = body.Top
	}
	if startY > termH {
		startY = termH
	}

	for dy := 0; dy < effH; dy++ {
		line := startY + dy
		if line >= len(contentLines) {
			break
		}
		// Truncate dialog line to effW (ANSI-aware).
		dl := ansi.TruncateWc(dialogLines[dy], effW, "")
		if visW := lipgloss.Width(dl); visW < effW {
			dl += strings.Repeat(" ", effW-visW)
		}

		base := contentLines[line]
		left := ansi.TruncateWc(base, startX, "")
		left += strings.Repeat(" ", max(0, startX-ansi.StringWidth(left)))
		rightStart := startX + effW
		right := ansi.TruncateLeftWc(ansi.TruncateWc(base, termW, ""), rightStart, "")
		right += strings.Repeat(" ", max(0, termW-rightStart-ansi.StringWidth(right)))
		// Reset at both splice boundaries. Without this, a selected table row's
		// background leaks into transparent dialog cells and moves on repaint.
		contentLines[line] = left + ansiReset + dl + ansiReset + right
	}

	return strings.Join(contentLines, "\n")
}

// PlaceDialogInPanel is a high-level wrapper around CenterOnPanel that
// derives termW / termH from content itself. Callers that already hold
// the terminal dimensions can call CenterOnPanel directly to skip the
// ANSI-aware line measurement.
func PlaceDialogInPanel(content, dialogBox string, body PanelBody, cfg dialogConfig) string {
	contentLines := strings.Split(content, "\n")
	termH := len(contentLines)
	termW := 0
	for _, line := range contentLines {
		if w := ansi.StringWidth(line); w > termW {
			termW = w
		}
	}
	return CenterOnPanel(content, dialogBox, body, termW, termH, cfg)
}

// CenterOnPanelDefault is the cross-package entry point for app-level
// overlays that already know the terminal dimensions.
func CenterOnPanelDefault(content, dialogBox string, body PanelBody, termW, termH int) string {
	return CenterOnPanel(content, dialogBox, body, termW, termH, dialogConfig{})
}
