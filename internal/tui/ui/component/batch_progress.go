package component

import (
	"fmt"
	"strings"
)

// BatchProgressIndicator renders the progress of a fan-out batch operation
// (e.g. "2/5 containers stopped"). It is business-agnostic: the caller
// supplies the action text and the current/total counts.
type BatchProgressIndicator struct {
	Action  string // i18n label of the running action, e.g. "stopping"
	Current int    // number of targets completed so far
	Total   int    // total number of targets
}

// NewBatchProgressIndicator creates a progress indicator.
func NewBatchProgressIndicator(action string, current, total int) BatchProgressIndicator {
	return BatchProgressIndicator{Action: action, Current: current, Total: total}
}

// Done reports whether every target has been processed.
func (p BatchProgressIndicator) Done() bool {
	return p.Total > 0 && p.Current >= p.Total
}

// Render draws the indicator: a percentage block and a "current/total action"
// text line. When no targets exist it renders an empty string.
func (p BatchProgressIndicator) Render(width int) string {
	if width <= 0 {
		width = 40
	}
	if p.Total <= 0 {
		return ""
	}
	pct := p.Current * 100 / p.Total
	if pct > 100 {
		pct = 100
	}
	barW := width - 12 // reserve room for " 3/5 · stopping"
	if barW < 4 {
		barW = 4
	}
	filled := pct * barW / 100
	bar := strings.Repeat(BlockCursor, filled) +
		strings.Repeat(BoxHorizontal, barW-filled)
	text := fmt.Sprintf("%d/%d %s", p.Current, p.Total, p.Action)
	return fmt.Sprintf("[%s] %3d%%  %s", bar, pct, text)
}
