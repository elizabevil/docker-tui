package detail

import (
	"fmt"
	"strings"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func TestRenderViewClampsImageDetailOverscroll(t *testing.T) {
	environment := make([]string, 30)
	for i := range environment {
		environment[i] = fmt.Sprintf("KEY_%02d=value", i)
	}
	app := &state.AppModel{}
	app.Detail.OpenImage("image", "Image Detail", &runtimeapi.ImageDetail{
		ID: "image",
		Runtime: runtimeapi.ImageRuntimeConfig{
			Environment: environment,
		},
	})
	app.Detail.DetailOffset = 1000

	RenderView(app, 8)
	bottom := app.Detail.DetailOffset
	if bottom <= 0 || bottom >= 1000 {
		t.Fatalf("render did not clamp overscroll: %d", bottom)
	}

	app.Detail.Scroll(-1)
	RenderView(app, 8)
	if app.Detail.DetailOffset != bottom-1 {
		t.Fatalf("up scroll remained stuck: got %d, want %d", app.Detail.DetailOffset, bottom-1)
	}
}

func TestRenderSectionsFooterUsesActualVisibleCount(t *testing.T) {
	app := &state.AppModel{}
	rendered := renderSections(app, []detailSection{{
		Title: "Basic",
		Lines: []string{"ID: image"},
	}}, 8)

	if plain := component.StripANSI(rendered); !strings.Contains(plain, "1-2/2") {
		t.Fatalf("footer does not use actual visible count: %q", plain)
	}
}
