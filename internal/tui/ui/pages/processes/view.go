package processes

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func Render(processes state.ProcessState, width, height int) string {
	if processes.Loading {
		return i18n.T("msg.loading")
	}
	if processes.Error != "" {
		return component.GetStyle("toastError").Render(processes.Error)
	}
	if len(processes.Titles) == 0 || len(processes.Rows) == 0 {
		return component.GetStyle("dim").Render(i18n.T("container.top.empty"))
	}
	cols := make([]tables.ColumnDef, len(processes.Titles))
	widths := make([]int, len(processes.Titles))
	available := max(width-4-(len(cols)-1), len(cols))
	for i, title := range processes.Titles {
		cols[i] = tables.ColumnDef{Key: strings.ToLower(title), Header: title, Flex: 1}
		widths[i] = max(1, available/len(cols))
	}
	rowHeight := component.CalcRowHeight(height)
	offset := processes.ViewOffset
	component.EnsureVisible(&offset, processes.Cursor, rowHeight, len(processes.Rows))
	end := min(len(processes.Rows), offset+rowHeight)
	rows := processes.Rows[offset:end]
	return component.RenderTable(component.TableData{
		Cols: cols, Widths: widths, Rows: rows, Selected: processes.Cursor - offset,
		Total: len(processes.Rows), Offset: offset, Limit: rowHeight, BodyHeight: height,
		BannerW: width - 4, FooterHint: i18n.T("container.top.hint"),
	})
}
