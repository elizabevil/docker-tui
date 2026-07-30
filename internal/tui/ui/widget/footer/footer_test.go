package footer

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func TestRenderAlwaysUsesThreeRows(t *testing.T) {
	apps := []*state.AppModel{
		{Navigation: state.NavigationState{Mode: state.ModeNormal, ActivePanel: state.PanelHelp}},
		{Navigation: state.NavigationState{Mode: state.ModeMark, ActivePanel: state.PanelContainers}},
		{Navigation: state.NavigationState{Mode: state.ModeDetail, ActivePanel: state.PanelImages}, Feedback: state.FeedbackState{InfoMessage: "loaded"}},
	}
	for _, app := range apps {
		got := Render(app, 120)
		if rows := strings.Count(got, "\n") + 1; rows != 2 {
			t.Fatalf("Render() rows = %d, want 2: %q", rows, got)
		}
		for _, row := range strings.Split(got, "\n") {
			if width := component.VisibleLen(row); width != 120 {
				t.Fatalf("footer row width = %d, want 120", width)
			}
		}
	}
}
