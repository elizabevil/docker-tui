package component

import (
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
)

// TableStyleConfig 对应 table.jsonc 顶层结构
type TableStyleConfig struct {
	Table        config.TableLayoutConfig `json:"table"`
	RowStyles    RowStyleRefs             `json:"rowStyles"`
	StateStyles  StateStyleRefs           `json:"stateStyles"`
	ColumnStyles ColumnStyleRefs          `json:"columnStyles"`
}

type RowStyleRefs struct {
	Selected styleRef `json:"selected"`
	Normal   styleRef `json:"normal"`
	Alt      styleRef `json:"alt"`
	Marked   styleRef `json:"marked"`
}

type StateStyleRefs struct {
	Running    styleRef `json:"running"`
	Stopping   styleRef `json:"stopping"`
	Stopped    styleRef `json:"stopped"`
	Exited     styleRef `json:"exited"`
	Paused     styleRef `json:"paused"`
	Created    styleRef `json:"created"`
	Restarting styleRef `json:"restarting"`
	Dead       styleRef `json:"dead"`
}

type ColumnStyleRefs struct {
	ID       styleRef `json:"id"`
	Name     styleRef `json:"name"`
	Project  styleRef `json:"project"`
	Service  styleRef `json:"service"`
	Registry styleRef `json:"registry"`
	Tag      styleRef `json:"tag"`
	Platform styleRef `json:"platform"`
	State    styleRef `json:"state"`
	Status   styleRef `json:"status"`
	Created  styleRef `json:"created"`
	Time     styleRef `json:"time"`
	Ports    styleRef `json:"ports"`
	Subnet   styleRef `json:"subnet"`
}

type TableLayoutConfig = config.TableLayoutConfig

// ColumnStyle is the resolved visual style for one rendered column.
type ColumnStyle struct {
	Style styleRef `json:"style"`
}

var tableCfg TableStyleConfig

func init() {
	tableCfg = defaultTableConfig()
}

func defaultTableConfig() TableStyleConfig {
	return TableStyleConfig{
		Table: config.DefaultAppConfig().UI.Table,
		RowStyles: RowStyleRefs{
			Selected: styleRef{Bold: true},
			Normal:   styleRef{Color: string(config.ColorTokenOrange)},
			Alt:      styleRef{Faint: true, Background: string(config.ColorTokenDark)},
			Marked:   styleRef{Color: string(config.FallbackColorTableNameForeground), Background: string(config.FallbackColorTableMarkedBackground), Bold: true},
		},
		StateStyles: StateStyleRefs{
			Running: styleRef{Color: string(config.ColorTokenGreen)}, Stopping: styleRef{Color: string(config.ColorTokenYellow)},
			Stopped: styleRef{Color: string(config.ColorTokenWhite), Faint: true}, Exited: styleRef{Color: string(config.ColorTokenWhite), Faint: true},
			Paused: styleRef{Color: string(config.ColorTokenYellow)}, Created: styleRef{Color: string(config.ColorTokenBlue)},
			Restarting: styleRef{Color: string(config.ColorTokenOrange)}, Dead: styleRef{Color: string(config.ColorTokenRed)},
		},
		ColumnStyles: defaultColumnStyles(string(config.FallbackColorTableColumnForeground), string(config.FallbackColorTableNameForeground)),
	}
}

func defaultColumnStyles(columnForeground, nameForeground string) ColumnStyleRefs {
	standard := styleRef{Color: columnForeground}
	dimmed := styleRef{Color: columnForeground, Faint: true}
	bold := styleRef{Color: columnForeground, Bold: true}
	return ColumnStyleRefs{
		ID: dimmed, Name: styleRef{Color: nameForeground, Bold: true}, Project: bold, Service: bold,
		Registry: dimmed, Tag: styleRef{Color: string(config.ColorTokenCyan)}, Platform: dimmed,
		State: standard, Status: standard, Created: dimmed, Time: dimmed, Ports: standard, Subnet: standard,
	}
}

func ApplyTableLayout(layout config.TableLayoutConfig) {
	tableCfg.Table = layout
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
	switch name {
	case "selected":
		return tableCfg.RowStyles.Selected
	case "normal":
		return tableCfg.RowStyles.Normal
	case "alt":
		return tableCfg.RowStyles.Alt
	case "marked":
		return tableCfg.RowStyles.Marked
	}
	return styleRef{}
}

// GetStateStyle 返回容器状态颜色（running / exited / paused …）
func GetStateStyle(containerState string) styleRef {
	switch containerState {
	case state.ContainerStateRunning:
		return tableCfg.StateStyles.Running
	case state.ContainerStateStopping:
		return tableCfg.StateStyles.Stopping
	case state.ContainerStateStopped:
		return tableCfg.StateStyles.Stopped
	case state.ContainerStateExited:
		return tableCfg.StateStyles.Exited
	case state.ContainerStatePaused:
		return tableCfg.StateStyles.Paused
	case state.ContainerStateCreated:
		return tableCfg.StateStyles.Created
	case state.ContainerStateRestarting:
		return tableCfg.StateStyles.Restarting
	case state.ContainerStateDead:
		return tableCfg.StateStyles.Dead
	}
	return styleRef{}
}

// GetColumnStyles resolves semantic column styles by column key.
func GetColumnStyles(colsDef []tables.ColumnDef) []ColumnStyle {
	styles := make([]ColumnStyle, len(colsDef))
	for i, cd := range colsDef {
		styles[i] = ColumnStyle{Style: tableCfg.ColumnStyles.lookup(cd.Key)}
	}
	return styles
}

func (s ColumnStyleRefs) lookup(key string) styleRef {
	switch key {
	case "id":
		return s.ID
	case "name":
		return s.Name
	case "project":
		return s.Project
	case "service":
		return s.Service
	case "registry":
		return s.Registry
	case "tag":
		return s.Tag
	case "platform":
		return s.Platform
	case "state":
		return s.State
	case "status":
		return s.Status
	case "created":
		return s.Created
	case "time":
		return s.Time
	case "ports":
		return s.Ports
	case "subnet":
		return s.Subnet
	default:
		return styleRef{}
	}
}
