package component

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/utils"
)

// WaitIndicator renders an in-flight container wait as a compact
// "Waiting on <short-id>" line. The Message rail layer paints it with
// the Warning style so the user sees both progress (the indicator)
// and risk (the colour).
//
// targetID may be the full container hash; Render truncates to the
// 12-char short form. generation is reserved for a future
// stale-generation indicator and currently unused — passed in so the
// caller can record it for diagnostics without a separate call.
type WaitIndicator struct {
	TargetID   string
	Generation uint64
}

// NewWaitIndicator creates a wait indicator for the given target.
func NewWaitIndicator(targetID string, generation uint64) WaitIndicator {
	return WaitIndicator{TargetID: targetID, Generation: generation}
}

// Render draws the indicator as "⏳ Waiting on <short-id>". The width
// argument is currently ignored — the rail's wrapMessageRail handles
// wrapping — but is kept for symmetry with BatchProgressIndicator.Render
// and for future use when the indicator grows a progress bar.
func (w WaitIndicator) Render(width int) string {
	if w.TargetID == "" {
		return "⏳ Waiting…"
	}
	return fmt.Sprintf("⏳ Waiting on %s…", utils.ShortID(w.TargetID))
}