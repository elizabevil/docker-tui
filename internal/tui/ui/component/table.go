package component

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// SelectionProjector projects the current selection into a compact preview.
type SelectionProjector interface {
	SelectionInfo() string
}

// SelectionInfoProvider is kept as a compatibility alias during page migration.
type SelectionInfoProvider = SelectionProjector

// FuncInfoProvider 包装函数实现 SelectionInfoProvider。
type FuncInfoProvider struct {
	Fn func() string
}

func (p *FuncInfoProvider) SelectionInfo() string {
	if p == nil || p.Fn == nil {
		return ""
	}
	return p.Fn()
}

// TableData defines a complete resource table for rendering.
type TableData struct {
	Cols     []tables.ColumnDef
	Rows     [][]string
	Selected int
	TopLabel string // 表格上框标签（如：搜索内容）

	Total  int
	Offset int
	Limit  int

	Banner     string
	BannerW    int
	FooterHint string
	MarkedRows map[int]bool
	BodyHeight int // 竖向布局高度；>0 时统计行按 space-between 贴底

	HeaderOverrides []string

	// SortColKey is the column key currently sorted by (empty = unsorted).
	// When non-empty, a sort arrow (↑/↓) is shown next to that column's header.
	SortColKey string
	SortAsc    bool

	ColStyles []ColumnStyle

	// SelectionProvider 选中项详情接口（若 nil 则不显示）。
	SelectionProvider SelectionProjector
}

// RenderTable renders a complete resource table using shared stable grid tracks.
func RenderTable(d TableData) string {
	n := len(d.Cols)
	if n == 0 {
		return ""
	}

	// Resolve cheap inputs first, then reuse the viewport layout when unchanged.
	headers := resolveHeaders(d)

	// Container width for column layout
	containerW := d.BannerW
	if containerW <= 0 {
		return ""
	}

	tableLayout := GetTableLayout()
	prefix := tableLayout.RowPrefix
	prefixSel := tableLayout.RowPrefixSelected
	prefixWidth := utils.DisplayWidth(prefix)
	if selectedWidth := utils.DisplayWidth(prefixSel); selectedWidth > prefixWidth {
		prefixWidth = selectedWidth
	}
	columnGap := tableLayout.ColumnSpacing
	layout := resolveTableLayout(d, headers, max(1, containerW-prefixWidth), columnGap)
	headers = layout.headers
	colW := layout.widths

	// Row renderer — column layout computed once, reused for all rows
	rr := &RowRenderer{
		colW:         colW,
		gapStrings:   layout.gapStrings,
		rowPrefix:    prefix,
		rowPrefixSel: prefixSel,
		colStyles:    d.ColStyles,
		rowWidth:     containerW,
		Background:   resolveTableBackground(),
	}

	selectionInfo := ""
	if d.SelectionProvider != nil {
		selectionInfo = d.SelectionProvider.SelectionInfo()
	}

	var sb strings.Builder

	if d.TopLabel != "" {
		sb.WriteString(renderTopFrameLabel(d.TopLabel, containerW))
		sb.WriteString("\n")
	}

	// SelectionInfo：选中项预览（表格上方，居中，空行分隔）
	if selectionInfo != "" {
		centered := lipgloss.NewStyle().Width(containerW).Align(lipgloss.Center).Render(renderSelectionInfo(selectionInfo))
		sb.WriteString(centered)
		sb.WriteString(strings.Repeat("\n", SelectionPreviewPadding()))
	}

	// Page info: 当前显示行（右对齐）
	if d.Total > 0 {
		end := min(d.Offset+len(d.Rows), d.Total)
		pageInfo := fmt.Sprintf("%d-%d/%d", d.Offset+1, end, d.Total)
		if d.FooterHint != "" {
			pageInfo += " " + BorderLineVertical + " " + d.FooterHint
		}
		rightAligned := lipgloss.NewStyle().Width(containerW).Align(lipgloss.Right).Render(GetStyle(StyleDim).Render(pageInfo))
		sb.WriteString(rightAligned)
		sb.WriteString("\n")
	}

	// 表头（与 page info 之间空行分隔）

	// Header line
	sb.WriteString(rr.RenderHeader(headers))
	sb.WriteString("\n")

	// Data rows
	rowSpacing := RowSpacing()
	for ri, cells := range d.Rows {
		marked := d.MarkedRows != nil && d.MarkedRows[ri]
		sb.WriteString(rr.RenderRow(cells, ri, marked, ri == d.Selected, ri%2 == 1))
		sb.WriteString("\n")
		if rowSpacing > 0 {
			sb.WriteString(strings.Repeat("\n", rowSpacing))
		}
	}

	// Padding（表体行区）
	limit := d.Limit
	if limit <= 0 {
		limit = len(d.Rows)
	}
	for p := len(d.Rows); p < limit; p++ {
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// adaptiveGapWidths distributes unused viewport space across column gutters.
// This keeps content-driven tracks compact while aligning the final track with
// the right edge. Remainders are spread left-to-right and differ by at most one cell.
func adaptiveGapWidths(base, count, extra int) []int {
	if count <= 0 {
		return nil
	}
	widths := make([]int, count)
	share, remainder := max(0, extra)/count, max(0, extra)%count
	for i := range widths {
		widths[i] = max(1, base) + share
		if i < remainder {
			widths[i]++
		}
	}
	return widths
}

func renderTopFrameLabel(label string, width int) string {
	if width < 8 {
		return GetStyle(StyleDim).Render(label)
	}
	left := BorderLineHorizontal + BorderLineHorizontal + " " + label + " "
	remain := max(width-utils.VisibleLen(left), 0)
	return GetStyle(StyleDim).Render(left + strings.Repeat(BorderLineHorizontal, remain))
}

// resolveHeaders resolves i18n keys and applies HeaderOverrides,
// then appends a sort arrow (↑/↓) on the active sort column.
func resolveHeaders(d TableData) []string {
	n := len(d.Cols)
	headers := make([]string, n)
	for i, cd := range d.Cols {
		h := ""
		if i < len(d.HeaderOverrides) && d.HeaderOverrides[i] != "" {
			h = d.HeaderOverrides[i]
		} else {
			h = cd.Header
			if strings.HasPrefix(h, "table.") {
				h = i18n.T(h)
			}
		}
		// Append sort indicator on the active sort column
		if d.SortColKey != "" && cd.Key == d.SortColKey {
			if d.SortAsc {
				h += " " + keys.KUp // ↑
			} else {
				h += " " + keys.KDown // ↓
			}
		}
		headers[i] = h
	}
	return headers
}

// BuildMarkedRows maps visible row indices to true for items whose IDs are in markedIDs.
func BuildMarkedRows[T any](rows [][]string, items []T, offset int, markedIDs map[string]bool, idFn func(T) string) map[int]bool {
	if len(markedIDs) == 0 {
		return nil
	}
	result := make(map[int]bool)
	for i := offset; i < len(items) && i-offset < len(rows); i++ {
		if markedIDs[idFn(items[i])] {
			result[i-offset] = true
		}
	}
	return result
}

// RenderTableFooter renders a standalone table footer line.
func RenderTableFooter(offset, end, total int, hint string) string {
	left := fmt.Sprintf(" %d-%d/%d", offset+1, end, total)
	ft := left
	if hint != "" {
		ft = left + " " + BorderLineVertical + " " + hint
	}
	return GetStyle(StyleFooter).Render(ft)
}

// renderSelectionInfo 按配置样式渲染选中项预览信息。
func renderSelectionInfo(text string) string {
	return GetStyle(StyleDetailSelection).Render(text)
}

// resolveTableBackground returns the table's component-level background,
// or "" when the table should be transparent. The transparent value is the
// fallback: the global renderAppBackground wrapper provides the fill for the
// table area, and any row that explicitly sets a Background overrides it.
func resolveTableBackground() string {
	return ""
}
