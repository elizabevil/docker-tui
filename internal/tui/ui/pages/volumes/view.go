package volumes

import (
	_ "embed"
	"fmt"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/utils"
)

//go:embed volumes.jsonc
var volumesDefaultData []byte

// volumesConfig maps the JSONC structure for volumes panel defaults.
type volumesConfig struct {
	DefaultSort string `json:"defaultSort"`
}

func (c *volumesConfig) normalize() {
	if c.DefaultSort == "" {
		c.DefaultSort = "name"
	}
}

// DefaultVolumesConfig returns default values parsed from the embedded volumes.jsonc.
func DefaultVolumesConfig() volumesConfig {
	loader := component.ConfigLoader[volumesConfig]{
		RawData:  volumesDefaultData,
		Fallback: volumesConfig{DefaultSort: "name"},
		Normalize: func(c *volumesConfig) {
			c.normalize()
		},
	}
	return loader.Load()
}

var tc = tables.MustLoad("volumes")

var profileSelector component.TableProfileSelector = component.BreakpointProfileSelector{
	Default: "default",
	Rules: []component.ProfileRule{
		{Name: "more", MinWidth: tc.ShowBreak("more")},
	},
}

func RenderList(vm *state.VolumeListModel, cm *state.ContainerListModel, width int, panelHeight int, markedIDs map[string]bool, selectionDisabled bool) string {
	if vm == nil {
		return i18n.T("msg.loading")
	}
	if vm.DetailName != "" {
		return renderContainers(cm, width, vm.DetailName, panelHeight)
	}
	w := width - 8
	if w < 42 {
		w = 42
	}

	profile := profileSelector.Select(w)
	widths := tc.ColumnWidths(profile, w)
	colsDef := tc.Columns[profile]
	if widths == nil {
		return i18n.T("msg.loading")
	}

	items := vm.FilteredItems()
	total := len(items)
	if total == 0 {
		return component.GetStyle("dim").Render(i18n.T("msg.no_volumes"))
	}

	rowHeight := component.CalcRowHeight(panelHeight)
	component.EnsureVisible(&vm.ViewOffset, vm.Cursor, rowHeight, total)

	rows := component.BuildRows(items, colsDef, vm.ViewOffset, rowHeight,
		func(vol dockerclient.VolumeItem, cd tables.ColumnDef, _ int) string {
			switch cd.Key {
			case "name":
				return vol.Name
			case "driver":
				return vol.Driver
			case "scope":
				return vol.Scope
			case "mountpoint":
				return vol.Mountpoint
			case "created":
				return vol.CreatedAt
			}
			return ""
		})

	banner := ""
	if !selectionDisabled {
		if sel := vm.Selected(); sel != nil {
			banner = sel.Name
		}
	}

	var selProv component.SelectionInfoProvider
	if !selectionDisabled {
		selProv = component.NewSelectionProviderFromFn(func() string {
			if vm.Cursor >= len(items) {
				return ""
			}
			v := items[vm.Cursor]
			return v.Name
		})
	}

	selected := -1
	if !selectionDisabled {
		selected = vm.Cursor - vm.ViewOffset
	}
	colStyles := component.GetPageColumnStyles("volume", colsDef)
	ts := tc.EffectiveTableStyle()
	return component.RenderTable(component.TableData{
		Cols:              colsDef,
		Widths:            widths,
		Rows:              rows,
		Selected:          selected,
		Total:             total,
		Offset:            vm.ViewOffset,
		Limit:             rowHeight,
		BodyHeight:        panelHeight,
		Banner:            banner,
		BannerW:           w,
		MarkedRows:        component.BuildMarkedRows(rows, items, vm.ViewOffset, markedIDs, func(v dockerclient.VolumeItem) string { return v.Name }),
		RowPrefix:         ts.RowPrefix,
		RowPrefixSelected: ts.RowPrefixSelected,
		ColStyles:         colStyles,
		SelectionProvider: selProv,
	})
}

func renderContainers(cm *state.ContainerListModel, width int, volName string, panelHeight int) string {
	w := width - 8
	widths := tc.ColumnWidths("containers_sub", w)
	colsDef := tc.Columns["containers_sub"]
	if widths == nil {
		return i18n.T("msg.loading")
	}

	rowLimit := component.CalcRowHeight(panelHeight)
	total := len(cm.Items)

	rows := make([][]string, 0, rowLimit)
	for i := 0; i < total && len(rows) < rowLimit; i++ {
		c := cm.Items[i]
		created := utils.FormatCreated(c.Created)
		cells := make([]string, len(colsDef))
		for j, cd := range colsDef {
			switch cd.Key {
			case "id":
				cells[j] = c.ID[:12]
			case "name":
				cells[j] = c.Name
			case "image":
				cells[j] = utils.ShortImage(c.Image)
			case "state":
				cells[j] = "\u25cf " + c.State
			case "created":
				cells[j] = created
			}
		}
		rows = append(rows, cells)
	}

	more := ""
	if total > rowLimit {
		more = fmt.Sprintf(" +%d more", total-rowLimit)
	}

	colStyles := component.GetPageColumnStyles("volume", colsDef)
	ts := tc.EffectiveTableStyle()
	return component.RenderSelectionBanner(volName, w) + "\n" +
		component.RenderTable(component.TableData{
			Cols:              colsDef,
			Widths:            widths,
			Rows:              rows,
			Total:             total,
			Limit:             rowLimit,
			BodyHeight:        panelHeight,
			FooterHint:        fmt.Sprintf("%d containers%s │ Esc back", total, more),
			RowPrefix:         ts.RowPrefix,
			RowPrefixSelected: ts.RowPrefixSelected,
			ColStyles:         colStyles,
		})
}
