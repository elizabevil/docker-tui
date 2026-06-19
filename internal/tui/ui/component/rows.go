package component

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
	"github.com/elizabevil/docker-tui/internal/tui/utils"
)

// RowRenderer 高效渲染数据行，支持预计算列宽、列间距、列级样式。
type RowRenderer struct {
	colW         []int
	gapStr       string
	rowPrefix    string
	rowPrefixSel string
	colStyles    []ColumnStyle
	rowWidth     int // 行总宽，用于背景填充
}

// RenderRow 渲染一行数据。
func (r *RowRenderer) RenderRow(cells []string, rowIdx int, marked, selected, alt bool) string {
	prefix := r.rowPrefix
	if selected {
		prefix = r.rowPrefixSel
	}
	// 选中/标记行统一使用内联方式（StripANSI + 列样式无 \033[0m）
	useInline := selected || marked
	line := r.joinRow(cells, useInline)

	// 选中/标记行：填充至行宽并应用样式
	if useInline {
		mark := ""
		// Marked rows: no prefix, just background color
		content := prefix + mark + line
		// 填充空格至行宽，确保背景铺满
		visLen := utils.VisibleLen(content)
		if r.rowWidth > visLen {
			content += strings.Repeat(" ", r.rowWidth-visLen)
		}
		ref := GetRowStyle("selected")
		if marked {
			ref = GetRowStyle("marked")
		}
		return buildStyle(ref).Render(content)
	}

	switch {
	case alt:
		return buildStyle(GetRowStyle("alt")).Render(prefix + line)
	default:
		return prefix + line
	}
}

// joinRow 拼接一行单元格，按列应用独立样式。
// noReset=true：清除单元格内嵌 ANSI，列样式不追加 \033[0m，外部统一 wrap。
func (r *RowRenderer) joinRow(cells []string, noReset bool) string {
	n := len(cells)
	if n == 0 {
		return ""
	}
	var sb strings.Builder
	for i, cell := range cells {
		if i > 0 {
			sb.WriteString(r.gapStr)
		}
		display := cell
		if i < len(r.colW) && r.colW[i] > 0 {
			display = utils.PadVisible(utils.TruncateVisible(cell, r.colW[i]), r.colW[i])
		}

		if i < len(r.colStyles) {
			sty := r.colStyles[i].Style
			if noReset {
				// 清除单元格内嵌 ANSI（如 RenderStateText 的 \033[0m），再应用前景色
				clean := utils.StripANSI(display)
				sb.WriteString(cellStyleANSI(clean, sty))
			} else {
				s := lipgloss.NewStyle()
				if sty.Color != "" {
					if c := style.Color(sty.Color); c != nil {
						s = s.Foreground(c)
					}
				}
				if sty.Bold {
					s = s.Bold(true)
				}
				if sty.Faint {
					s = s.Faint(true)
				}
				sb.WriteString(s.Render(display))
			}
		} else {
			sb.WriteString(display)
		}
	}
	return sb.String()
}

// cellStyleANSI 构建内联 ANSI 前景色（不闭合），使外层背景持续生效。
func cellStyleANSI(display string, ref styleRef) string {
	var codes string
	if ref.Color != "" {
		if c := style.Color(ref.Color); c != nil {
			r, g, b, _ := c.RGBA()
			codes = fmt.Sprintf("\033[38;2;%d;%d;%dm", r/257, g/257, b/257)
		}
	}
	if ref.Bold {
		codes += "\033[1m"
	}
	if ref.Faint {
		codes += "\033[2m"
	}
	return codes + display
}

// RenderHeader 使用相同的布局渲染表头行。
func (r *RowRenderer) RenderHeader(headers []string) string {
	return r.rowPrefix + r.joinRow(headers, false)
}
