package view

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func TestCalculateRailHeights(t *testing.T) {
	for _, height := range []int{20, 24, 32, 50} {
		rails := calculateRailHeights(height)
		if rails.header != headerRailHeight || rails.message != messageRailHeight ||
			rails.query != queryRailHeight || rails.footer != footerRailHeight {
			t.Fatalf("fixed rails changed at height %d: %+v", height, rails)
		}
		if rails.total() != height {
			t.Fatalf("rail total at height %d = %d, want %d", height, rails.total(), height)
		}
	}
}

func TestPanelBodyHeight(t *testing.T) {
	if got := panelBodyHeight(10); got != 7 {
		t.Fatalf("panelBodyHeight(10) = %d, want 7", got)
	}
	if got := panelBodyHeight(2); got != 1 {
		t.Fatalf("panelBodyHeight(2) = %d, want 1", got)
	}
}

func TestQueryKindFor(t *testing.T) {
	tests := []struct {
		name string
		app  *state.AppModel
		want queryKind
	}{
		{name: "normal", app: &state.AppModel{Mode: state.ModeNormal}, want: queryNone},
		{name: "filter", app: &state.AppModel{Mode: state.ModeFilter}, want: queryFilter},
		{name: "search", app: &state.AppModel{Mode: state.ModeFilter, LogContainerID: "abc"}, want: querySearch},
		{name: "command", app: &state.AppModel{Mode: state.ModeCommand}, want: queryCommand},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := queryKindFor(tt.app); got != tt.want {
				t.Fatalf("queryKindFor() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRenderQueryInputUsesResourceSemantics(t *testing.T) {
	filter := renderQueryInput(queryFilter, "nginx", 5, 100)
	search := renderQueryInput(querySearch, "error", 5, 100)
	if !strings.Contains(filter, "Filter:") {
		t.Fatalf("filter query missing Filter label: %q", filter)
	}
	if !strings.Contains(search, "Search:") {
		t.Fatalf("search query missing Search label: %q", search)
	}
}

func TestFitRailHeight(t *testing.T) {
	if got := strings.Count(fitRailHeight("one", 3), "\n") + 1; got != 3 {
		t.Fatalf("fitRailHeight rows = %d, want 3", got)
	}
	if got := fitRailHeight("one\ntwo\nthree", 2); got != "one\ntwo" {
		t.Fatalf("fitRailHeight truncation = %q", got)
	}
}

func TestRenderAppHeightDoesNotChangeWithQueryOrMessage(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Width = 120
	app.Height = 32

	normalHeight := strings.Count(RenderApp(app), "\n") + 1
	app.Mode = state.ModeFilter
	app.FilterText = "nginx"
	filterHeight := strings.Count(RenderApp(app), "\n") + 1
	app.ToastMessage = "filter applied"
	app.ToastLevel = component.ToastInfo
	messageHeight := strings.Count(RenderApp(app), "\n") + 1

	if filterHeight != normalHeight || messageHeight != normalHeight {
		t.Fatalf(
			"layout height changed: normal=%d filter=%d message=%d",
			normalHeight, filterHeight, messageHeight,
		)
	}
}
