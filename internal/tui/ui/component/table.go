package component

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
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
	Widths   []int
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

	RowPrefix         string
	RowPrefixSelected string

	ColStyles []ColumnStyle

	// SelectionProvider 选中项详情接口（若 nil 则不显示）。
	SelectionProvider SelectionProjector
}

// RenderTable renders a complete resource table with 两端对齐 (justified) layout.
// Output: banner + header line + data rows + empty padding + footer line.
func RenderTable(d TableData) string {
	n := len(d.Cols)
	if n == 0 || len(d.Widths) != n {
		return ""
	}

	// Resolve cheap inputs first, then reuse the viewport layout when unchanged.
	headers := resolveHeaders(d)

	// Container width for column layout
	containerW := d.BannerW
	if containerW <= 0 {
		// Fallback recompute from old Widths
		for _, w := range d.Widths {
			containerW += w
		}
		containerW += len(d.Cols) - 1
	}

	layout := defaultTableLayoutCache.resolve(d, headers, containerW)
	headers = layout.headers
	colW := layout.widths
	gap := layout.gap
	gapStr := strings.Repeat(" ", gap)

	// Row renderer — column layout computed once, reused for all rows
	prefix := d.RowPrefix
	if prefix == "" {
		prefix = GetFlexConfig().RowPrefix
	}
	prefixSel := d.RowPrefixSelected
	if prefixSel == "" {
		prefixSel = GetFlexConfig().RowPrefixSelected
	}
	// 计算行总宽度（含前缀和间隙）
	rowW := utils.DisplayWidth(prefix)
	for _, w := range colW {
		rowW += w
	}
	if len(colW) > 1 {
		rowW += len(gapStr) * (len(colW) - 1)
	}

	rr := &RowRenderer{
		colW:         colW,
		gapStr:       gapStr,
		rowPrefix:    prefix,
		rowPrefixSel: prefixSel,
		colStyles:    d.ColStyles,
		rowWidth:     rowW,
	}

	var sb strings.Builder

	if d.TopLabel != "" {
		sb.WriteString(renderTopFrameLabel(d.TopLabel, containerW))
		sb.WriteString("\n")
	}

	// SelectionInfo：选中项预览（表格上方，居中，空行分隔）
	if d.SelectionProvider != nil {
		if info := d.SelectionProvider.SelectionInfo(); info != "" {
			centered := lipgloss.NewStyle().Width(containerW).Align(lipgloss.Center).Render(renderSelectionInfo(info))
			sb.WriteString(centered)
			pad := SelectionPreviewPadding()
			if pad < 1 {
				pad = 1
			}
			sb.WriteString(strings.Repeat("\n", pad))
		}
	}

	// Page info: 当前显示行（右对齐）
	if d.Total > 0 {
		end := d.Offset + len(d.Rows)
		if end > d.Total {
			end = d.Total
		}
		pageInfo := fmt.Sprintf("%d-%d/%d", d.Offset+1, end, d.Total)
		rightAligned := lipgloss.NewStyle().Width(containerW).Align(lipgloss.Right).Render(GetStyle("dim").Render(pageInfo))
		sb.WriteString(rightAligned)
		sb.WriteString("\n")
	}

	// 表头（与 page info 之间空行分隔）
	sb.WriteString("\n")

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

	// Footer
	if d.Total > 0 {
		end := d.Offset + len(d.Rows)
		if end > d.Total {
			end = d.Total
		}
		left := fmt.Sprintf(" %d-%d/%d", d.Offset+1, end, d.Total)
		right := ""
		if d.FooterHint != "" {
			right = d.FooterHint
		}
		ft := left
		if right != "" {
			ft = left + " │ " + right
		}
		sb.WriteString(GetStyle("footer").Render(ft))
	}
	return sb.String()
}

func renderTopFrameLabel(label string, width int) string {
	if width < 8 {
		return GetStyle("dim").Render(label)
	}
	left := "── " + label + " "
	remain := width - utils.VisibleLen(left)
	if remain < 0 {
		remain = 0
	}
	return GetStyle("dim").Render(left + strings.Repeat("─", remain))
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
				h += " \u2191" // ↑
			} else {
				h += " \u2193" // ↓
			}
		}
		headers[i] = h
	}
	return headers
}

// computeContentWidths computes natural column widths based on content.
// For each column: width = max(visible header length, max visible cell length).
// If d.Widths[i] > 0 and width exceeds it, clamp to d.Widths[i].
func computeContentWidths(d TableData, headers []string) []int {
	n := len(headers)
	colW := make([]int, n)
	for i := 0; i < n; i++ {
		maxW := utils.VisibleLen(headers[i])
		for _, row := range d.Rows {
			if i < len(row) {
				if w := utils.VisibleLen(row[i]); w > maxW {
					maxW = w
				}
			}
		}
		if i < len(d.Widths) && d.Widths[i] > 0 && maxW > d.Widths[i] {
			maxW = d.Widths[i]
		}
		colW[i] = maxW
	}
	return colW
}

// computeGap computes the flexible gap between columns to fill the container width.
// gap = (containerWidth - sum(contentWidths)) / (n-1)
// If gap < 1, it's set to 1. If n <= 1, returns 0.
func computeGap(contentWidths []int, containerWidth int) int {
	n := len(contentWidths)
	if n <= 1 {
		return 0
	}
	total := 0
	for _, w := range contentWidths {
		total += w
	}
	gap := (containerWidth - total) / (n - 1)
	if gap < 1 {
		gap = 1
	}
	return gap
}

// BuildMarkedRows maps visible row indices to true for items whose IDs are in markedIDs.
func BuildMarkedRows[T any](rows [][]string, items []T, offset int, markedIDs map[string]bool, idFn func(T) string) map[int]bool {
	if markedIDs == nil || len(markedIDs) == 0 {
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
		ft = left + " │ " + hint
	}
	return GetStyle("footer").Render(ft)
}

// renderSelectionInfo 按配置样式渲染选中项预览信息。
func renderSelectionInfo(text string) string {
	cfg := tableCfg.Table.SelectionInfo
	s := lipgloss.NewStyle()
	if cfg.Color != "" {
		if c := style.Color(cfg.Color); c != nil {
			s = s.Foreground(c)
		}
	}
	if cfg.Background != "" {
		if c := style.Color(cfg.Background); c != nil {
			s = s.Background(c)
		}
	}
	return s.Render(text)
}
