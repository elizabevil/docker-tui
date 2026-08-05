package networks

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/filter"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"
)

var tc = tables.MustLoad("networks")

var profileSelector component.TableProfileSelector = component.BreakpointProfileSelector{
	Default: "default",
	Rules: []component.ProfileRule{
		{Name: "more", MinWidth: tc.ShowBreak("more")},
	},
}

func RenderList(nm *state.NetworkListModel, width int, panelHeight int, markedIDs map[string]bool, selectionDisabled bool) string {
	if nm == nil {
		return component.StrLoading
	}
	w := width
	if w < 42 {
		w = 42
	}

	profile := profileSelector.Select(w)
	colsDef := tc.Columns.Get(profile)
	if len(colsDef) == 0 {
		return component.StrLoading
	}

	items := nm.FilteredItems()
	total := len(items)
	if total == 0 {
		if nm.FilterText() != "" {
			return component.GetStyle(component.StyleDim).Render(i18n.T("msg.no_networks_match"))
		}
		return component.GetStyle(component.StyleDim).Render(i18n.T("msg.no_networks"))
	}

	rowHeight := component.CalcTableRowHeight(panelHeight, !selectionDisabled)
	component.EnsureVisible(&nm.ViewOffset, nm.Cursor, rowHeight, total)
	viewOffset := nm.ViewOffset

	// Header overrides with sort arrows
	arrow := sortArrow(nm.SortAsc)
	overrides := make([]string, len(colsDef))
	for i, cd := range colsDef {
		switch cd.Key {
		case "name":
			if nm.SortBy == state.NetworkSortByName {
				overrides[i] = i18n.T(cd.Header) + arrow
			}
		case "driver":
			if nm.SortBy == state.NetworkSortByDriver {
				overrides[i] = i18n.T(cd.Header) + arrow
			}
		case "created":
			if nm.SortBy == state.NetworkSortByCreated {
				overrides[i] = i18n.T(cd.Header) + arrow
			}
		}
	}

	rows := component.BuildRows(items, colsDef, viewOffset, rowHeight,
		func(net runtimeapi.Network, cd tables.ColumnDef, _ int) string {
			switch cd.Key {
			case "id":
				return utils.ShortID(net.ID)
			case "name":
				if net.Internal {
					return net.Name + " \u26b2"
				}
				return net.Name
			case "driver":
				return net.Driver
			case "scope":
				return net.Scope
			case "subnet":
				if len(net.IPAM) > 0 {
					return net.IPAM[0]
				}
				return ""
			case "ctr":
				if net.Containers > 0 {
					return fmt.Sprintf("%d", net.Containers)
				}
				return component.StrDash
			case "created":
				return utils.FormatCreated(net.Created)
			}
			return ""
		})

	banner := filter.BannerPrefixForCount(nm, total, nm.Len())
	if !selectionDisabled {
		if sel := nm.Selected(); sel != nil {
			label := sel.Name
			if sel.ID != "" {
				label += " (" + utils.ShortID(sel.ID) + ")"
			}
			banner += label
		}
	}

	var selProv component.SelectionInfoProvider
	if !selectionDisabled {
		selProv = component.NewSelectionProviderFromFn(func() string {
			if nm.Cursor >= len(items) {
				return ""
			}
			n := items[nm.Cursor]
			return n.Name + "  " + n.Driver
		})
	}

	selected := -1
	if !selectionDisabled {
		selected = nm.Cursor - viewOffset
	}
	colStyles := component.GetColumnStyles(colsDef)
	return component.RenderTable(component.TableData{
		Cols:              colsDef,
		Rows:              rows,
		Selected:          selected,
		Total:             total,
		Offset:            viewOffset,
		Limit:             rowHeight,
		Banner:            banner,
		BannerW:           w,
		HeaderOverrides:   overrides,
		BodyHeight:        panelHeight,
		MarkedRows:        component.BuildMarkedRows(rows, items, viewOffset, markedIDs, func(n runtimeapi.Network) string { return n.ID }),
		ColStyles:         colStyles,
		SelectionProvider: selProv,
	})
}

func sortArrow(asc bool) string {
	if asc {
		return " " + component.TriangleUp
	}
	return " " + component.TriangleDown
}
