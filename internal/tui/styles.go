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
	c := theme.Colors
	style.Colors.Green = lipgloss.Color(c.Green)
	style.Colors.Cyan = lipgloss.Color(c.Cyan)
	style.Colors.Blue = lipgloss.Color(c.Blue)
	style.Colors.Red = lipgloss.Color(c.Red)
	style.Colors.Yellow = lipgloss.Color(c.Yellow)
	style.Colors.Orange = lipgloss.Color(c.Orange)
	style.Colors.Purple = lipgloss.Color(c.Purple)
	style.Colors.White = lipgloss.Color(c.White)
	style.Colors.Gray = lipgloss.Color(c.Gray)
	style.Colors.Dark = lipgloss.Color(c.Dark)
	style.Colors.Surface = lipgloss.Color(c.Surface)
	style.Colors.BG = lipgloss.Color(c.Background)

	br := component.ResolveBorder(theme.Border.Style)
	ActiveBorderStyle = lipgloss.NewStyle().Border(br).Foreground(style.Color(theme.Main.BorderActive)).Padding(0)
	InactiveBorderStyle = lipgloss.NewStyle().Foreground(style.Color(theme.Main.BorderInactive)).Padding(0)
	FocusedBorderStyle = lipgloss.NewStyle().Border(br).Foreground(style.Color(theme.Border.Focused)).Padding(0)
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

	if bg.Fallthrough || (bg.Type == "image" && bg.Image.Src != "") {
		ActiveBorderStyle = ActiveBorderStyle.Background(lipgloss.NoColor{})
	}
}

func init() {
	ApplyTheme(config.DefaultTheme())
}
