package tui

import (
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"

	"charm.land/lipgloss/v2"
)

var (
	ActiveBorderStyle   lipgloss.Style
	InactiveBorderStyle lipgloss.Style
	FocusedBorderStyle  lipgloss.Style
)

func ApplyTheme(theme *config.Theme) {
	c := theme.Palette
	style.Colors.Success = style.Color(string(c.Success))
	style.Colors.Primary = style.Color(string(c.Primary))
	style.Colors.Info = style.Color(string(c.Info))
	style.Colors.Danger = style.Color(string(c.Danger))
	style.Colors.Warning = style.Color(string(c.Warning))
	style.Colors.Accent = style.Color(string(c.Accent))
	style.Colors.AccentSecondary = style.Color(string(c.AccentSecondary))
	style.Colors.Foreground = style.Color(string(c.Foreground))
	style.Colors.ForegroundMuted = style.Color(string(c.ForegroundMuted))
	style.Colors.BackgroundSubtle = style.Color(string(c.BackgroundSubtle))
	style.Colors.BackgroundDeep = style.Color(string(c.BackgroundDeep))
	style.Colors.BG = style.Color(string(c.Background))
	style.SyncPalette()
	component.ApplyThemeStyles(theme)

	br := component.ResolveBorder(theme.Border.Kind)
	panelBackground := style.Colors.BG
	ActiveBorderStyle = lipgloss.NewStyle().Border(br).Foreground(style.Color(theme.ResolveColor(theme.Main.BorderActive))).Background(panelBackground).Padding(0)
	InactiveBorderStyle = lipgloss.NewStyle().Foreground(style.Color(theme.ResolveColor(theme.Main.BorderInactive))).Background(panelBackground).Padding(0)
	FocusedBorderStyle = lipgloss.NewStyle().Border(br).Foreground(style.Color(theme.ResolveColor(theme.Border.Focused))).Background(panelBackground).Padding(0)
}

// ApplyLayoutConfig clears border backgrounds when image/fallthrough backgrounds
// are enabled. Per-component background clearing is handled by layout.go's image
// layer system (renderContentLayer), so only the 3 border globals need adjustment.
func ApplyLayoutConfig(layout *config.LayoutConfig) {
	if layout == nil {
		return
	}
	bg := layout.Background
	if !bg.Enable {
		return
	}

	if bg.Fallthrough || (bg.Type == config.BackgroundImageType && bg.Image.Src != "") {
		ActiveBorderStyle = ActiveBorderStyle.Background(lipgloss.NoColor{})
	}
}

func ApplyUIConfig(ui *config.UIConfig) {
	if ui == nil {
		return
	}
	component.ApplyTableLayout(ui.Table)
}

func init() {
	ApplyTheme(config.DefaultTheme())
}
