package containers

import (
	"fmt"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/filter"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"
)

var tc = tables.MustLoad("containers")

var profileSelector component.TableProfileSelector = component.ContainerProfileSelector{
	Default:       "default",
	DefaultStats:  "default_stats",
	More:          "more",
	MoreStats:     "more_stats",
	MoreMinWidth:  tc.ShowBreak("more"),
	StatsEnabled:  tc.Stats.Enabled,
	StatsMinWidth: 60,
}

func RenderList(cm *state.ContainerListModel, width int, panelHeight int, markedIDs map[string]bool, selectionDisabled bool) string {
	if cm == nil {
		return i18n.T("msg.loading")
	}
	w := max(width, 52)

	profile := profileSelector.Select(w)

	colsDef := tc.Columns.Get(profile)
	if len(colsDef) == 0 {
		return i18n.T("msg.loading")
	}

	items := cm.SortedItems()
	total := len(items)
	if total == 0 {
		if cm.FilterText() != "" {
			return component.GetStyle(component.StyleDim).Render(i18n.T("msg.no_containers_match"))
		}
		return component.GetStyle(component.StyleDim).Render(i18n.T("msg.no_containers"))
	}

	rowHeight := component.CalcTableRowHeight(panelHeight, !selectionDisabled)
	component.EnsureVisible(&cm.ViewOffset, cm.Cursor, rowHeight, total)
	viewOffset := cm.ViewOffset

	running, exited, createdSt := 0, 0, 0
	for _, c := range items {
		switch c.State {
		case state.ContainerStateRunning:
			running++
		case state.ContainerStateExited:
			exited++
		case state.ContainerStateCreated:
			createdSt++
		}
	}

	banner := filter.BannerPrefixForCount(cm, total, cm.Len())
	if !selectionDisabled && cm.Cursor < total {
		sel := items[cm.Cursor]
		b := utils.ShortID(sel.ID)
		if sel.Name != "" {
			if b != "" {
				b += "  "
			}
			b += sel.Name
		}
		banner += b
	}

	rows := make([][]string, 0, rowHeight)
	for i := viewOffset; i < total && len(rows) < rowHeight; i++ {
		c := items[i]
		created := utils.FormatCreated(c.Created)
		status := c.Status
		portList := FormatPorts(c.PortBindings, w < 80)
		for pi, p := range portList {
			if len(rows) >= rowHeight {
				break
			}
			cells := make([]string, len(colsDef))
			for j, cd := range colsDef {
				switch {
				case cd.Key == "stats":
					if pi == 0 {
						if st, ok := cm.Stats[c.ID]; ok {
							cells[j] = statsString(st, tc)
						}
					}
				case pi == 0:
					cells[j] = CellValue(cd.Key, &c, status, p, created)
				case cd.Key == "ports":
					cells[j] = p
				}
			}
			rows = append(rows, cells)
		}
	}

	hint := fmt.Sprintf("%d %s, %d %s, %d %s", running, i18n.T("container.state.running"), exited, i18n.T("container.state.exited"), createdSt, i18n.T("container.state.created"))

	colStyles := component.GetColumnStyles(colsDef)

	// 选中项详情预览
	var selProv component.SelectionInfoProvider
	if !selectionDisabled {
		selProv = component.NewSelectionProviderFromFn(func() string {
			if cm.Cursor >= total {
				return ""
			}
			sel := items[cm.Cursor]
			return fmt.Sprintf("%s  %s  %s", utils.ShortID(sel.ID), sel.Name, utils.ShortImage(sel.Image))
		})
	}

	selected := -1
	if !selectionDisabled {
		selected = selectedRowForCursor(items, viewOffset, cm.Cursor)
	}

	view := component.RenderTable(component.TableData{
		Cols:              colsDef,
		Rows:              rows,
		Selected:          selected,
		Total:             total,
		Offset:            viewOffset,
		Limit:             rowHeight,
		Banner:            banner,
		BannerW:           w,
		FooterHint:        hint,
		BodyHeight:        panelHeight,
		MarkedRows:        buildMarkedRows(rows, items, viewOffset, markedIDs),
		ColStyles:         colStyles,
		SelectionProvider: selProv,
		SortColKey:        containerSortColKey(cm.SortBy),
		SortAsc:           cm.SortAsc,
	})
	if banner := component.MarkedItemsBanner(state.PanelContainers, markedIDs); banner != "" {
		return banner + "\n" + view
	}
	return view
}

func selectedRowForCursor(items []dockerclient.ContainerSummary, offset, cursor int) int {
	if cursor < offset {
		return 0
	}
	row := 0
	for i := offset; i < len(items) && i < cursor; i++ {
		row += len(FormatPorts(items[i].PortBindings, false))
	}
	return row
}

func buildMarkedRows(rows [][]string, items []dockerclient.ContainerSummary, offset int, markedIDs map[string]bool) map[int]bool {
	if len(markedIDs) == 0 {
		return nil
	}
	result := make(map[int]bool)
	rowIdx := 0
	for i := offset; i < len(items) && rowIdx < len(rows); i++ {
		portsRows := len(FormatPorts(items[i].PortBindings, false))
		if markedIDs[items[i].ID] {
			for k := 0; k < portsRows && rowIdx+k < len(rows); k++ {
				result[rowIdx+k] = true
			}
		}
		rowIdx += portsRows
	}
	return result
}

func CellValue(key string, c *dockerclient.ContainerSummary, status, ports, created string) string {
	switch key {
	case "id":
		return utils.ShortID(c.ID)
	case "name":
		return c.Name
	case "image":
		return utils.ShortImage(c.Image)
	case "state":
		return component.RenderStateText(c.State) + " " + status
	case "ports":
		return ports
	case "mounts":
		if c.MountCount > 0 {
			return fmt.Sprintf("%d", c.MountCount)
		}
		return component.StrDash
	case "ip":
		ips := c.OrderedIPs()
		if len(ips) > 0 {
			return strings.Join(ips, ",")
		}
		return component.StrDash
	case "created":
		return created
	default:
		return ""
	}
}

func statsString(st state.ContainerStats, cfg *tables.TableConfig) string {
	var parts []string
	for _, f := range cfg.Stats.Fields {
		switch f {
		case "cpu":
			parts = append(parts, utils.FormatPercent(st.CPU))
		case "mem":
			parts = append(parts, utils.FormatPercent(st.MemPerc))
		case "net_rx":
			parts = append(parts, "RX:"+utils.FormatBytes(st.NetRx))
		case "net_tx":
			parts = append(parts, "TX:"+utils.FormatBytes(st.NetTx))
		}
	}
	return strings.Join(parts, " ")
}

func FormatPorts(bindings []dockerclient.PortBinding, compact bool) []string {
	if len(bindings) == 0 {
		return []string{component.StrDash}
	}
	result := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		container := fmt.Sprintf("%d/%s", binding.ContainerPort, binding.Protocol)
		if binding.HostPort == 0 {
			result = append(result, container)
			continue
		}
		if compact {
			result = append(result, fmt.Sprintf("%d:%s", binding.HostPort, container))
			continue
		}
		host := binding.HostIP
		if host == "" || host == "0.0.0.0" {
			host = ""
		} else if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
			host = "[" + host + "]"
		}
		result = append(result, fmt.Sprintf("%s -> %s:%d", container, host, binding.HostPort))
	}
	return result
}

// containerSortColKey maps the current sort column to a table column key for the sort indicator.
func containerSortColKey(col state.ContainerSortColumn) string {
	switch col {
	case state.ContainerSortByName:
		return "name"
	case state.ContainerSortByID:
		return "id"
	case state.ContainerSortByCPU:
		return "stats"
	case state.ContainerSortByMem:
		return "stats"
	case state.ContainerSortByState:
		return "state"
	case state.ContainerSortByCreated:
		return "created"
	default:
		return ""
	}
}
