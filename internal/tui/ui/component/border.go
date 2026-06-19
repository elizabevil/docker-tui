package component

import (
	_ "embed"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
)

//go:embed borders.jsonc
var bordersData []byte

var (
	bordersOnce  sync.Once
	bordersCache bordersConfig
)

// BorderDef defines the 8 characters that make up a complete box-drawing border.
type BorderDef struct {
	Top         string `json:"top"`
	Bottom      string `json:"bottom"`
	Left        string `json:"left"`
	Right       string `json:"right"`
	TopLeft     string `json:"topLeft"`
	TopRight    string `json:"topRight"`
	BottomLeft  string `json:"bottomLeft"`
	BottomRight string `json:"bottomRight"`
}

// bordersConfig maps border style names to their character definitions.
type bordersConfig struct {
	Borders map[string]BorderDef `json:"borders"`
}

func (c *bordersConfig) normalize() {
	if c.Borders == nil {
		c.Borders = defaultBorderDefs()
	}
}

func loadBordersConfig() bordersConfig {
	loader := ConfigLoader[bordersConfig]{
		RawData:  bordersData,
		Fallback: bordersConfig{Borders: defaultBorderDefs()},
		Normalize: func(c *bordersConfig) {
			c.normalize()
		},
	}
	return loader.Load()
}

// ResolveBorder resolves a border style name (e.g. "rounded", "double") to a
// lipgloss.Border. Returns the rounded border as fallback if the style is
// unknown or empty. This is used by the top-level theme to select the outer
// window border character set.
func ResolveBorder(name string) lipgloss.Border {
	bordersOnce.Do(func() { bordersCache = loadBordersConfig() })
	if name == "" {
		name = "rounded"
	}
	name = strings.ToLower(strings.TrimSpace(name))
	def, ok := bordersCache.Borders[name]
	if !ok {
		def = bordersCache.Borders["rounded"]
	}
	return lipgloss.Border{
		Top:         def.Top,
		Bottom:      def.Bottom,
		Left:        def.Left,
		Right:       def.Right,
		TopLeft:     def.TopLeft,
		TopRight:    def.TopRight,
		BottomLeft:  def.BottomLeft,
		BottomRight: def.BottomRight,
	}
}

// defaultBorderDefs returns a map with the "rounded" border.
// Used as fallback when the JSONC file fails to load.
func defaultBorderDefs() map[string]BorderDef {
	return map[string]BorderDef{
		"rounded": {
			Top: "─", Bottom: "─", Left: "│", Right: "│",
			TopLeft: "╭", TopRight: "╮", BottomLeft: "╰", BottomRight: "╯",
		},
		"double": {
			Top: "═", Bottom: "═", Left: "║", Right: "║",
			TopLeft: "╔", TopRight: "╗", BottomLeft: "╚", BottomRight: "╝",
		},
		"thick": {
			Top: "━", Bottom: "━", Left: "┃", Right: "┃",
			TopLeft: "┏", TopRight: "┓", BottomLeft: "┗", BottomRight: "┛",
		},
		"single": {
			Top: "─", Bottom: "─", Left: "│", Right: "│",
			TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
		},
		"hidden": {
			Top: " ", Bottom: " ", Left: " ", Right: " ",
			TopLeft: " ", TopRight: " ", BottomLeft: " ", BottomRight: " ",
		},
	}
}
