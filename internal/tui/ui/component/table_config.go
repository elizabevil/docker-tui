package component

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

//go:embed table.jsonc
var tableDefaultData []byte

// TableStyleConfig 对应 table.jsonc 顶层结构
type TableStyleConfig struct {
	Table       TableLayoutConfig          `json:"table"`
	RowStyles   map[string]styleRef        `json:"rowStyles"`
	StateStyles map[string]styleRef        `json:"stateStyles"`
	Footer      styleRef                   `json:"footer"`
	Pages       map[string]PageTableConfig `json:"pages"`
}

// TableLayoutConfig 表格布局参数
type TableLayoutConfig struct {
	ColumnSpacing     int              `json:"columnSpacing"`
	RowPrefix         string           `json:"rowPrefix"`
	RowPrefixSelected string           `json:"rowPrefixSelected"`
	MinColumnWidth    int              `json:"minColumnWidth"`
	SelectionInfo     SelectionInfoCfg `json:"selectionInfo"`
	RowSpacing        int              `json:"rowSpacing"` // 行间距（额外空行数）
}

// SelectionInfoCfg 选中项详细信息显示配置
type SelectionInfoCfg struct {
	Enabled    bool   `json:"enabled"`
	PadLines   int    `json:"padLines"`   // 预览与表格之间的空行数，默认 1
	Color      string `json:"color"`      // 文字颜色 palette 名，默认 "white"
	Background string `json:"background"` // 背景颜色 palette 名，可选
}

// PageTableConfig 单页的列定义与断点
type PageTableConfig struct {
	Columns     []ColumnStyle  `json:"columns"`
	Breakpoints map[string]int `json:"breakpoints"`
}

// ColumnStyle 表头、列宽、列内联样式（类似 Compose TextStyle）
type ColumnStyle struct {
	Key   string   `json:"key"`
	Width int      `json:"width"`
	Style styleRef `json:"style"`
}

func (c *TableStyleConfig) normalize() {
	if c.Table.ColumnSpacing <= 0 {
		c.Table.ColumnSpacing = 1
	}
	if c.Table.RowPrefix == "" {
		c.Table.RowPrefix = "  "
	}
	if c.Table.RowPrefixSelected == "" {
		c.Table.RowPrefixSelected = "▸ "
	}
	if c.Table.MinColumnWidth <= 0 {
		c.Table.MinColumnWidth = 8
	}
}

// validateColors 启动时校验所有列颜色名（仅一次，不刷屏）。
func (c *TableStyleConfig) validateColors() {
	for pageName, page := range c.Pages {
		for _, col := range page.Columns {
			if col.Style.Color != "" && !isValidPaletteColor(col.Style.Color) {
				fmt.Fprintf(os.Stderr, "[dtui] table.jsonc: %s.%s: unknown color %q\n", pageName, col.Key, col.Style.Color)
			}
		}
	}
}

var validPaletteColors = map[string]struct{}{
	"green": {}, "cyan": {}, "blue": {}, "red": {},
	"yellow": {}, "orange": {}, "purple": {},
	"white": {}, "gray": {}, "dark": {}, "surface": {}, "bg": {},
}

func isValidPaletteColor(name string) bool {
	_, ok := validPaletteColors[name]
	return ok
}

func (c *TableStyleConfig) pageConfig(page string) (PageTableConfig, bool) {
	p, ok := c.Pages[page]
	return p, ok
}

var tableCfg TableStyleConfig

func init() {
	loader := ConfigLoader[TableStyleConfig]{
		RawData:  tableDefaultData,
		Fallback: defaultTableConfig(),
		Normalize: func(c *TableStyleConfig) {
			c.normalize()
		},
	}
	tableCfg = loader.Load()
	tableCfg.validateColors()
}

func defaultTableConfig() TableStyleConfig {
	return TableStyleConfig{
		Table: TableLayoutConfig{
			ColumnSpacing:     1,
			RowPrefix:         "  ",
			RowPrefixSelected: "\u25b8 ",
			MinColumnWidth:    8,
		},
	}
}

// GetTableLayout 返回表格布局参数（间距、前缀）
func GetTableLayout() TableLayoutConfig {
	return tableCfg.Table
}

// SelectionInfoEnabled 返回是否显示选中项详细信息。
func SelectionInfoEnabled() bool {
	return tableCfg.Table.SelectionInfo.Enabled
}

// SelectionPreviewPadding 返回预览与表格之间的空行数。
func SelectionPreviewPadding() int {
	n := tableCfg.Table.SelectionInfo.PadLines
	if n <= 0 {
		return 1
	}
	return n
}

// RowSpacing 返回表格行间距（额外空行数）。
func RowSpacing() int {
	n := tableCfg.Table.RowSpacing
	if n < 0 {
		return 0
	}
	return n
}

// GetRowStyle 返回行状态样式（selected / normal / alt / marked）
func GetRowStyle(name string) styleRef {
	if s, ok := tableCfg.RowStyles[name]; ok {
		return s
	}
	return styleRef{}
}

// GetStateStyle 返回容器状态颜色（running / exited / paused …）
func GetStateStyle(containerState string) styleRef {
	if s, ok := tableCfg.StateStyles[containerState]; ok {
		return s
	}
	// 兼容旧版命名（stateRunning → running）
	switch containerState {
	case state.ContainerStateRunning:
		return tableCfg.StateStyles["running"]
	case state.ContainerStateStopped, state.ContainerStateExited, state.ContainerStateDead:
		return tableCfg.StateStyles["stopped"]
	case state.ContainerStatePaused:
		return tableCfg.StateStyles["paused"]
	case state.ContainerStateCreated:
		return tableCfg.StateStyles["created"]
	case state.ContainerStateRestarting:
		return tableCfg.StateStyles["restarting"]
	}
	return styleRef{}
}

// GetTableFooterStyle 返回表格页脚样式
func GetTableFooterStyle() styleRef {
	return tableCfg.Footer
}

// GetPageColumns 返回某页的列定义
func GetPageColumns(page string) []ColumnStyle {
	if p, ok := tableCfg.pageConfig(page); ok {
		return p.Columns
	}
	return nil
}

// GetPageBreakpoints 返回某页的响应式断点
func GetPageBreakpoints(page string) map[string]int {
	if p, ok := tableCfg.pageConfig(page); ok {
		return p.Breakpoints
	}
	return nil
}

// GetPageColumnStyles 从 table.jsonc 读取页面列样式，按 colsDef 顺序返回。
func GetPageColumnStyles(page string, colsDef []tables.ColumnDef) []ColumnStyle {
	pageCols := GetPageColumns(page)
	if len(pageCols) == 0 {
		return nil
	}
	byKey := make(map[string]ColumnStyle, len(pageCols))
	for _, pc := range pageCols {
		byKey[pc.Key] = pc
	}
	styles := make([]ColumnStyle, len(colsDef))
	for i, cd := range colsDef {
		styles[i] = byKey[cd.Key]
	}
	return styles
}
