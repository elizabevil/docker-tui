package box

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// TableRow renders one data row of a resource table as a self-contained box.
//
// Earlier versions used `lipgloss.NewStyle().Width(w).Render(...)` with no
// background; padding spaces then fell back to the terminal default, making
// light-theme tables appear with a dark data area. This component explicitly
// carries a Background so the row width is filled by the intended colour.
type TableRow struct {
	Cells         []string
	CellStyles    []StyleName // per-cell style; nil → use StyleDim
	Prefix        string      // leading prefix (e.g. "  ", "║ ")
	PrefixStyle   StyleName   // style for the prefix; "" → no style
	Width         int         // total visible width to fill
	Background    string      // hex; "" → terminal default
	Bold          bool
	Foreground    string // optional override for cells; "" → style default
	TruncateCells bool   // if true, truncate cells to their column widths
	ColWidths     []int  // when TruncateCells is true, per-column widths
	Gap           string // separator between cells, typically " "
}

// Render produces the table row string.
func (tr *TableRow) Render() string {
	if tr.Width <= 0 {
		return tr.Prefix + joinCells(tr.Cells, tr.Gap)
	}

	// Build cell content.
	cells := make([]string, len(tr.Cells))
	cellGap := tr.Gap
	if cellGap == "" {
		cellGap = " "
	}
	for i, c := range tr.Cells {
		display := c
		if tr.TruncateCells && i < len(tr.ColWidths) && tr.ColWidths[i] > 0 {
			display = utils.PadVisible(utils.TruncateVisible(c, tr.ColWidths[i]), tr.ColWidths[i])
		}
		sty := tr.cellStyle(i)
		if cc := backgroundColor(tr.Foreground); cc != nil {
			sty = sty.Foreground(cc)
		}
		if tr.Bold {
			sty = sty.Bold(true)
		}
		cells[i] = sty.Render(display)
	}
	content := tr.renderPrefix()
	if len(cells) > 0 {
		if content != "" {
			content += cells[0]
		} else {
			content = cells[0]
		}
		for _, c := range cells[1:] {
			content += cellGap + c
		}
	}

	// Build the box style with width and background.
	style := lipgloss.NewStyle()
	if c := backgroundColor(tr.Background); c != nil {
		style = style.Background(c)
	}
	style = style.Width(tr.Width)
	return style.Render(content)
}

func (tr *TableRow) cellStyle(i int) lipgloss.Style {
	name := StyleDim
	if i < len(tr.CellStyles) && tr.CellStyles[i] != "" {
		name = tr.CellStyles[i]
	}
	return styleWithBackground(component.GetStyle(name), tr.Background)
}

func (tr *TableRow) renderPrefix() string {
	if tr.Prefix == "" {
		return ""
	}
	name := tr.PrefixStyle
	if name == "" {
		// Prefix without style: just return as-is.
		return tr.Prefix
	}
	style := styleWithBackground(component.GetStyle(name), tr.Background)
	return style.Render(tr.Prefix)
}

func joinCells(cells []string, gap string) string {
	if gap == "" {
		gap = " "
	}
	var out strings.Builder
	for i, c := range cells {
		if i > 0 {
			out.WriteString(gap)
		}
		out.WriteString(c)
	}
	return out.String()
}
