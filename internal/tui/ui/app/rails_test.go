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
		{name: "search", app: &state.AppModel{Mode: state.ModeSearch, LogContainerID: "abc"}, want: querySearch},
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
	app.FilterInput = state.QueryInputState{Text: "nginx", Cursor: 5}
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

func TestRenderAppRegressionMatrix(t *testing.T) {
	sizes := []struct{ width, height int }{{120, 32}, {100, 24}, {80, 20}}
	views := []struct {
		name  string
		panel state.PanelType
		mode  state.AppMode
	}{
		{name: "containers", panel: state.PanelContainers, mode: state.ModeNormal},
		{name: "images", panel: state.PanelImages, mode: state.ModeNormal},
		{name: "compose", panel: state.PanelCompose, mode: state.ModeNormal},
		{name: "logs", panel: state.PanelContainers, mode: state.ModeLogView},
		{name: "detail", panel: state.PanelImages, mode: state.ModeDetail},
		{name: "help", panel: state.PanelHelp, mode: state.ModeHelp},
		{name: "filter", panel: state.PanelContainers, mode: state.ModeFilter},
		{name: "search", panel: state.PanelContainers, mode: state.ModeSearch},
		{name: "command", panel: state.PanelContainers, mode: state.ModeCommand},
		{name: "mark", panel: state.PanelContainers, mode: state.ModeMark},
	}
	for _, size := range sizes {
		var expectedRows int
		for _, view := range views {
			t.Run(view.name, func(t *testing.T) {
				app := state.NewAppModel(config.DefaultConfig(), nil, "test")
				app.Width, app.Height = size.width, size.height
				app.ActivePanel, app.Mode = view.panel, view.mode
				app.FilterText = "query"
				app.FilterInput = state.QueryInputState{Text: "query", Cursor: 5}
				app.SearchInput = state.QueryInputState{Text: "query", Cursor: 5}
				app.LogContainerID = "container"
				if view.name != "logs" && view.name != "search" {
					app.LogContainerID = ""
				}
				app.ImageDetailContent = "Name: test"
				rendered := RenderApp(app)
				if strings.Contains(rendered, "Terminal too small") {
					t.Fatalf("supported size %dx%d degraded", size.width, size.height)
				}
				rows := strings.Count(rendered, "\n") + 1
				if expectedRows == 0 {
					expectedRows = rows
				} else if rows != expectedRows {
					t.Fatalf("layout rows changed at %dx%d: got %d want %d", size.width, size.height, rows, expectedRows)
				}
			})
		}
	}
}

func TestRenderAppRejectsUnsupportedTerminal(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Width, app.Height = 79, 19
	if got := RenderApp(app); !strings.Contains(got, "minimum 80x20") {
		t.Fatalf("unexpected degradation message: %q", got)
	}
}
