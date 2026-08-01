package volumes

import (
	_ "embed"
	"fmt"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"
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
	w := width
	if w < 42 {
		w = 42
	}

	profile := profileSelector.Select(w)
	colsDef := tc.Columns[profile]
	if len(colsDef) == 0 {
		return i18n.T("msg.loading")
	}

	items := vm.FilteredItems()
	total := len(items)
	if total == 0 {
		return component.GetStyle("dim").Render(i18n.T("msg.no_volumes"))
	}

	rowHeight := component.CalcTableRowHeight(panelHeight, !selectionDisabled)
	viewOffset := vm.ViewOffset
	component.EnsureVisible(&viewOffset, vm.Cursor, rowHeight, total)

	rows := component.BuildRows(items, colsDef, viewOffset, rowHeight,
		func(vol runtimeapi.Volume, cd tables.ColumnDef, _ int) string {
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
		selected = vm.Cursor - viewOffset
	}
	colStyles := component.GetColumnStyles(colsDef)
	return component.RenderTable(component.TableData{
		Cols:              colsDef,
		Rows:              rows,
		Selected:          selected,
		Total:             total,
		Offset:            viewOffset,
		Limit:             rowHeight,
		BodyHeight:        panelHeight,
		Banner:            banner,
		BannerW:           w,
		MarkedRows:        component.BuildMarkedRows(rows, items, viewOffset, markedIDs, func(v runtimeapi.Volume) string { return v.Name }),
		ColStyles:         colStyles,
		SelectionProvider: selProv,
	})
}

func renderContainers(cm *state.ContainerListModel, width int, volName string, panelHeight int) string {
	w := width
	colsDef := tc.Columns["containers_sub"]
	if len(colsDef) == 0 {
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
				cells[j] = utils.ShortID(c.ID)
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

	colStyles := component.GetColumnStyles(colsDef)
	return component.RenderSelectionBanner(volName, w) + "\n" +
		component.RenderTable(component.TableData{
			Cols:       colsDef,
			Rows:       rows,
			Total:      total,
			Limit:      rowLimit,
			BodyHeight: panelHeight,
			BannerW:    w,
			FooterHint: fmt.Sprintf("%d containers%s │ Esc back", total, more),
			ColStyles:  colStyles,
		})
}
