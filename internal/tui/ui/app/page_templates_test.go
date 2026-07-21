package view

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestTemplateFor(t *testing.T) {
	tests := []struct {
		name string
		app  state.AppModel
		want pageTemplateKind
	}{
		{name: "list", app: state.AppModel{NavigationState: state.NavigationState{ActivePanel: state.PanelImages}}, want: listPageTemplate},
		{name: "split", app: state.AppModel{NavigationState: state.NavigationState{ActivePanel: state.PanelCompose}}, want: splitPageTemplate},
		{name: "help", app: state.AppModel{NavigationState: state.NavigationState{ActivePanel: state.PanelHelp}}, want: helpPageTemplate},
		{name: "detail", app: state.AppModel{NavigationState: state.NavigationState{Mode: state.ModeDetail}}, want: detailPageTemplate},
		{name: "logs", app: state.AppModel{NavigationState: state.NavigationState{Mode: state.ModeLogView}}, want: logPageTemplate},
		{name: "log search", app: state.AppModel{NavigationState: state.NavigationState{Mode: state.ModeSearch}, LogState: state.LogState{LogContainerID: "container"}}, want: logPageTemplate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := templateFor(&tt.app); got != tt.want {
				t.Fatalf("templateFor() = %q, want %q", got, tt.want)
			}
		})
	}
}
