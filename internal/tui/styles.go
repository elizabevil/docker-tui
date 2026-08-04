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
	style.Colors.Green = style.Color(string(c.Green))
	style.Colors.Cyan = style.Color(string(c.Cyan))
	style.Colors.Blue = style.Color(string(c.Blue))
	style.Colors.Red = style.Color(string(c.Red))
	style.Colors.Yellow = style.Color(string(c.Yellow))
	style.Colors.Orange = style.Color(string(c.Orange))
	style.Colors.Purple = style.Color(string(c.Purple))
	style.Colors.White = style.Color(string(c.White))
	style.Colors.Gray = style.Color(string(c.Gray))
	style.Colors.Dark = style.Color(string(c.Dark))
	style.Colors.Surface = style.Color(string(c.Surface))
	style.Colors.BG = style.Color(string(c.Background))
	style.SyncPalette()
	component.ApplyThemeStyles(theme)

	br := component.ResolveBorder(theme.Border.Kind)
	ActiveBorderStyle = lipgloss.NewStyle().Border(br).Foreground(style.Color(theme.ResolveColor(theme.Main.BorderActive))).Padding(0)
	InactiveBorderStyle = lipgloss.NewStyle().Foreground(style.Color(theme.ResolveColor(theme.Main.BorderInactive))).Padding(0)
	FocusedBorderStyle = lipgloss.NewStyle().Border(br).Foreground(style.Color(theme.ResolveColor(theme.Border.Focused))).Padding(0)
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
