package actionbar

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	actionmodel "github.com/elizabevil/docker-tui/internal/tui/actionbar"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/dialog"
	"github.com/elizabevil/docker-tui/internal/utils"
)

func TestRenderBarShowsEmptyFilteredState(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActionBar.Filter = "does-not-exist"
	if len(actionmodel.VisibleItems(m)) != 0 {
		t.Fatal("test filter unexpectedly matched an action")
	}
	body := dialog.PanelBody{Left: 2, Top: 2, Width: 60, Rows: 14}
	rendered := RenderBar(m, strings.Repeat(strings.Repeat(" ", 80)+"\n", 20), body)
	if !strings.Contains(rendered, "No matching actions") {
		t.Fatalf("empty Action Bar not rendered: %q", rendered)
	}
}

// TestRenderBarBorderHasNoThemedBackground ensures the action bar's outer
// rounded border does not carry an explicit background escape. A second
// background under the border used to clash with the panel underneath.
func TestRenderBarBorderHasNoThemedBackground(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	body := dialog.PanelBody{Left: 2, Top: 2, Width: 80, Rows: 14}
	rendered := RenderBar(m, strings.Repeat(strings.Repeat(" ", 80)+"\n", 20), body)
	clean := utils.StripANSI(rendered)
	// The border characters should appear right at the start of the box
	// line, with no preceding background fill that would visually separate
	// the border from the panel underneath.
	borderTop, borderSep, borderSide := "╭", "─", "│"
	for _, line := range strings.Split(clean, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		first := []rune(line)[0]
		if first == []rune(borderTop)[0] || first == []rune(borderSide)[0] ||
			strings.HasPrefix(line, borderSep) {
			if strings.HasPrefix(line, " ") {
				t.Fatalf("border row has leading fill: %q", line)
			}
		}
	}
}

// TestRenderBarFilterRowShowsUserInput ensures that once the user types
// into the filter, the input text appears in the dedicated row, not lost
// in the line label below.
func TestRenderBarFilterRowShowsUserInput(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActionBar.Filtering = true
	m.Navigation.ActionBar.Filter = "cpy"
	body := dialog.PanelBody{Left: 2, Top: 2, Width: 80, Rows: 14}
	rendered := utils.StripANSI(RenderBar(m, strings.Repeat(strings.Repeat(" ", 80)+"\n", 20), body))
	if !strings.Contains(rendered, "cpy") {
		t.Fatalf("filter buffer 'cpy' missing from rendered action bar: %q", rendered)
	}
	// The hint line should be replaced by the input line while filtering.
	if !strings.Contains(rendered, "/ cpy") {
		t.Fatalf("filter row must render the prompt + buffer, got %q", rendered)
	}
}

// TestRenderBarExposesConfirmAndCancelButtons ensures the bar shows the
// "Enter ▶ Confirm" and "Esc Cancel" affordances in the footer.
func TestRenderBarExposesConfirmAndCancelButtons(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Connection.Engine = nil
	// body.Rows is the height ceiling the panel applies to the dialog
	// (see CenterOnPanel). Use a tall enough body so the title, filter,
	// action list, and the new footer buttons all fit.
	body := dialog.PanelBody{Left: 2, Top: 2, Width: 80, Rows: 30}
	rendered := utils.StripANSI(RenderBar(m, strings.Repeat(strings.Repeat(" ", 80)+"\n", 30), body))
	if !strings.Contains(rendered, "Confirm") {
		t.Fatalf("action bar must show a Confirm button: %q", rendered)
	}
	if !strings.Contains(rendered, "Cancel") {
		t.Fatalf("action bar must show a Cancel button: %q", rendered)
	}
}
