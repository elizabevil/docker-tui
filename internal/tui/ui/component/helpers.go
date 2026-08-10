package component

import (
	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// CalcRowHeight returns the number of data rows that fit in the given panel height.
// Account for row spacing from table.jsonc configuration.
func CalcRowHeight(panelHeight int) int {
	h := panelHeight - 3
	if h < 1 {
		return 1
	}
	spacing := RowSpacing()
	if spacing > 0 {
		// Each row takes (1 + spacing) lines, round down
		h = h / (1 + spacing)
	}
	if h < 1 {
		return 1
	}
	return h
}

// CalcTableRowHeight returns the number of data rows left after the table's
// page indicator, blank separator, header, footer, and optional selection
// preview have consumed their fixed lines.
func CalcTableRowHeight(panelHeight int, selectionPreview bool) int {
	overhead := 3
	if selectionPreview {
		overhead += SelectionPreviewPadding()
	}
	h := panelHeight - overhead
	if h < 1 {
		return 1
	}
	if spacing := RowSpacing(); spacing > 0 {
		h /= 1 + spacing
	}
	return max(1, h)
}

// TableHitLayoutInput describes the lines rendered before table data rows.
type TableHitLayoutInput struct {
	TopLabel             bool
	SelectionPreview     bool
	SelectionPreviewPad  int
	Total                int
	BannerLines          int
}

// TableHitLayout maps panel-body rows to visible table data rows.
type TableHitLayout struct {
	dataStartRow     int
	rowSpacing       int
	selectionPreview bool
	previewPad       int
}

// NewTableHitLayout builds the row contract shared by table rendering and mouse hit testing.
func NewTableHitLayout(input TableHitLayoutInput) TableHitLayout {
	start := max(input.BannerLines, 0)
	if input.TopLabel {
		start++
	}
	if input.SelectionPreview {
		pad := input.SelectionPreviewPad
		if pad < 0 {
			pad = 0
		}
		start += 1 + pad
	}
	if input.Total > 0 {
		start++
	}
	start++
	return TableHitLayout{
		dataStartRow:     start,
		rowSpacing:       RowSpacing(),
		selectionPreview: input.SelectionPreview,
		previewPad:       input.SelectionPreviewPad,
	}
}

// HeaderRow returns the zero-based body row that holds the table header.
func (l TableHitLayout) HeaderRow() int {
	if l.dataStartRow <= 0 {
		return 0
	}
	return l.dataStartRow - 1
}

// DataStartRow returns the zero-based body row where data row zero begins.
func (l TableHitLayout) DataStartRow() int {
	return l.dataStartRow
}

// SelectionLineBreaks returns the newline count after the selection preview.
func (l TableHitLayout) SelectionLineBreaks() int {
	if !l.selectionPreview {
		return 0
	}
	return l.previewPad + 1
}

// DataRowAt resolves a body row to a visible data-row index.
func (l TableHitLayout) DataRowAt(bodyRow, visibleRows int) (int, bool) {
	if visibleRows <= 0 {
		return 0, false
	}
	relative := bodyRow - l.dataStartRow
	if relative < 0 {
		return 0, false
	}
	stride := l.rowSpacing + 1
	if relative%stride != 0 {
		return 0, false
	}
	row := relative / stride
	if row >= visibleRows {
		return 0, false
	}
	return row, true
}

// ShortID is a delegate to utils.ShortID.
func ShortID(id string) string { return utils.ShortID(id) }

// FormatSize is a delegate to utils.FormatSize.
func FormatSize(bytes int64) string { return utils.FormatSize(bytes) }

// Truncate is a delegate to utils.Truncate.
func Truncate(s string, maxLen int) string { return utils.Truncate(s, maxLen) }

// StripANSI is a delegate to utils.StripANSI.
func StripANSI(s string) string { return utils.StripANSI(s) }

// TruncateVisible is a delegate to utils.TruncateVisible.
func TruncateVisible(s string, maxVisible int) string { return utils.TruncateVisible(s, maxVisible) }

// PadVisible is a delegate to utils.PadVisible.
func PadVisible(s string, width int) string { return utils.PadVisible(s, width) }

// VisibleLen is a delegate to utils.VisibleLen.
func VisibleLen(s string) int { return utils.VisibleLen(s) }

// RenderSelectionBanner returns a centered selection banner for the current item.
func RenderSelectionBanner(text string, width int) string {
	if text == "" {
		return ""
	}
	dimStyle := GetStyle(StyleFooter)
	centered := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(dimStyle.Render(text))
	return centered + "\n"
}

// ── 参数验证与默认值 ─────────────────────────────────────

// Coerce 确保 val 在 [lo, hi] 范围内。
func Coerce(val, lo, hi int) int {
	if val < lo {
		return lo
	}
	if val > hi {
		return hi
	}
	return val
}

// CoerceMin 确保 val 不小于 min。
func CoerceMin(val, min int) int {
	if val < min {
		return min
	}
	return val
}

// FirstNonZero 返回第一个非零值。
func FirstNonZero[T comparable](vals ...T) T {
	var zero T
	for _, v := range vals {
		if v != zero {
			return v
		}
	}
	return zero
}

// WrapStyle 安全包装 styleRef，空字段不设置。
func WrapStyle(s lipgloss.Style, ref styleRef) lipgloss.Style {
	return ref.BuildStyle(utils.FromStyle(s))
}
