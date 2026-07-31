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
	Table        TableLayoutConfig   `json:"table"`
	RowStyles    map[string]styleRef `json:"rowStyles"`
	StateStyles  map[string]styleRef `json:"stateStyles"`
	ColumnStyles map[string]styleRef `json:"columnStyles"`
}

// TableLayoutConfig 表格布局参数
type TableLayoutConfig struct {
	ColumnSpacing     int              `json:"columnSpacing"`
	RowPrefix         string           `json:"rowPrefix"`
	RowPrefixSelected string           `json:"rowPrefixSelected"`
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

// ColumnStyle is the resolved visual style for one rendered column.
type ColumnStyle struct {
	Style styleRef `json:"style"`
}

func (c *TableStyleConfig) normalize() {
	if c.Table.ColumnSpacing <= 0 {
		c.Table.ColumnSpacing = 2
	}
	if c.Table.RowPrefix == "" {
		c.Table.RowPrefix = "  "
	}
	if c.Table.RowPrefixSelected == "" {
		c.Table.RowPrefixSelected = "▸ "
	}
}

// validateColors 启动时校验所有列颜色名（仅一次，不刷屏）。
func (c *TableStyleConfig) validateColors() {
	for key, columnStyle := range c.ColumnStyles {
		if columnStyle.Color != "" && !isValidPaletteColor(columnStyle.Color) {
			fmt.Fprintf(os.Stderr, "[dtui] table.jsonc: columnStyles.%s: unknown color %q\n", key, columnStyle.Color)
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
			ColumnSpacing:     2,
			RowPrefix:         "  ",
			RowPrefixSelected: "\u25b8 ",
		},
	}
}

// GetTableLayout 返回表格布局参数（间距、前缀）
func GetTableLayout() TableLayoutConfig {
	return tableCfg.Table
}

// SelectionInfoEnabled reports whether resource selection previews are shown.
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

// GetColumnStyles resolves semantic column styles by column key.
func GetColumnStyles(colsDef []tables.ColumnDef) []ColumnStyle {
	if len(tableCfg.ColumnStyles) == 0 {
		return nil
	}
	styles := make([]ColumnStyle, len(colsDef))
	for i, cd := range colsDef {
		styles[i] = ColumnStyle{Style: tableCfg.ColumnStyles[cd.Key]}
	}
	return styles
}
