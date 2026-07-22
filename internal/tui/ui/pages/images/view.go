package images

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/pages/containers"
	"github.com/elizabevil/docker-tui/internal/utils"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
)

var tc = tables.MustLoad("images")

var profileSelector component.TableProfileSelector = component.BreakpointProfileSelector{
	Default: "default",
	Rules: []component.ProfileRule{
		{Name: "wide", MinWidth: tc.ShowBreak("wide")},
		{Name: "more", MinWidth: tc.ShowBreak("more")},
		{Name: "compact", MinWidth: tc.ShowBreak("compact")},
	},
}

func RenderList(im *state.ImageListModel, cm *state.ContainerListModel, width int, panelHeight int, markedIDs map[string]bool, selectionDisabled bool) string {
	if im == nil {
		return i18n.T("msg.loading")
	}
	if im.ContainersViewID != "" {
		return renderContainers(im, cm, width, im.ContainersViewID, panelHeight)
	}
	w := width - 8

	profile := profileSelector.Select(w)
	widths := tc.ColumnWidths(profile, w)
	colsDef := tc.Columns[profile]
	if widths == nil {
		return i18n.T("msg.loading")
	}

	items := im.SortedItems()
	total := len(items)
	if total == 0 {
		return component.GetStyle("dim").Render(i18n.T("msg.no_images"))
	}

	rowHeight := component.CalcRowHeight(panelHeight)
	viewOffset := im.ViewOffset
	component.EnsureVisible(&viewOffset, im.Cursor, rowHeight, total)

	// Header overrides with sort arrows
	arrow := sortArrow(im.SortAsc)
	overrides := make([]string, len(colsDef))
	for i, cd := range colsDef {
		switch cd.Key {
		case "name":
			if im.SortBy == state.ImageSortByRepo {
				overrides[i] = resolveHdr(cd) + arrow
			}
		case "id":
			if im.SortBy == state.ImageSortByID {
				overrides[i] = resolveHdr(cd) + arrow
			}
		case "created":
			if im.SortBy == state.ImageSortByCreated {
				overrides[i] = resolveHdr(cd) + arrow
			}
		case "size":
			if im.SortBy == state.ImageSortBySize {
				overrides[i] = resolveHdr(cd) + arrow
			}
		}
	}

	banner := ""
	if !selectionDisabled {
		if sel := im.Selected(); sel != nil {
			label := fullRef(sel)
			if sel.IsManifest {
				label += " (multi-arch)"
			}
			banner = label
		}
	}

	rows := make([][]string, 0, rowHeight)
	containerImages := precomputeContainerIDs(cm)
	for i := viewOffset; i < total && len(rows) < rowHeight; i++ {
		img := items[i]
		_, name, tag := splitRef(img.RepoTags)
		registry := img.Registry
		if registry == "" {
			registry = "\u2014"
		}
		arch := img.Arch
		if img.IsManifest {
			arch = "multi"
		} else if arch == "" {
			arch = "\u2014"
		}
		created := utils.FormatCreated(img.Created)
		size := component.FormatSize(img.Size)
		arrowMark := " "
		if imgHasContainer(containerImages, img.ID, img.RepoTags) {
			arrowMark = component.ArrowRight
		}

		cells := make([]string, len(colsDef))
		for j, cd := range colsDef {
			switch cd.Key {
			case "registry":
				cells[j] = registry
			case "name":
				cells[j] = name
			case "tag":
				cells[j] = tag
			case "id":
				cells[j] = img.ID
			case "arch":
				cells[j] = arch
			case "created":
				cells[j] = created
			case "size":
				cells[j] = size
			}
		}
		cells[0] = arrowMark + " " + cells[0]
		rows = append(rows, cells)
	}

	var selProv component.SelectionInfoProvider
	if !selectionDisabled {
		selProv = component.NewSelectionProviderFromFn(func() string {
			items := im.FilteredItems()
			if im.Cursor >= len(items) {
				return ""
			}
			img := items[im.Cursor]
			reg, name, tag := splitRef(img.RepoTags)
			if reg == "" {
				reg = img.Registry
			}
			if name == "" {
				return img.ID[:16]
			}
			return reg + "/" + name + ":" + tag
		})
	}

	selected := -1
	if !selectionDisabled {
		selected = im.Cursor - viewOffset
	}
	colStyles := component.GetPageColumnStyles("image", colsDef)
	ts := tc.EffectiveTableStyle()
	return component.RenderTable(component.TableData{
		Cols:              colsDef,
		Widths:            widths,
		Rows:              rows,
		Selected:          selected,
		Total:             total,
		Offset:            viewOffset,
		Limit:             rowHeight,
		Banner:            banner,
		BannerW:           w,
		HeaderOverrides:   overrides,
		BodyHeight:        panelHeight,
		MarkedRows:        component.BuildMarkedRows(rows, items, viewOffset, markedIDs, func(img dockerclient.ImageSummary) string { return img.ID }),
		RowPrefix:         ts.RowPrefix,
		RowPrefixSelected: ts.RowPrefixSelected,
		ColStyles:         colStyles,
		SelectionProvider: selProv,
	})
}

func renderContainers(im *state.ImageListModel, cm *state.ContainerListModel, width int, imgID string, panelHeight int) string {
	if cm == nil {
		return i18n.T("msg.loading")
	}
	w := width - 8

	widths := tc.ColumnWidths("containers_sub", w)
	colsDef := tc.Columns["containers_sub"]
	if widths == nil {
		return i18n.T("msg.loading")
	}

	imgShort := imgID[:12]
	var imgNames []string
	for _, item := range im.Items {
		if item.ID == imgID || item.ID[:12] == imgShort {
			for _, tag := range item.RepoTags {
				imgNames = append(imgNames, tag)
			}
			break
		}
	}

	// Filter containers matching this image
	var matched []dockerclient.ContainerSummary
	for _, c := range cm.Items {
		match := strings.Contains(c.Image, imgShort)
		if !match {
			for _, n := range imgNames {
				if strings.Contains(c.Image, n) {
					match = true
					break
				}
			}
		}
		if match {
			matched = append(matched, c)
		}
	}
	if len(matched) == 0 {
		return renderInfoBlock([]string{component.StrNoContainer, "Press Esc to return"}, panelHeight)
	}

	imgRef := fullRef(&im.Items[0])
	for _, item := range im.Items {
		if item.ID == imgID || item.ID[:12] == imgID[:12] {
			imgRef = fullRef(&item)
			break
		}
	}

	total := len(matched)
	rowHeight := component.CalcRowHeight(panelHeight)
	containerCursor := min(max(0, im.ContainerCursor), total-1)
	containerOffset := im.ContainerOffset
	component.EnsureVisible(&containerOffset, containerCursor, rowHeight, total)

	// Override stats header to "CPU/MEM" instead of "STATISTICS"
	overrides := make([]string, len(colsDef))
	for i, cd := range colsDef {
		if cd.Key == "stats" {
			overrides[i] = "CPU/MEM"
		}
	}

	rows := make([][]string, 0, rowHeight)
	for i := containerOffset; i < total && len(rows) < rowHeight; i++ {
		c := matched[i]
		created := utils.FormatCreated(c.Created)
		portList := containers.FormatPorts(c.PortBindings, w < 80)

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
							cells[j] = utils.FormatPercent(st.CPU) + "/" + utils.FormatPercent(st.MemPerc)
						}
					}
				case pi == 0:
					cells[j] = containers.CellValue(cd.Key, &c, c.Status, p, created)
				case cd.Key == "ports":
					cells[j] = p
				}
			}
			rows = append(rows, cells)
		}
	}

	selRow := selectedContainerSubRow(matched, containerOffset, containerCursor)

	selProv := component.NewSelectionProviderFromFn(func() string {
		items := im.FilteredItems()
		if im.Cursor >= len(items) {
			return ""
		}
		img := items[im.Cursor]
		reg, name, tag := splitRef(img.RepoTags)
		if reg == "" {
			reg = img.Registry
		}
		if name == "" {
			return img.ID[:16]
		}
		return reg + "/" + name + ":" + tag
	})
	colStyles := component.GetPageColumnStyles("image", colsDef)
	ts := tc.EffectiveTableStyle()
	return component.RenderSelectionBanner(imgRef, w) + "\n" +
		component.RenderTable(component.TableData{
			Cols:              colsDef,
			Widths:            widths,
			Rows:              rows,
			Selected:          selRow,
			Total:             total,
			Offset:            containerOffset,
			Limit:             rowHeight,
			HeaderOverrides:   overrides,
			FooterHint:        "l:logs Enter:logs Esc:back",
			BodyHeight:        panelHeight,
			RowPrefix:         ts.RowPrefix,
			RowPrefixSelected: ts.RowPrefixSelected,
			ColStyles:         colStyles,
			SelectionProvider: selProv,
		})
}

func selectedContainerSubRow(items []dockerclient.ContainerSummary, offset, cursor int) int {
	if cursor < offset {
		return 0
	}
	row := 0
	for i := offset; i < len(items) && i < cursor; i++ {
		row += len(containers.FormatPorts(items[i].PortBindings, false))
	}
	return row
}

func renderInfoBlock(lines []string, panelHeight int) string {
	if len(lines) == 0 {
		return ""
	}
	bodyHeight := panelHeight - 3
	if bodyHeight < len(lines) {
		bodyHeight = len(lines)
	}
	b := make([]string, 0, bodyHeight)
	for _, line := range lines {
		b = append(b, component.GetStyle("dim").Render("  "+line))
	}
	for len(b) < bodyHeight {
		b = append(b, "")
	}
	return strings.Join(b, "\n")
}

func resolveHdr(cd tables.ColumnDef) string {
	if cd.Header != "" && strings.HasPrefix(cd.Header, "table.") {
		return i18n.T(cd.Header)
	}
	return cd.Header
}

func fullRef(img *dockerclient.ImageSummary) string {
	if len(img.RepoTags) > 0 && img.RepoTags[0] != "<none>:<none>" {
		return img.RepoTags[0]
	}
	if len(img.ID) >= 12 {
		return img.ID[:12]
	}
	return img.ID
}

func sortArrow(asc bool) string {
	if asc {
		return " \u25b2"
	}
	return " \u25bc"
}

func precomputeContainerIDs(cm *state.ContainerListModel) []string {
	if cm == nil {
		return nil
	}
	out := make([]string, len(cm.Items))
	for i, c := range cm.Items {
		out[i] = c.Image
	}
	return out
}

func imgHasContainer(containerImages []string, imgID string, tags []string) bool {
	if containerImages == nil {
		return false
	}
	_, n, _ := splitRef(tags)
	for _, cImg := range containerImages {
		if strings.Contains(cImg, imgID[:12]) || strings.Contains(cImg, n) {
			return true
		}
	}
	return false
}

func hasContainers(cm *state.ContainerListModel, imgID string, tags []string) bool {
	containerImages := precomputeContainerIDs(cm)
	return imgHasContainer(containerImages, imgID, tags)
}

func splitRef(tags []string) (reg, name, tag string) {
	if len(tags) == 0 || tags[0] == "" {
		return component.StrDash, component.StrNone, component.StrNone
	}
	ref := tags[0]
	if ref == "<none>:<none>" {
		return component.StrDash, component.StrNone, component.StrNone
	}
	lc := strings.LastIndex(ref, ":")
	if lc > 0 {
		tag = ref[lc+1:]
		ref = ref[:lc]
	} else {
		tag = "latest"
	}
	parts := strings.Split(ref, "/")
	if len(parts) == 1 {
		return "docker.io", "library/" + parts[0], tag
	}
	if strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":") || parts[0] == "localhost" {
		return parts[0], strings.Join(parts[1:], "/"), tag
	}
	return "docker.io", strings.Join(parts, "/"), tag
}
