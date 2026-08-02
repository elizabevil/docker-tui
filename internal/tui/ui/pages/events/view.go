// Package events renders the BR-035 Events panel (F3).
package events

import (
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	dockerruntime "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// RenderView renders the runtime events panel. It builds only the rows visible
// in the viewport and delegates frame/table layout to the shared table
// component (FlexLayout-driven).
func RenderView(ev *state.EventsPanelState, panelHeight, panelWidth int) string {
	tc := tables.MustLoad("events")
	cols := tc.Columns["default"]
	if panelWidth < 30 {
		panelWidth = 30
	}

	items := ev.FilteredEvents()

	// Reserve header/footer lines; clamp the offset and window just the
	// visible slice so only visible rows are constructed. ClampOffset also
	// corrects the state's ViewOffset/Cursor so navigation stays consistent
	// when the filtered result shrinks.
	visible := max(panelHeight-4, 1)
	ev.ClampOffset(len(items), visible)
	offset := ev.ViewOffset
	end := min(offset+visible, len(items))
	window := items[offset:end]

	selectedWindow := -1
	if ev.Cursor >= offset && ev.Cursor < end {
		selectedWindow = ev.Cursor - offset
	}

	rows := make([][]string, 0, len(window))
	for _, it := range window {
		rows = append(rows, eventRow(it))
	}

	return component.RenderTable(component.TableData{
		Cols:       cols,
		Rows:       rows,
		Selected:   selectedWindow,
		Total:      len(items),
		Offset:     offset,
		Limit:      panelHeight - 2,
		BodyHeight: panelHeight,
		BannerW:    panelWidth - 4,
		TopLabel:   panelTitle(ev),
		FooterHint: footerHint(ev),
	})
}

// panelTitle renders the always-visible top frame label, carrying the paused
// and filter state even when the buffer is empty (the page-info footer line is
// only drawn by RenderTable when Total > 0).
func panelTitle(ev *state.EventsPanelState) string {
	label := i18n.T("events.title")
	if ev.Paused {
		label += " (" + i18n.T("events.paused") + ")"
	}
	if ev.Filter != "" {
		label += " · " + i18n.T("history.filter_label") + ": " + ev.Filter
	}
	return label
}

func footerHint(ev *state.EventsPanelState) string {
	status := i18n.T("events.live")
	if ev.Paused {
		status = i18n.T("events.paused")
	}
	if ev.Filter != "" {
		status += " · " + i18n.T("history.filter_label") + ": " + ev.Filter
	}
	return status
}

func eventRow(ev dockerruntime.Event) []string {
	resource := ev.ActorID
	if name := ev.Attributes["name"]; name != "" {
		resource = name
	}
	if resource == "" {
		resource = "—"
	}
	created := ""
	if ev.Time > 0 {
		created = utils.FormatCreated(ev.Time)
	}
	return []string{created, ev.ResourceType, ev.Action, resource}
}
