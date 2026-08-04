package header

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"
)

func TestRenderHeaderRespectsTransparentLabelValueTokens(t *testing.T) {
	previousStyles := component.RawStylesForTest()
	previousSafe := component.SafeFallbackForTest()
	defer func() {
		component.SetRawStylesForTest(previousStyles)
		component.SetSafeFallbackForTest(previousSafe)
	}()

	theme := config.DefaultTheme()
	theme.Header.Background = config.ValueRef(config.Color("#2b2d30cc"))
	theme.Header.Label = config.TokenRef(config.ColorTokenTransparent)
	theme.Header.Value = config.TokenRef(config.ColorTokenTransparent)
	tui.ApplyTheme(theme)

	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width = 80
	app.Viewport.HeaderVisible = true
	app.Metrics.HostCPU = 42

	rendered := Render(app, 80)
	for _, line := range strings.Split(rendered, "\n") {
		plain := utils.StripANSI(line)
		idx := strings.Index(plain, "TimeZone")
		if idx < 0 {
			continue
		}
		post := plain[idx+len("TimeZone"):]
		if strings.TrimSpace(post) == "" {
			t.Fatalf("value cell rendered empty after transparent label: %q", plain)
		}
		tail := line[idx+len("TimeZone"):]
		if strings.Contains(tail, "48;2;39;41;44m") {
			t.Fatalf("transparent value still paints header background: %q", tail)
		}
	}
}

func TestRenderKeyStrokeInheritsHeaderBackground(t *testing.T) {
	previousStyles := component.RawStylesForTest()
	previousSafe := component.SafeFallbackForTest()
	defer func() {
		component.SetRawStylesForTest(previousStyles)
		component.SetSafeFallbackForTest(previousSafe)
	}()

	theme := config.DefaultTheme()
	theme.Palette.Background = config.Color("#18191b")
	theme.Header.Background = config.ValueRef(config.Color("#2b2d30cc"))
	theme.Header.Label = config.TokenRef(config.ColorTokenForegroundMuted)
	theme.Header.Value = config.TokenRef(config.ColorTokenForeground)
	tui.ApplyTheme(theme)

	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width = 80
	app.Viewport.HeaderVisible = true
	app.Feedback.KeyStrokeBuffer = []state.KeyStrokeEvent{{Key: "R", Action: "Restart"}}
	app.Metrics.HostCPU = 42

	rendered := Render(app, 80)
	if !strings.Contains(rendered, "48;2;39;41;44m") {
		t.Fatalf("header output does not include blended background: %q", rendered)
	}
}

func TestRenderCPUUsesHeaderBackground(t *testing.T) {
	theme := config.DefaultTheme()
	theme.Palette.Background = config.Color("#18191b")
	theme.Header.Background = config.ValueRef(config.Color("#2b2d30cc"))
	theme.Header.Label = config.TokenRef(config.ColorTokenForegroundMuted)
	theme.Header.Value = config.TokenRef(config.ColorTokenForeground)
	component.SetRawStylesForTest(component.RawStylesForTest())
	component.SetSafeFallbackForTest(component.SafeFallbackForTest())
	tui.ApplyTheme(theme)
	defer tui.ApplyTheme(config.DefaultTheme())

	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width = 80
	app.Viewport.HeaderVisible = true
	rendered := Render(app, 80)
	if !strings.Contains(rendered, "48;2;39;41;44m") {
		t.Fatalf("header CPU area does not use blended header background: %q", rendered)
	}
}

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
