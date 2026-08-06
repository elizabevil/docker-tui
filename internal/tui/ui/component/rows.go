package component

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// RowRenderer 高效渲染数据行，支持预计算列宽、列间距、列级样式。
type RowRenderer struct {
	colW         []int
	gapStrings   []string
	rowPrefix    string
	rowPrefixSel string
	colStyles    []ColumnStyle
	rowWidth     int // 行总宽，用于背景填充
	// Background fills the Width sub-area for normal / alt rows so padding
	// spaces inherit the intended fill instead of the terminal default.
	// Set by the caller (typically RenderTable) to the panel or table
	// background from the theme. An empty string leaves the terminal default.
	Background string
}

// RenderRow 渲染一行数据。
func (r *RowRenderer) RenderRow(cells []string, rowIdx int, marked, selected, alt bool) string {
	prefix := r.rowPrefix
	if selected {
		prefix = r.rowPrefixSel
	}

	// 选中 / 标记行：用 lipgloss Width + Background 强制背景铺满整行，
	// 不依赖手填空格 + 内嵌 ANSI 拼接，避免多列 SGR 互相截断背景。
	if selected || marked {
		// 列宽已在 joinRow 中按 colW + TruncateVisible 严格 padding。
		// 选中行需要保留列前景色，但不能让单元格 reset 截断整行背景。
		line := r.joinRow(cells, true)
		ref := GetRowStyle(RowStyleSelected)
		if marked {
			ref = GetRowStyle(RowStyleMarked)
		}
		s := ref.BuildStyle(utils.WithWidth(r.rowWidth))
		if r.rowPrefix != "" {
			// 选中行在 prefix 上保留高亮前缀（rowPrefixSel = "║ "）
			s = s.PaddingLeft(0)
		}
		// 用 prefix + line 组合，lipgloss 会按 Width + Background 自动填满空格。
		content := prefix + line
		return s.Render(content)
	}

	// 普通行与 alt 行：同样需要 Width + Background，否则 padding 空格
	// 落回终端默认背景，light 主题下整片表格会变深。
	rowStyle := lipgloss.NewStyle().Width(r.rowWidth)
	if r.Background != "" {
		if c, ok := utils.ParseColor(r.Background); ok {
			rowStyle = rowStyle.Background(c)
		}
	}
	if alt {
		rowStyle = GetRowStyle(RowStyleAlt).BuildStyle(utils.FromStyle(rowStyle))
	}
	return rowStyle.Render(prefix + r.joinRow(cells, false))
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
			if i-1 < len(r.gapStrings) {
				sb.WriteString(r.gapStrings[i-1])
			}
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
				sb.WriteString(sty.BuildStyle().Render(display))
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
