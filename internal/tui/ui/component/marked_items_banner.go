package component

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// MarkedItemsBanner summarizes marks owned by the active panel.
func MarkedItemsBanner(panel state.PanelType, marks map[string]bool) string {
	if len(marks) == 0 {
		return ""
	}
	return GetStyle(StyleFooter).Render(fmt.Sprintf("Marked: %d  (Esc to exit, clear marks)", len(marks)))
}
