package tables

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elizabevil/docker-tui/internal/utils"
)

// ColumnDef defines one stable grid track shared by the header and every row.
// Basis is the preferred width before the table fits its viewport.
type ColumnDef struct {
	Key      string `json:"key"`
	Basis    int    `json:"basis,omitempty"`  // preferred width before shrinking
	Min      int    `json:"min,omitempty"`    // minimum width
	Max      int    `json:"max,omitempty"`    // maximum width (0 = unlimited)
	Shrink   int    `json:"shrink,omitempty"` // negative free-space weight
	Fill     int    `json:"fill,omitempty"`   // final free-space weight
	Fixed    int    `json:"fixed,omitempty"`  // deprecated: basis=min=max
	Header   string `json:"header"`
	Sortable bool   `json:"sortable,omitempty"` // column can be sorted by
}

const DefaultColumnGap = 2

type StatsConfig struct {
	Enabled  bool         `json:"enabled"`
	Interval int          `json:"interval"`
	Fields   []StatsField `json:"fields"`
}

type StatsField string

type ColumnProfiles struct {
	Default       []ColumnDef `json:"default"`
	More          []ColumnDef `json:"more"`
	Compact       []ColumnDef `json:"compact"`
	Wide          []ColumnDef `json:"wide"`
	DefaultStats  []ColumnDef `json:"default_stats"`
	MoreStats     []ColumnDef `json:"more_stats"`
	ContainersSub []ColumnDef `json:"containers_sub"`
	ServicesSub   []ColumnDef `json:"services_sub"`
	PodsSub       []ColumnDef `json:"pods_sub"`
	Detail        []ColumnDef `json:"detail"`
}

type NamedColumnProfile struct {
	Name    string
	Columns []ColumnDef
}

func (p ColumnProfiles) All() []NamedColumnProfile {
	profiles := []NamedColumnProfile{
		{"default", p.Default}, {"more", p.More}, {"compact", p.Compact}, {"wide", p.Wide},
		{"default_stats", p.DefaultStats}, {"more_stats", p.MoreStats}, {"containers_sub", p.ContainersSub},
		{"services_sub", p.ServicesSub}, {"pods_sub", p.PodsSub}, {"detail", p.Detail},
	}
	result := profiles[:0]
	for _, profile := range profiles {
		if len(profile.Columns) > 0 {
			result = append(result, profile)
		}
	}
	return result
}

func (p ColumnProfiles) Get(profile string) []ColumnDef {
	switch profile {
	case "default":
		return p.Default
	case "more":
		return p.More
	case "compact":
		return p.Compact
	case "wide":
		return p.Wide
	case "default_stats":
		return p.DefaultStats
	case "more_stats":
		return p.MoreStats
	case "containers_sub":
		return p.ContainersSub
	case "services_sub":
		return p.ServicesSub
	case "pods_sub":
		return p.PodsSub
	case "detail":
		return p.Detail
	default:
		return nil
	}
}

type ShowBreakpoints struct {
	More    int `json:"more"`
	Compact int `json:"compact"`
	Wide    int `json:"wide"`
}

func (b ShowBreakpoints) Get(profile string) int {
	switch profile {
	case "more":
		return b.More
	case "compact":
		return b.Compact
	case "wide":
		return b.Wide
	default:
		return 0
	}
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

type TableConfig struct {
	Columns ColumnProfiles  `json:"columns"`
	Stats   StatsConfig     `json:"stats"`
	Show    ShowBreakpoints `json:"show"`
	Layout  *ComposeLayout  `json:"layout,omitempty"`
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
	var tc TableConfig
	decoder := json.NewDecoder(bytes.NewReader(utils.StripJSONCComments(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&tc); err != nil {
		return nil, fmt.Errorf("parse table config %q: %w", name, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("parse table config %q: trailing content", name)
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

// ColumnWidths distributes totalWidth among columns using the table's fixed gap.
func (tc *TableConfig) ColumnWidths(profile string, totalWidth int) []int {
	cols := tc.Columns.Get(profile)
	if len(cols) == 0 {
		return nil
	}
	return ResolveColumnWidths(cols, totalWidth, DefaultColumnGap)
}

// ResolvedLayout describes the data grid inside a table viewport.
type ResolvedLayout struct {
	Widths        []int
	Gap           int
	ContentWidth  int
	TrailingWidth int
}

// ResolveLayout applies the shared grid contract without content measurements.
func ResolveLayout(cols []ColumnDef, totalWidth int, gapOverride ...int) ResolvedLayout {
	return resolveLayout(cols, nil, totalWidth, resolvedGap(gapOverride))
}

// ResolveContentLayout fits stable tracks to visible cell content. Columns
// expand only when content needs the space; unused viewport width remains
// trailing table space instead of becoming oversized empty cells.
func ResolveContentLayout(cols []ColumnDef, desiredWidths []int, totalWidth, gap int) ResolvedLayout {
	return resolveLayout(cols, desiredWidths, totalWidth, resolvedGap([]int{gap}))
}

func resolvedGap(gapOverride []int) int {
	gap := DefaultColumnGap
	if len(gapOverride) > 0 && gapOverride[0] > 0 {
		gap = gapOverride[0]
	}
	return max(1, gap)
}

func resolveLayout(cols []ColumnDef, desiredWidths []int, totalWidth, gap int) ResolvedLayout {
	if len(cols) == 0 {
		return ResolvedLayout{}
	}

	n := len(cols)
	widths := make([]int, n)
	specs := make([]columnSizing, n)
	columnBudget := max(0, totalWidth-gap*(n-1))
	occupied := 0
	for i, col := range cols {
		specs[i] = resolveColumnSizing(col)
		widths[i] = specs[i].basis
		occupied += widths[i]
	}

	if occupied > columnBudget {
		shrinkColumns(widths, specs, occupied-columnBudget)
	} else if occupied < columnBudget {
		remaining := growColumnsTowardContent(widths, specs, desiredWidths, columnBudget-occupied)
		if desiredWidths == nil {
			fillColumns(widths, specs, remaining)
		}
	}
	contentWidth := gap * (n - 1)
	for _, width := range widths {
		contentWidth += width
	}
	return ResolvedLayout{
		Widths:        widths,
		Gap:           gap,
		ContentWidth:  contentWidth,
		TrailingWidth: max(0, totalWidth-contentWidth),
	}
}

// spreadAcrossFluid distributes any remaining viewport width across all
// fluid columns so adjacent tracks do not leave a visible gap. Columns
// whose max is set cap the expansion; fixed columns are skipped.
func spreadAcrossFluid(widths []int, specs []columnSizing, remaining int) {
	if remaining <= 0 {
		return
	}
	for remaining > 0 {
		flexible := 0
		for i, spec := range specs {
			if !spec.fluid {
				continue
			}
			if spec.max > 0 && widths[i] >= spec.max {
				continue
			}
			flexible++
		}
		if flexible == 0 {
			return
		}
		share := max(1, remaining/flexible)
		consumed := 0
		for i, spec := range specs {
			if !spec.fluid {
				continue
			}
			if spec.max > 0 && widths[i] >= spec.max {
				continue
			}
			add := share
			if spec.max > 0 {
				add = min(add, spec.max-widths[i])
			}
			if add <= 0 {
				continue
			}
			widths[i] += add
			consumed += add
		}
		if consumed == 0 {
			return
		}
		remaining -= consumed
	}
}

// ResolveColumnWidths is kept as a compatibility wrapper for callers that only
// need grid track widths.
func ResolveColumnWidths(cols []ColumnDef, totalWidth int, gapOverride ...int) []int {
	return ResolveLayout(cols, totalWidth, gapOverride...).Widths
}

type columnSizing struct {
	basis  int
	min    int
	max    int
	shrink int
	fill   int
	fluid  bool
}

func resolveColumnSizing(col ColumnDef) columnSizing {
	if col.Fixed > 0 {
		return columnSizing{basis: col.Fixed, min: col.Fixed, max: col.Fixed}
	}

	minimum := col.Min
	if minimum <= 0 {
		minimum = 2
	}
	basis := col.Basis
	if basis <= 0 {
		basis = minimum
	}
	if basis < minimum {
		basis = minimum
	}
	if col.Max > 0 && basis > col.Max {
		basis = col.Max
	}

	shrink := col.Shrink
	if shrink <= 0 {
		shrink = 1
	}
	return columnSizing{basis: basis, min: minimum, max: col.Max, shrink: shrink, fill: max(0, col.Fill), fluid: true}
}

func growColumnsTowardContent(widths []int, specs []columnSizing, desired []int, remaining int) int {
	for remaining > 0 {
		active := 0
		for i, spec := range specs {
			if spec.fluid && i < len(desired) && widths[i] < desired[i] && (spec.max <= 0 || widths[i] < spec.max) {
				active++
			}
		}
		if active == 0 {
			return remaining
		}

		share := max(1, remaining/active)
		before := remaining
		for i, spec := range specs {
			if !spec.fluid || i >= len(desired) || widths[i] >= desired[i] || (spec.max > 0 && widths[i] >= spec.max) {
				continue
			}
			add := min(share, desired[i]-widths[i])
			if spec.max > 0 {
				add = min(add, spec.max-widths[i])
			}
			add = min(add, remaining)
			widths[i] += add
			remaining -= add
			if remaining == 0 {
				return 0
			}
		}
		if remaining == before {
			return remaining
		}
	}
	return 0
}

func fillColumns(widths []int, specs []columnSizing, remaining int) {
	for remaining > 0 {
		weight := 0
		for i, spec := range specs {
			if spec.fill > 0 && (spec.max <= 0 || widths[i] < spec.max) {
				weight += spec.fill
			}
		}
		if weight == 0 {
			return
		}

		before := remaining
		for i, spec := range specs {
			if spec.fill <= 0 || (spec.max > 0 && widths[i] >= spec.max) {
				continue
			}
			add := max(1, before*spec.fill/weight)
			add = min(add, remaining)
			if spec.max > 0 {
				add = min(add, spec.max-widths[i])
			}
			widths[i] += add
			remaining -= add
			if remaining == 0 {
				return
			}
		}
	}
}

func shrinkColumns(widths []int, specs []columnSizing, deficit int) {
	for deficit > 0 {
		weight := 0
		for i, spec := range specs {
			if spec.shrink > 0 && widths[i] > spec.min {
				weight += spec.shrink * widths[i]
			}
		}
		if weight == 0 {
			return
		}

		before := deficit
		for i, spec := range specs {
			if spec.shrink <= 0 || widths[i] <= spec.min {
				continue
			}
			remove := min(max(1, before*spec.shrink*widths[i]/weight), deficit)
			if widths[i]-remove < spec.min {
				remove = widths[i] - spec.min
			}
			widths[i] -= remove
			deficit -= remove
			if deficit == 0 {
				return
			}
		}
	}
}

func (tc *TableConfig) ShowBreak(name string) int {
	return tc.Show.Get(name)
}
