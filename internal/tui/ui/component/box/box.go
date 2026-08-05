// Package box provides self-contained, component-level rendering primitives.
//
// Each component manages its own background as a property of the box itself
// rather than relying on character-level ANSI reset barriers. The "transparent"
// semantic is: a component with no background color let the terminal default
// show through; it does not rely on a parent wrapper to provide the fill.
package box

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// StyleName 是 component 包样式缓存的类型化标识(类型别名),
// 直接对接 component.GetStyle 的入参,box 组件传入的样式名
// 与全局样式注册表在编译期保持一致。
type StyleName = component.StyleName

// 常用样式名。未注册的保留名(rowNormal/marked/alt)经
// component.GetStyle 解析时回退到 safe fallback。
const (
	StyleHeaderLabel = component.StyleHeaderLabel
	StyleHeaderValue = component.StyleHeaderBar
	StyleHeaderKey   = component.StyleKeyBadge
	StyleHeaderLast  = component.StyleKeyLast
	StylePanelTitle  = component.StylePanelTitle
	StyleDim         = component.StyleDim
	StyleRowNormal   = component.StyleName("rowNormal") // 保留;未注册 → safe fallback
	StyleRowSelected = component.StyleSelectedRow
	StyleRowMarked   = component.StyleName("marked") // 保留;未注册 → safe fallback
	StyleRowAlt      = component.StyleName("alt")    // 保留;未注册 → safe fallback
)

// backgroundColor returns the parsed color or nil when bg is empty.
func backgroundColor(bg string) color.Color {
	if bg == "" {
		return nil
	}
	if c, ok := utils.ParseColor(bg); ok {
		return c
	}
	return nil
}

// styleWithBackground returns the style with an optional background applied.
// Centralised so components never inject background codes at the character
// level — the background is part of the box's lipgloss.Style.
func styleWithBackground(base lipgloss.Style, bg string) lipgloss.Style {
	if bg == "" {
		return base
	}
	if c := backgroundColor(bg); c != nil {
		return base.Background(c)
	}
	return base
}
