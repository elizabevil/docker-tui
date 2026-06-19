// Package style is the single source of truth for all resolved lipgloss styles.
// It is populated at startup by tui.ApplyTheme() and read by all UI components.
package style

import (
	"image/color"

	"charm.land/lipgloss/v2"
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

// Color 按调色板名返回颜色，未知名返回白色兜底。
func Color(name string) color.Color {
	switch name {
	case "green":
		return Colors.Green
	case "cyan":
		return Colors.Cyan
	case "blue":
		return Colors.Blue
	case "red":
		return Colors.Red
	case "yellow":
		return Colors.Yellow
	case "orange":
		return Colors.Orange
	case "purple":
		return Colors.Purple
	case "white":
		return Colors.White
	case "gray":
		return Colors.Gray
	case "dark":
		return Colors.Dark
	case "surface":
		return Colors.Surface
	case "bg":
		return Colors.BG
	default:
		return Colors.White
	}
}
