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
	Success          color.Color
	Primary          color.Color
	Info             color.Color
	Danger           color.Color
	Warning          color.Color
	Accent           color.Color
	AccentSecondary  color.Color
	Foreground       color.Color
	ForegroundMuted  color.Color
	BackgroundSubtle color.Color
	BackgroundDeep   color.Color
	BG               color.Color
}

// Colors 是全局唯一的调色板实例。
var Colors = Palette{
	Success:          lipgloss.Color("#499c54"),
	Primary:          lipgloss.Color("#56b4c2"),
	Info:             lipgloss.Color("#589df6"),
	Danger:           lipgloss.Color("#db5a5a"),
	Warning:          lipgloss.Color("#c8a35e"),
	Accent:           lipgloss.Color("#cc7832"),
	AccentSecondary:  lipgloss.Color("#a962b5"),
	Foreground:       lipgloss.Color("#c9d1d9"),
	ForegroundMuted:  lipgloss.Color("#5a6270"),
	BackgroundSubtle: lipgloss.Color("#1e1f22"),
	BackgroundDeep:   lipgloss.Color("#2b2d30"),
	BG:               lipgloss.Color("#18191b"),
}

// init 把编译期默认调色板注册到 utils，使 utils.ParseColor 在
// tui.ApplyTheme() 运行之前即可解析调色板色名。
func init() {
	SyncPalette()
}

// SyncPalette 将当前 Colors 调色板注册到 utils，供 ParseColor 解析
// 调色板色名（如 constants.ColorSuccess）。主题应用后需重新调用以同步新值。
func SyncPalette() {
	utils.SetPaletteColor(constants.ColorSuccess, Colors.Success)
	utils.SetPaletteColor(constants.ColorPrimary, Colors.Primary)
	utils.SetPaletteColor(constants.ColorInfo, Colors.Info)
	utils.SetPaletteColor(constants.ColorDanger, Colors.Danger)
	utils.SetPaletteColor(constants.ColorWarning, Colors.Warning)
	utils.SetPaletteColor(constants.ColorAccent, Colors.Accent)
	utils.SetPaletteColor(constants.ColorAccentSecondary, Colors.AccentSecondary)
	utils.SetPaletteColor(constants.ColorForeground, Colors.Foreground)
	utils.SetPaletteColor(constants.ColorForegroundMuted, Colors.ForegroundMuted)
	utils.SetPaletteColor(constants.ColorForegroundMuted, Colors.ForegroundMuted)
	utils.SetPaletteColor(constants.ColorBackgroundSubtle, Colors.BackgroundSubtle)
	utils.SetPaletteColor(constants.ColorBackgroundDeep, Colors.BackgroundDeep)
	utils.SetPaletteColor(constants.ColorBG, Colors.BG)
}

// Color resolves a configured color and preserves the historical foreground
// fallback for unknown values.
func Color(value string) color.Color {
	if resolved, ok := utils.ParseColor(value); ok {
		return resolved
	}
	return Colors.Foreground
}

// ApplyForeground returns s with c set as the foreground color, unless c is
// nil or fully transparent (α == 0). In those cases the foreground channel
// stays unset so lipgloss does not render a hard-coded black on top of
// whatever the user wanted to be transparent (Phase 0 evidence in
// test/diagnostics/lipgloss-transparent-probe). Mirrors the three-state
// dispatch in component.buildStyle.
func ApplyForeground(s lipgloss.Style, c color.Color) lipgloss.Style {
	if c == nil {
		return s
	}
	nrgba, ok := color.NRGBAModel.Convert(c).(color.NRGBA)
	if ok && nrgba.A == 0 {
		return s
	}
	return s.Foreground(c)
}

// ApplyBackground is the background counterpart of ApplyForeground.
func ApplyBackground(s lipgloss.Style, c color.Color) lipgloss.Style {
	if c == nil {
		return s
	}
	nrgba, ok := color.NRGBAModel.Convert(c).(color.NRGBA)
	if ok && nrgba.A == 0 {
		return s
	}
	return s.Background(c)
}
