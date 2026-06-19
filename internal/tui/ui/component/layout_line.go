package component

import "github.com/elizabevil/docker-tui/internal/tui/utils"

// JustifyBetween returns a single line with left/right text aligned to edges.
// When width is too small, it degrades gracefully to left text only.
func JustifyBetween(left, right string, width int) string {
	if width <= 0 {
		if right == "" {
			return left
		}
		return left + " " + right
	}
	if right == "" {
		return utils.PadVisible(left, width)
	}
	lw := utils.VisibleLen(left)
	rw := utils.VisibleLen(right)
	if lw+1+rw > width {
		return utils.PadVisible(utils.TruncateVisible(left, width), width)
	}
	gap := width - lw - rw
	if gap < 1 {
		gap = 1
	}
	return left + RepeatSpaces(gap) + right
}

func RepeatSpaces(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}
