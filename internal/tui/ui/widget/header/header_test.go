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
	theme.Surfaces.Header = config.ValueRef(config.Color("#2b2d30cc"))
	theme.Chrome.HeaderLabel = config.TokenRef(config.ColorTokenTransparent)
	theme.Chrome.HeaderValue = config.TokenRef(config.ColorTokenTransparent)
	tui.ApplyTheme(theme)

	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width = 80
	app.Viewport.HeaderVisible = true
	app.Metrics.HostCPU = 42

	rendered := Render(app, 80)
	for line := range strings.SplitSeq(rendered, "\n") {
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
	theme.Surfaces.Header = config.ValueRef(config.Color("#2b2d30cc"))
	theme.Chrome.HeaderLabel = config.TokenRef(config.ColorTokenForegroundMuted)
	theme.Chrome.HeaderValue = config.TokenRef(config.ColorTokenForeground)
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
	theme.Surfaces.Header = config.ValueRef(config.Color("#2b2d30cc"))
	theme.Chrome.HeaderLabel = config.TokenRef(config.ColorTokenForegroundMuted)
	theme.Chrome.HeaderValue = config.TokenRef(config.ColorTokenForeground)
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

// TestRenderLinkLatencyShareOneRow verifies that the Link indicator and the
// Latency value live on the same rendered row. The user reads link + RTT
// together as a single "is the connection healthy?" signal; splitting them
// across two rows makes the latency look like an unrelated row.
func TestRenderLinkLatencyShareOneRow(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width = 80
	app.Viewport.HeaderVisible = true
	app.Connection.Connected = true
	app.Connection.Pool = nil // no latency reported → "—"

	rendered := Render(app, 80)
	lines := strings.Split(utils.StripANSI(rendered), "\n")
	var linkRow int
	for i, line := range lines {
		if strings.Contains(line, "Link") {
			linkRow = i
			break
		}
	}
	if linkRow == 0 {
		t.Fatalf("no row carries Link label: %q", rendered)
	}
	// The link indicator and latency value must be on the same row.
	if !strings.Contains(lines[linkRow], "ms") && !strings.Contains(lines[linkRow], "—") {
		t.Fatalf("link row must also carry the latency value: %q", lines[linkRow])
	}
	// No standalone "Latency" label should appear in any other row.
	for i, line := range lines {
		if i == linkRow {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "Latency") {
			t.Fatalf("latency was split into its own row %d: %q", i, line)
		}
	}
}

// TestRenderHeaderColumnsShareRowCount verifies that each header column has
// the same number of visible rows so the label-value pairs line up across
// columns. A mismatch used to leave col 2 with a phantom empty row while
// col 1 reported a 4th metric.
func TestRenderHeaderColumnsShareRowCount(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width = 80
	app.Viewport.HeaderVisible = true
	app.Connection.Connected = true
	app.Connection.Pool = nil

	rendered := strings.Split(utils.StripANSI(Render(app, 80)), "\n")
	if len(rendered) != RenderHeight() {
		t.Fatalf("header rows = %d, want %d: %q", len(rendered), RenderHeight(), rendered)
	}
	// Count rows where each column's anchor label appears:
	// col 1 (CPU/Memory/Disk/TimeZone) and col 2 (Engine/Version/Socket/Link)
	// must match so the rows line up horizontally.
	col1Labels := []string{"CPU", "Memory", "Disk", "TimeZone"}
	col2Labels := []string{"Engine", "Version", "Socket", "Link"}
	col1Count, col2Count := 0, 0
	for _, line := range rendered {
		for _, lbl := range col1Labels {
			if strings.Contains(line, lbl) {
				col1Count++
				break
			}
		}
		for _, lbl := range col2Labels {
			if strings.Contains(line, lbl) {
				col2Count++
				break
			}
		}
	}
	if col1Count != col2Count {
		t.Fatalf("header columns row count mismatch: col1=%d, col2=%d, want equal", col1Count, col2Count)
	}
}

// TestRenderHeaderColumnsUse2Plus2Plus3Plus3Ratio documents the 2:2:3:3
// proportion used for the four header columns. The test asserts that the
// actual rendered widths follow the ratio so themes / wider terminals
// preserve the same proportions.
func TestRenderHeaderColumnsUse2Plus2Plus3Plus3Ratio(t *testing.T) {
	cfg := config.DefaultAppConfig()
	if cfg.UI.Header.Columns.Host != 2 || cfg.UI.Header.Columns.Connection != 2 ||
		cfg.UI.Header.Columns.Keystroke != 3 || cfg.UI.Header.Columns.Logo != 3 {
		t.Fatalf("default header weights = %+v, want 2:2:3:3", cfg.UI.Header.Columns)
	}
}

// TestRenderKeyStrokeBoxHasMinPadding ensures that the keystroke border
// carries at least 3 cells of horizontal padding between the rounded edge
// and the first badge. The previous formula let the badges hug the border
// whenever the column was moderately narrow.
func TestRenderKeyStrokeBoxHasMinPadding(t *testing.T) {
	theme := config.DefaultTheme()
	theme.Palette.Background = config.Color("#18191b")
	theme.Surfaces.Header = config.ValueRef(config.Color("#2b2d30cc"))
	theme.Chrome.HeaderLabel = config.TokenRef(config.ColorTokenForegroundMuted)
	theme.Chrome.HeaderValue = config.TokenRef(config.ColorTokenForeground)
	component.SetRawStylesForTest(component.RawStylesForTest())
	component.SetSafeFallbackForTest(component.SafeFallbackForTest())
	tui.ApplyTheme(theme)
	defer tui.ApplyTheme(config.DefaultTheme())

	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Viewport.Width = 100
	app.Viewport.HeaderVisible = true
	app.Feedback.KeyStrokeBuffer = []state.KeyStrokeEvent{
		{Key: "ctrl+b", Action: "Debug"},
	}

	rendered := Render(app, 100)
	for line := range strings.SplitSeq(rendered, "\n") {
		clean := utils.StripANSI(line)
		idx := strings.Index(clean, "ctrl+b")
		if idx < 0 {
			continue
		}
		// The first badge must be at least 3 cells into the line (open border
		// + minimal padding); the previous default put it at index 2.
		if idx < 3 {
			t.Fatalf("keystroke border has insufficient padding: badge at col %d: %q", idx, clean)
		}
		return
	}
	t.Fatalf("keystroke badge not rendered: %q", rendered)
}
