package component

import (
	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
	"github.com/elizabevil/docker-tui/internal/tui/utils"
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
	dimStyle := GetStyle("footer")
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
	if ref.Color != "" {
		if c := style.Color(ref.Color); c != nil {
			s = s.Foreground(c)
		}
	}
	if ref.Background != "" {
		if c := style.Color(ref.Background); c != nil {
			s = s.Background(c)
		}
	}
	if ref.Bold {
		s = s.Bold(true)
	}
	if ref.Faint {
		s = s.Faint(true)
	}
	return s
}
