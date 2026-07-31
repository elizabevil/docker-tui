package view

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestTemplateFor(t *testing.T) {
	tests := []struct {
		name string
		app  state.AppModel
		want pageTemplateKind
	}{
		{name: "list", app: state.AppModel{Navigation: state.NavigationState{ActivePanel: state.PanelImages}}, want: listPageTemplate},
		{name: "split", app: state.AppModel{Navigation: state.NavigationState{ActivePanel: state.PanelCompose}}, want: splitPageTemplate},
		{name: "help", app: state.AppModel{Navigation: state.NavigationState{ActivePanel: state.PanelHelp}}, want: helpPageTemplate},
		{name: "detail", app: state.AppModel{Navigation: state.NavigationState{Mode: state.ModeDetail}}, want: detailPageTemplate},
		{name: "logs", app: state.AppModel{Navigation: state.NavigationState{Mode: state.ModeLogView}}, want: logPageTemplate},
		{name: "log search", app: state.AppModel{Navigation: state.NavigationState{Mode: state.ModeSearch}, Log: state.LogState{LogContainerID: "container"}}, want: logPageTemplate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := templateFor(&tt.app); got != tt.want {
				t.Fatalf("templateFor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComposeDetailUsesStandardBreadcrumb(t *testing.T) {
	app := &state.AppModel{
		Navigation: state.NavigationState{
			ActivePanel: state.PanelCompose,
			Mode:        state.ModeDetail,
		},
		Viewport: state.ViewportState{Width: 120},
	}
	items := buildBreadcrumbItems(app)
	if len(items) != 2 || items[0].Label != state.PanelLabel(state.PanelCompose) || items[1].Label != "detail" {
		t.Fatalf("compose detail breadcrumb items = %#v", items)
	}
	if got := breadcrumb(app); !strings.Contains(got, state.PanelLabel(state.PanelCompose)) || !strings.Contains(got, "detail") {
		t.Fatalf("compose detail breadcrumb = %q", got)
	}
}
