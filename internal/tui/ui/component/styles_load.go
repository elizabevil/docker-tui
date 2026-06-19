package component

import (
	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
)

type styleConfig struct {
	Styles map[string]styleRef `json:"styles"`
}

// styleRef 定义 JSONC 中单一样式条目的属性。
// 字段与 styles.jsonc 中的 "styles" 对象一一对应。
type styleRef struct {
	Color      string `json:"color,omitempty"`
	Background string `json:"background,omitempty"`
	Bold       bool   `json:"bold,omitempty"`
	Faint      bool   `json:"faint,omitempty"`
}

// rawStyles 存储从 styles.jsonc 加载的原始样式引用。
// 延迟解析以避开 init() 顺序依赖（style.Colors 在 ApplyTheme 中填充）。
var rawStyles map[string]styleRef

// componentStyles 存储按组件注册的私有样式覆盖。
// 通过 RegisterComponentStyles() 在组件 init() 中填充。
var componentStyles map[string]map[string]styleRef

func loadComponentStyles() {
	data, err := embeddedConfig.ReadFile("styles.jsonc")
	if err != nil {
		return
	}
	loader := ConfigLoader[styleConfig]{
		RawData:   data,
		Fallback:  styleConfig{},
		Normalize: nil,
	}
	cfg := loader.Load()
	rawStyles = cfg.Styles
}

// SafeFallbackStyles 定义 5 个安全兜底样式，确保 GetStyle 永不返回空值。
type SafeFallbackStyles struct {
	Normal lipgloss.Style
	Bold   lipgloss.Style
	Dim    lipgloss.Style
	Accent lipgloss.Style
	Error  lipgloss.Style
}

// SafeFallback 提供 5 个内置安全默认样式，作为 getStyleChain 的最后兜底。
// 灵感来自 Compose 的 staticCompositionLocalOf { default } 模式。
var SafeFallback = SafeFallbackStyles{
	Normal: lipgloss.NewStyle().Foreground(lipgloss.Color("#c9d1d9")),
	Bold:   lipgloss.NewStyle().Foreground(lipgloss.Color("#c9d1d9")).Bold(true),
	Dim:    lipgloss.NewStyle().Foreground(lipgloss.Color("#5a6270")).Faint(true),
	Accent: lipgloss.NewStyle().Foreground(lipgloss.Color("#56b4c2")),
	Error:  lipgloss.NewStyle().Foreground(lipgloss.Color("#db5a5a")),
}

// RegisterComponentStyles 注册组件的私有样式覆盖。
// 在组件的 init() 中调用，例如：
//
//	func init() {
//	    RegisterComponentStyles("filter", map[string]styleRef{
//	        "searchCursor": {Color: "white", Bold: true},
//	    })
//	}
//
// 组件样式在 getStyleChain 中优先于全局 styles.jsonc 查找。
func RegisterComponentStyles(component string, refs map[string]styleRef) {
	if componentStyles == nil {
		componentStyles = make(map[string]map[string]styleRef)
	}
	componentStyles[component] = refs
}

// GetStyle 按名称解析样式，颜色从当前调色板解析。
// 使用 getStyleChain 进行多层查找。
func GetStyle(name string) lipgloss.Style {
	return getStyleChain(name, "")
}

// GetComponentStyle 优先从组件私有样式查找，然后回退到全局 styles.jsonc、
// 调色板名、最后 SafeFallback。
func GetComponentStyle(component, name string) lipgloss.Style {
	return getStyleChain(name, component)
}

// getStyleChain 实现多层样式查找链：
//
//	第 2 层：组件私有（RegisterComponentStyles 注册）
//	第 1 层：全局 styles.jsonc
//	第 0 层：调色板颜色名（如 "green"）
//	兜底：SafeFallback.Normal（始终可见，永不空）
func getStyleChain(name, component string) lipgloss.Style {
	// 第 2 层：组件私有
	if component != "" && componentStyles != nil {
		if refs, ok := componentStyles[component]; ok {
			if ref, ok := refs[name]; ok {
				return buildStyle(ref)
			}
		}
	}
	// 第 1 层：全局 styles.jsonc
	if ref, ok := rawStyles[name]; ok {
		return buildStyle(ref)
	}
	// 第 0 层：调色板颜色名
	if c := style.Color(name); c != nil {
		return lipgloss.NewStyle().Foreground(c)
	}
	// 兜底：SafeFallback.Normal
	return SafeFallback.Normal
}

// buildStyle 将 styleRef 转换为 lipgloss.Style，通过当前调色板解析颜色名。
// 由 getStyleChain 在第 2 层和第 1 层调用。
func buildStyle(ref styleRef) lipgloss.Style {
	s := lipgloss.NewStyle()
	if ref.Color != "" {
		if ref.Color[0] == '#' {
			s = s.Foreground(lipgloss.Color(ref.Color))
		} else if c := style.Color(ref.Color); c != nil {
			s = s.Foreground(c)
		}
	}
	if ref.Background != "" {
		if ref.Background[0] == '#' {
			s = s.Background(lipgloss.Color(ref.Background))
		} else if c := style.Color(ref.Background); c != nil {
			s = s.Background(c)
		}
	}
	if ref.Bold {
		s = s.Bold(true)
	}
	if ref.Faint {
		s = s.Faint(true)
	}
	return s
}
