package actionbar

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	actionmodel "github.com/elizabevil/docker-tui/internal/tui/actionbar"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/dialog"
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
