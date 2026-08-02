package history

import (
	"fmt"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"
)

func RenderView(m *state.AppModel, panelHeight, panelWidth int) string {
	h := &m.History
	tc := tables.MustLoad("history")
	cols := tc.Columns["default"]
	if panelWidth < 30 {
		panelWidth = 30
	}

	if content := renderStatus(h); content != "" {
		return content
	}

	items := FilterLayers(h.Layers, h.Filter)

	visible := panelHeight - 4
	if visible < 1 {
		visible = 1
	}
	if h.ViewOffset < 0 {
		h.ViewOffset = 0
	}
	if h.ViewOffset > max(0, len(items)-visible) {
		h.ViewOffset = max(0, len(items)-visible)
	}
	end := h.ViewOffset + visible
	if end > len(items) {
		end = len(items)
	}
	window := items[h.ViewOffset:end]

	selectedWindow := -1
	if h.Cursor >= h.ViewOffset && h.Cursor < end {
		selectedWindow = h.Cursor - h.ViewOffset
	}

	rows := make([][]string, 0, len(window))
	for _, layer := range window {
		rows = append(rows, layerRow(layer))
	}

	footer := fmt.Sprintf(i18n.T("history.range"), h.ViewOffset+1, h.ViewOffset+len(window), len(items))
	if h.Filter != "" {
		footer += "   " + i18n.T("history.filter_label") + ": " + h.Filter
	}
	if h.Filtering {
		footer += "   " + i18n.T("history.filter_editing")
	}

	return component.RenderTable(component.TableData{
		Cols:       cols,
		Rows:       rows,
		Selected:   selectedWindow,
		Total:      len(items),
		Offset:     h.ViewOffset,
		Limit:      panelHeight - 2,
		BodyHeight: panelHeight,
		BannerW:    panelWidth - 4,
		FooterHint: footer,
	})
}

func renderStatus(h *state.HistoryState) string {
	if h.ImageID == "" {
		return i18n.T("history.no_selection")
	}
	if h.Loading {
		return i18n.T("history.loading")
	}
	if h.Error != "" {
		return i18n.T("history.error") + ": " + h.Error
	}
	if h.Source == dockerclient.ImageHistoryManifest {
		return i18n.T("history.manifest_notice")
	}
	if len(h.Layers) == 0 {
		return i18n.T("history.empty")
	}
	return ""
}

// FilterLayers returns the history rows matching CreatedBy or Comment.
func FilterLayers(layers []dockerclient.ImageHistoryLayer, filter string) []dockerclient.ImageHistoryLayer {
	filter = strings.ToLower(strings.TrimSpace(filter))
	if filter == "" {
		return layers
	}
	out := make([]dockerclient.ImageHistoryLayer, 0, len(layers))
	for _, l := range layers {
		if strings.Contains(strings.ToLower(l.CreatedBy), filter) ||
			strings.Contains(strings.ToLower(l.Comment), filter) {
			out = append(out, l)
		}
	}
	return out
}

func layerRow(l dockerclient.ImageHistoryLayer) []string {
	id := l.ID
	if len(id) > 14 {
		id = id[:12] + "…"
	}
	created := ""
	if l.Created > 0 {
		created = utils.FormatCreated(l.Created)
	}
	return []string{id, created, utils.FormatSize(l.Size), l.CreatedBy, l.Comment}
}
