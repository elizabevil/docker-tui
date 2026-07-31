package containers

import (
	"strings"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func TestRenderListFlexLayoutUsesFullWidthForStats(t *testing.T) {
	model := state.NewContainerListModel()
	model.Items = []runtimeapi.ContainerSummary{{
		ID:      "10b2c97269cc",
		Name:    "dtui-test-redis",
		Image:   "redis:7-alpine",
		State:   state.ContainerStateRunning,
		Status:  "running",
		Created: 1785388988,
	}}
	model.Stats["10b2c97269cc"] = state.ContainerStats{
		CPU: 0, MemPerc: 0, NetRx: 4.3 * 1024, NetTx: 1.1 * 1024,
	}

	const pageWidth = 200
	rendered := component.StripANSI(RenderList(model, pageWidth, 12, nil, false))
	if !strings.Contains(rendered, "RX:4.3KB TX:1.1KB") {
		t.Fatalf("stats were truncated despite available width: %q", rendered)
	}

	wantRowWidth := pageWidth
	for _, line := range strings.Split(rendered, "\n") {
		if strings.Contains(line, "dtui-test-redis") {
			if width := component.VisibleLen(line); width != wantRowWidth {
				t.Fatalf("container row width = %d, want %d: %q", width, wantRowWidth, line)
			}
			return
		}
	}
	t.Fatal("container row was not rendered")
}
