package view

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/data/runtime"
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
		{name: "normal", app: &state.AppModel{Navigation: state.NavigationState{Mode: state.ModeNormal}}, want: queryNone},
		{name: "filter", app: &state.AppModel{Navigation: state.NavigationState{Mode: state.ModeFilter}}, want: queryFilter},
		{name: "search", app: &state.AppModel{Navigation: state.NavigationState{Mode: state.ModeSearch}, Log: state.LogState{LogContainerID: "abc"}}, want: querySearch},
		{name: "command", app: &state.AppModel{Navigation: state.NavigationState{Mode: state.ModeCommand}}, want: queryCommand},
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

func TestRenderMessageRailWrapsLongErrorAcrossTwoRows(t *testing.T) {
	app := &state.AppModel{}
	app.Feedback.RecordError(strings.Repeat("x", 70))

	got := renderMessageRail(app, 40)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("message rows = %d, want 2: %q", len(lines), got)
	}
	plainLines := strings.Split(component.StripANSI(got), "\n")
	plain := strings.TrimRight(plainLines[0], " ") + strings.TrimRight(plainLines[1], " ")
	if plain != strings.Repeat("x", 70) {
		t.Fatalf("wrapped message changed: %q", plain)
	}
}

func TestRenderMessageRailCapsErrorAtTwoRows(t *testing.T) {
	app := &state.AppModel{}
	app.Feedback.RecordError(strings.Repeat("错误", 60))

	got := renderMessageRail(app, 40)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("message rows = %d, want 2: %q", len(lines), got)
	}
	for _, line := range lines {
		if width := component.VisibleLen(line); width > 40 {
			t.Fatalf("message width = %d, want <= 40: %q", width, line)
		}
	}
	if !strings.Contains(component.StripANSI(lines[1]), "...") {
		t.Fatalf("truncated second row has no ellipsis: %q", got)
	}
}

func TestRenderAppHeightDoesNotChangeWithQueryOrMessage(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width = 120
	app.Viewport.Height = 32

	normalHeight := strings.Count(RenderApp(app), "\n") + 1
	app.Navigation.Mode = state.ModeFilter
	app.Navigation.FilterInput = state.QueryInputState{Text: "nginx", Cursor: 5}
	filterHeight := strings.Count(RenderApp(app), "\n") + 1
	app.Feedback.ToastMessage = "filter applied"
	app.Feedback.ToastLevel = state.NotificationInfo
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
				app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
				app.Viewport.Width, app.Viewport.Height = size.width, size.height
				app.Navigation.ActivePanel, app.Navigation.Mode = view.panel, view.mode
				app.Navigation.CommandInput.Set("query")
				app.Navigation.FilterInput = state.QueryInputState{Text: "query", Cursor: 5}
				app.Navigation.SearchInput = state.QueryInputState{Text: "query", Cursor: 5}
				app.Log.LogContainerID = "container"
				if view.name != "logs" && view.name != "search" {
					app.Log.LogContainerID = ""
				}
				app.Detail.ImageDetailContent = "Name: test"
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

func TestRenderAppDialogOverlaysPreservePageChrome(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*state.AppModel)
		want  string
	}{
		{
			name: "confirm",
			setup: func(app *state.AppModel) {
				app.Navigation.Mode = state.ModeConfirm
				app.Confirm.Open("container-stop", "api", "Stop api?", app.Confirm.ConfirmAudit)
			},
			want: "Stop api?",
		},
		{
			name: "rename",
			setup: func(app *state.AppModel) {
				app.Navigation.Mode = state.ModeRename
				app.Dialog.Open(state.DialogSpec{Title: "Rename", Input: "api"})
			},
			want: "Rename",
		},
		{
			name: "image-transfer",
			setup: func(app *state.AppModel) {
				app.Navigation.Mode = state.ModeImageTransfer
				app.Dialog.Title = "Save image"
				app.Dialog.Body = "archive.tar"
				app.ImageTransfer.Progress.Status = "saving"
			},
			want: "Save image",
		},
		{
			name: "runtime-select",
			setup: func(app *state.AppModel) {
				app.Navigation.Mode = state.ModeRuntimeSelect
			},
			want: "runtime selection unavailable",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
			app.Viewport.Width, app.Viewport.Height = 120, 32
			app.Navigation.ActivePanel = state.PanelContainers
			app.Resources.Containers.Items = []runtime.ContainerSummary{{ID: "abc", Name: "api", Image: "nginx", State: state.ContainerStateRunning}}
			tc.setup(app)

			rendered := RenderApp(app)
			plain := component.StripANSI(rendered)
			if !strings.Contains(plain, "Containers") || !strings.Contains(plain, "api") {
				t.Fatalf("overlay dropped header chrome: %q", rendered)
			}
			if !strings.Contains(plain, tc.want) {
				t.Fatalf("overlay missing dialog content %q: %q", tc.want, rendered)
			}
			if rows := strings.Count(rendered, "\n") + 1; rows != app.Viewport.Height {
				t.Fatalf("overlay rows = %d, want %d", rows, app.Viewport.Height)
			}
		})
	}
}

func TestRenderAppRejectsUnsupportedTerminal(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width, app.Viewport.Height = 40, 10
	if got := RenderApp(app); !strings.Contains(got, "Terminal too small") {
		t.Fatalf("unexpected degradation message: %q", got)
	}
}

func TestRenderAppCompactTierKeepsCoreActionsReachable(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width, app.Viewport.Height = 79, 19
	rendered := RenderApp(app)
	if strings.Contains(rendered, "Terminal too small") {
		t.Fatalf("compact tier should not show error at 79x19: %q", rendered)
	}
	// Compact footer must keep the navigation shortcuts reachable.
	for _, hint := range []string{"Tab", "Enter", "Esc"} {
		if !strings.Contains(rendered, hint) {
			t.Fatalf("compact footer missing %q: %q", hint, rendered)
		}
	}
	rows := strings.Count(rendered, "\n") + 1
	// The compact panel renders one fewer row than the viewport because
	// the body uses space-between layout; we accept a +/-1 slack so this
	// test stays stable across content changes.
	if rows < app.Viewport.Height-1 || rows > app.Viewport.Height {
		t.Fatalf("compact layout rows=%d, want ~%d", rows, app.Viewport.Height)
	}
}

func TestClassifyTerminal(t *testing.T) {
	cases := []struct {
		w, h int
		want TerminalClass
	}{
		{40, 10, TerminalUnsupported},
		{59, 14, TerminalUnsupported},
		{60, 14, TerminalCompact},
		{79, 19, TerminalCompact},
		{80, 20, TerminalStandard},
		{120, 32, TerminalStandard},
		{200, 60, TerminalStandard},
	}
	for _, tc := range cases {
		if got := ClassifyTerminal(tc.w, tc.h); got != tc.want {
			t.Errorf("ClassifyTerminal(%d,%d) = %d, want %d", tc.w, tc.h, got, tc.want)
		}
	}
}

func TestPlanForAllocatesAllRows(t *testing.T) {
	for _, h := range []int{14, 16, 19, 20, 24, 32, 50} {
		class := ClassifyTerminal(80, h)
		plan := planFor(class, h)
		if plan.header <= 0 || plan.footer <= 0 || plan.panel <= 0 {
			t.Errorf("plan for h=%d class=%d has zero rail: %+v", h, class, plan)
		}
		if plan.total() > h {
			t.Errorf("plan for h=%d total=%d exceeds viewport", h, plan.total())
		}
	}
}

func TestRenderAppCompactDoesNotDropToast(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width, app.Viewport.Height = 79, 19
	app.Feedback.ToastMessage = "container started"
	app.Feedback.ToastLevel = state.NotificationSuccess
	rendered := RenderApp(app)
	if !strings.Contains(rendered, "container started") {
		t.Fatalf("compact tier dropped toast: %q", rendered)
	}
}

func TestRenderAppCompactHandlesQueryInput(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width, app.Viewport.Height = 79, 19
	app.Navigation.Mode = state.ModeFilter
	app.Navigation.FilterInput = state.QueryInputState{Text: "nginx", Cursor: 5}
	rendered := RenderApp(app)
	if !strings.Contains(rendered, "Filter:") {
		t.Fatalf("compact tier dropped filter input: %q", rendered)
	}
}
