package networks

import (
	"fmt"
	"strings"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func TestRenderListFlexLayoutUsesAvailableWidth(t *testing.T) {
	model := state.NewNetworkListModel()
	model.Items = []runtimeapi.Network{{
		Name:       "integration-backend",
		ID:         "1234567890abcdef",
		Driver:     "bridge",
		Scope:      "local",
		IPAM:       []string{"172.28.0.0/16"},
		Containers: 3,
		Created:    1785388988,
	}}

	for _, pageWidth := range []int{70, 120} {
		rendered := component.StripANSI(RenderList(model, pageWidth, 12, nil, false))
		wantRowWidth := pageWidth
		found := false
		for _, line := range strings.Split(rendered, "\n") {
			if strings.Contains(line, "integration-backend") {
				found = true
				if got := component.VisibleLen(line); got != wantRowWidth {
					t.Fatalf("page width %d: network row width = %d, want %d: %q", pageWidth, got, wantRowWidth, line)
				}
				break
			}
		}
		if !found {
			t.Fatalf("page width %d: network row was not rendered: %q", pageWidth, rendered)
		}
		if !strings.Contains(rendered, "1234567890ab") || strings.Contains(rendered, "1234567890abc") {
			t.Fatalf("page width %d: network short ID is wrong: %q", pageWidth, rendered)
		}
	}
}

func TestRenderListKeepsNetworkCursorInsideVisibleRows(t *testing.T) {
	model := state.NewNetworkListModel()
	model.Items = make([]runtimeapi.Network, 30)
	for i := range model.Items {
		model.Items[i] = runtimeapi.Network{
			Name:   fmt.Sprintf("network-%02d", i),
			ID:     fmt.Sprintf("%064d", i),
			Driver: "bridge",
			Scope:  "local",
		}
	}
	model.Cursor = 20

	rendered := component.StripANSI(RenderList(model, 120, 12, nil, false))
	rowHeight := component.CalcTableRowHeight(12, true)
	wantOffset := model.Cursor - rowHeight + 1
	if model.ViewOffset != wantOffset {
		t.Fatalf("ViewOffset = %d, want %d", model.ViewOffset, wantOffset)
	}
	if !strings.Contains(rendered, "network-20") {
		t.Fatalf("selected network is outside rendered table: %q", rendered)
	}
}
