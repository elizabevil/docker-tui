// Package style is the single source of truth for all resolved lipgloss styles.
// It is populated at startup by tui.ApplyTheme() and read by all UI components.
package style

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/constants"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// ── Palette ────────────────────────────────────────────────────

// Palette 定义 12 色调色板，在启动时由 tui.ApplyTheme() 填充，
// 并供 component.GetStyle()/buildStyle() 读取以完成样式解析。
// 编译期默认值与 "default" 主题一致。
type Palette struct {
	Green   color.Color
	Cyan    color.Color
	Blue    color.Color
	Red     color.Color
	Yellow  color.Color
	Orange  color.Color
	Purple  color.Color
	White   color.Color
	Gray    color.Color
	Dark    color.Color
	Surface color.Color
	BG      color.Color
}

// Colors 是全局唯一的调色板实例。
var Colors = Palette{
	Green:   lipgloss.Color("#499c54"),
	Cyan:    lipgloss.Color("#56b4c2"),
	Blue:    lipgloss.Color("#589df6"),
	Red:     lipgloss.Color("#db5a5a"),
	Yellow:  lipgloss.Color("#c8a35e"),
	Orange:  lipgloss.Color("#cc7832"),
	Purple:  lipgloss.Color("#a962b5"),
	White:   lipgloss.Color("#c9d1d9"),
	Gray:    lipgloss.Color("#5a6270"),
	Dark:    lipgloss.Color("#1e1f22"),
	Surface: lipgloss.Color("#2b2d30"),
	BG:      lipgloss.Color("#18191b"),
}

// init 把编译期默认调色板注册到 utils，使 utils.ParseColor 在
// tui.ApplyTheme() 运行之前即可解析调色板色名。
func init() {
	SyncPalette()
}

// SyncPalette 将当前 Colors 调色板注册到 utils，供 ParseColor 解析
// 调色板色名（如 constants.ColorGreen）。主题应用后需重新调用以同步新值。
func SyncPalette() {
	utils.SetPaletteColor(constants.ColorGreen, Colors.Green)
	utils.SetPaletteColor(constants.ColorCyan, Colors.Cyan)
	utils.SetPaletteColor(constants.ColorBlue, Colors.Blue)
	utils.SetPaletteColor(constants.ColorRed, Colors.Red)
	utils.SetPaletteColor(constants.ColorYellow, Colors.Yellow)
	utils.SetPaletteColor(constants.ColorOrange, Colors.Orange)
	utils.SetPaletteColor(constants.ColorPurple, Colors.Purple)
	utils.SetPaletteColor(constants.ColorWhite, Colors.White)
	utils.SetPaletteColor(constants.ColorGray, Colors.Gray)
	utils.SetPaletteColor(constants.ColorGrey, Colors.Gray)
	utils.SetPaletteColor(constants.ColorDark, Colors.Dark)
	utils.SetPaletteColor(constants.ColorSurface, Colors.Surface)
	utils.SetPaletteColor(constants.ColorBG, Colors.BG)
}

// Color resolves a configured color and preserves the historical white
// fallback for unknown values.
func Color(value string) color.Color {
	if resolved, ok := utils.ParseColor(value); ok {
		return resolved
	}
	return Colors.White
}
