package component

import (
	"embed"
	"image/color"

	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
)

// FlexTableConfig 表格布局参数（间距、前缀、最小列宽）
type FlexTableConfig struct {
	ColumnSpacing     int    `json:"columnSpacing"`
	RowPrefix         string `json:"rowPrefix"`
	RowPrefixSelected string `json:"rowPrefixSelected"`
	MinColumnWidth    int    `json:"minColumnWidth"`
}

// Config 是 component 包的全局配置，对应 config.jsonc 结构。
// 仅含列宽配置 + 搜索/面包屑参数；表格专有样式移至 table.jsonc。
type Config struct {
	FlexTable FlexTableConfig       `json:"flexTable"`
	Cols      map[string]ColProfile `json:"cols"`
}

func (c *Config) normalize() {
	if c.FlexTable.ColumnSpacing <= 0 {
		c.FlexTable.ColumnSpacing = 1
	}
	if c.FlexTable.RowPrefix == "" {
		c.FlexTable.RowPrefix = "  "
	}
	if c.FlexTable.RowPrefixSelected == "" {
		c.FlexTable.RowPrefixSelected = "› "
	}
	if c.FlexTable.MinColumnWidth <= 0 {
		c.FlexTable.MinColumnWidth = 8
	}
}

// ColProfile 某页的列宽配置（百分比 + 响应式断点 + 显示阈值）
type ColProfile struct {
	Pct     []int          `json:"pct"`
	More    []int          `json:"more"`
	Wide    []int          `json:"wide"`
	Compact []int          `json:"compact"`
	Show    map[string]int `json:"show"`
}

//go:embed config.jsonc styles.jsonc
var embeddedConfig embed.FS

var active Config

func init() {
	active = loadConfig()
	loadComponentStyles()
}

func loadConfig() Config {
	data, err := embeddedConfig.ReadFile("config.jsonc")
	if err != nil {
		cfg := defaultConfig()
		cfg.normalize()
		return cfg
	}

	// 关键节点：统一走 ConfigLoader，减少页面/组件内散落的解析与兜底逻辑。
	loader := ConfigLoader[Config]{
		RawData:  data,
		Fallback: defaultConfig(),
		Normalize: func(c *Config) {
			c.normalize()
		},
	}
	return loader.Load()
}

func defaultConfig() Config {
	return Config{
		FlexTable: FlexTableConfig{
			ColumnSpacing:     1,
			RowPrefix:         "  ",
			RowPrefixSelected: "\u203a ",
			MinColumnWidth:    8,
		},
	}
}

// GetFlexConfig 返回 FlexTable 布局参数。
func GetFlexConfig() FlexTableConfig {
	return active.FlexTable
}

// GetColor 按调色板名返回颜色，未知名返回白色兜底。
// 简单委托给 style.Color，仅为保持 API 兼容。
func GetColor(name string) color.Color {
	return style.Color(name)
}
