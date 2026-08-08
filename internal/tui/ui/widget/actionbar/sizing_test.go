// Test fixture: the action bar should be 1/2 of panel width and at most
// 2/3 of panel rows when rendered into a wide enough panel.
package actionbar

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/dialog"
	"github.com/elizabevil/docker-tui/internal/utils"
)

func TestRenderBarWidthIsHalfPanel(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActionBar.Filtering = true
	m.Navigation.ActionBar.Filter = "cpy"
	body := dialog.PanelBody{Left: 2, Top: 2, Width: 80, Rows: 30}
	rendered := utils.StripANSI(RenderBar(m, strings.Repeat(strings.Repeat(" ", 80)+"\n", 30), body))
	// Find the top border line.
	for line := range strings.SplitSeq(rendered, "\n") {
		clean := strings.TrimSpace(line)
		if strings.HasPrefix(clean, "╭") {
			// Width = 1/2 of 80 = 40, plus 2 border cells = 42 visible
			if w := utils.DisplayWidth(clean); w < 40 || w > 44 {
				t.Fatalf("action bar width = %d, want ~42 (1/2 of panel 80 + border)", w)
			}
			return
		}
	}
	t.Fatalf("no border line found")
}

func TestRenderBarHeightIsTwoThirdsPanel(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActionBar.Filtering = true
	m.Navigation.ActionBar.Filter = "cpy"
	body := dialog.PanelBody{Left: 2, Top: 2, Width: 80, Rows: 30}
	rendered := utils.StripANSI(RenderBar(m, strings.Repeat(strings.Repeat(" ", 80)+"\n", 30), body))
	// Count top border through bottom border.
	lines := strings.Split(rendered, "\n")
	in := 0
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "╭") {
			in = 1
			continue
		}
		if in > 0 {
			if strings.HasPrefix(strings.TrimSpace(line), "╰") {
				return // counted
			}
			in++
		}
	}
	// Total height should be 2/3 of 30 = 20 (plus borders/no border).
	if in > 22 {
		t.Fatalf("action bar height = %d, exceeds 2/3 of panel 30 + borders", in)
	}
}
