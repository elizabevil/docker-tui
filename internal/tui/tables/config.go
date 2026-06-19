package tables

import (
	"embed"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/elizabevil/docker-tui/internal/data/config"
)

// ColumnDef defines one column in a table profile.
// Use Fixed for absolute-width columns, Flex for growable columns.
// Max caps how wide a Flex column can grow (0 = unlimited).
type ColumnDef struct {
	Key      string `json:"key"`
	Fixed    int    `json:"fixed,omitempty"` // fixed width in chars
	Flex     int    `json:"flex,omitempty"`  // flex-grow weight (0 = no grow)
	Max      int    `json:"max,omitempty"`   // max width cap (0 = unlimited)
	Header   string `json:"header"`
	Sortable bool   `json:"sortable,omitempty"` // column can be sorted by
}

type StatsConfig struct {
	Enabled  bool     `json:"enabled"`
	Interval int      `json:"interval"`
	Fields   []string `json:"fields"`
}

type LayoutRatio struct {
	Left  int `json:"left"`
	Right int `json:"right"`
}

// ComposeLayout 定义 Compose 面板布局参数。
type ComposeLayout struct {
	Ratio LayoutRatio `json:"ratio"`
	Focus FocusConfig `json:"focus"`
}

// FocusConfig 定义 Compose 双栏焦点指示器配置。
type FocusConfig struct {
	ActiveColor   string `json:"activeColor"`
	InactiveColor string `json:"inactiveColor"`
	SepHeight     int    `json:"sepHeight"` // 竖线高度百分比 (0-100)
}

// TableStyle holds per-page table appearance overrides.
// All fields are optional; zero values fall back to global defaults.
type TableStyle struct {
	RowPrefix         string `json:"rowPrefix,omitempty"`
	RowPrefixSelected string `json:"rowPrefixSelected,omitempty"`
	MinColumnWidth    int    `json:"minColumnWidth,omitempty"`
	ColumnSpacing     int    `json:"columnSpacing,omitempty"`
}

// globalStyle holds the default table style, loaded from _global.jsonc.
var globalStyle TableStyle

func init() {
	globalStyle = loadGlobalStyle()
}

// tableStyleJSON 是 JSON 反序列化用的包装类型。
type tableStyleJSON struct {
	TableStyle TableStyle `json:"tableStyle"`
}

func loadGlobalStyle() TableStyle {
	data, err := embeddedTables.ReadFile("_global.jsonc")
	if err != nil {
		return TableStyle{RowPrefix: "  ", RowPrefixSelected: "\u203a ", MinColumnWidth: 8, ColumnSpacing: 1}
	}
	clean := config.StripJSONComments(data)
	var wrapper tableStyleJSON
	if err := sonic.Unmarshal(clean, &wrapper); err != nil {
		return TableStyle{RowPrefix: "  ", RowPrefixSelected: "\u203a ", MinColumnWidth: 8, ColumnSpacing: 1}
	}
	s := wrapper.TableStyle
	if s.RowPrefix == "" {
		s.RowPrefix = "  "
	}
	if s.RowPrefixSelected == "" {
		s.RowPrefixSelected = "\u203a "
	}
	if s.MinColumnWidth <= 0 {
		s.MinColumnWidth = 8
	}
	if s.ColumnSpacing <= 0 {
		s.ColumnSpacing = 1
	}
	return s
}

type TableConfig struct {
	Columns    map[string][]ColumnDef `json:"columns"`
	Stats      StatsConfig            `json:"stats"`
	Show       map[string]int         `json:"show"`
	Layout     *ComposeLayout         `json:"layout,omitempty"`
	TableStyle *TableStyle            `json:"tableStyle,omitempty"`
}

//go:embed *.jsonc
var embeddedTables embed.FS

var cache map[string]*TableConfig

func init() {
	cache = make(map[string]*TableConfig)
}

func Load(name string) (*TableConfig, error) {
	if c, ok := cache[name]; ok {
		return c, nil
	}
	data, err := embeddedTables.ReadFile(name + ".jsonc")
	if err != nil {
		return nil, fmt.Errorf("table config %q not found: %w", name, err)
	}
	clean := config.StripJSONComments(data)
	var tc TableConfig
	if err := sonic.Unmarshal(clean, &tc); err != nil {
		return nil, fmt.Errorf("parse table config %q: %w", name, err)
	}
	if tc.Stats.Interval <= 0 {
		tc.Stats.Interval = 1
	}
	cache[name] = &tc
	return &tc, nil
}

func MustLoad(name string) *TableConfig {
	tc, err := Load(name)
	if err != nil {
		panic(err)
	}
	return tc
}

// ColumnWidths distributes totalWidth among columns.
//
//	Fixed > 0  → absolute width
//	Flex  > 0  → flex-grow weight (shares remaining space proportionally)
//	  Both 0   → treated as Flex:1
//	Max caps any column; overflow redistributes to uncapped flex columns.
//
// "两端对齐" (justify-content: space-between):
//
//	Sum(widths) + spacing = totalWidth. The last column's right edge aligns
//	exactly with the container right edge. No trailing space.
func (tc *TableConfig) ColumnWidths(profile string, totalWidth int) []int {
	cols, ok := tc.Columns[profile]
	if !ok || len(cols) == 0 {
		return nil
	}
	n := len(cols)
	widths := make([]int, n)
	spacing := n - 1
	sumFixed := spacing

	// Phase 1: fixed columns
	for i, c := range cols {
		if c.Fixed > 0 {
			widths[i] = c.Fixed
			sumFixed += c.Fixed
		}
	}

	remain := totalWidth - sumFixed
	if remain < 0 {
		remain = 0
	}

	// Phase 2: water-filling flex distribution
	for remain > 0 {
		type flexCol struct{ idx, weight int }
		var flexCols []flexCol
		flexWeight := 0
		for i, c := range cols {
			if c.Fixed > 0 {
				continue
			}
			w := c.Flex
			if w <= 0 {
				w = 1
			}
			if c.Max > 0 && widths[i] >= c.Max {
				continue // capped, skip
			}
			flexCols = append(flexCols, flexCol{i, w})
			flexWeight += w
		}
		if len(flexCols) == 0 {
			break
		}

		// Proportional distribution
		alloc := 0
		for _, fc := range flexCols {
			add := remain * fc.weight / flexWeight
			widths[fc.idx] += add
			alloc += add
		}
		// Rounding: give remainder to the column with the largest fraction
		if alloc < remain {
			widths[flexCols[len(flexCols)-1].idx] += remain - alloc
		}

		// Clamp to Max, collect overflow
		remain = 0
		for _, fc := range flexCols {
			mx := cols[fc.idx].Max
			if mx > 0 && widths[fc.idx] > mx {
				remain += widths[fc.idx] - mx
				widths[fc.idx] = mx
			}
		}
	}
	// Spillover: when all flex columns are capped but remain > 0,
	// give to the last flex column (overflow its max if needed).
	if remain > 0 {
		for i := len(cols) - 1; i >= 0; i-- {
			if cols[i].Fixed <= 0 {
				widths[i] += remain
				remain = 0
				break
			}
		}
	}

	// Enforce minimum
	for i, c := range cols {
		if c.Fixed > 0 {
			continue
		}
		if widths[i] < 2 {
			widths[i] = 2
		}
	}
	return widths
}

func (tc *TableConfig) ShowBreak(name string) int {
	if tc.Show == nil {
		return 0
	}
	return tc.Show[name]
}

// EffectiveTableStyle merges the page's TableStyle (if set) with global defaults.
// Page values override global defaults only when non-zero/non-empty.
func (tc *TableConfig) EffectiveTableStyle() TableStyle {
	s := globalStyle
	if tc.TableStyle == nil {
		return s
	}
	p := *tc.TableStyle
	if p.RowPrefix != "" {
		s.RowPrefix = p.RowPrefix
	}
	if p.RowPrefixSelected != "" {
		s.RowPrefixSelected = p.RowPrefixSelected
	}
	if p.MinColumnWidth > 0 {
		s.MinColumnWidth = p.MinColumnWidth
	}
	if p.ColumnSpacing > 0 {
		s.ColumnSpacing = p.ColumnSpacing
	}
	return s
}
