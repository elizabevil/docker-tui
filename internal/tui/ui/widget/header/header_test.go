package header

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/utils"
)

func TestRenderConstrainsLoadedConnectionDataToViewport(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width = 80
	app.Viewport.HeaderVisible = true
	app.Connection.RuntimeType = strings.Repeat("podman-", 12)
	app.Connection.EngineVersion = strings.Repeat("version-", 12)
	app.Connection.ConnectionTarget = strings.Repeat("socket-", 12)

	rendered := Render(app, 80)
	lines := strings.Split(rendered, "\n")
	if len(lines) != RenderHeight() {
		t.Fatalf("header rows = %d, want %d: %q", len(lines), RenderHeight(), rendered)
	}
	for i, line := range lines {
		if width := utils.DisplayWidth(line); width != 80 {
			t.Fatalf("header row %d width = %d, want 80: %q", i, width, line)
		}
	}
}
