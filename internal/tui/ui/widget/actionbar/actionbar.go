package actionbar

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	actionmodel "github.com/elizabevil/docker-tui/internal/tui/actionbar"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/widget/dialog"
)

const maxVisibleActions = 12

func RenderBar(m *state.AppModel, content string, body dialog.PanelBody) string {
	if m == nil || body.Width <= 0 || body.Rows <= 0 {
		return content
	}
	box := renderBox(actionmodel.VisibleItems(m), m.Navigation.ActionBar.Selected,
		m.Navigation.ActionBar.Filtering, m.Navigation.ActionBar.Filter, body, !m.CursorBlinkHidden)
	return dialog.PlaceDialogInPanel(content, box, body, dialog.LoadDialogConfig())
}

func renderBox(items []actionmodel.ActionItem, selected int, filtering bool, filter string, body dialog.PanelBody, cursorVisible ...bool) string {
	boxWidth := min(72, max(28, body.Width*2/3))
	boxWidth = min(boxWidth, max(8, body.Width-2))
	innerWidth := max(4, boxWidth-4)

	lines := []string{component.GetStyle("panelTitle").Render("Action Bar")}
	if filtering {
		cursor := "|"
		if len(cursorVisible) > 0 && !cursorVisible[0] {
			cursor = " "
		}
		lines = append(lines, component.TruncateVisible("/ "+filter+cursor, innerWidth))
	} else {
		lines = append(lines, component.GetStyle("dim").Render("/ filter   1-9 jump   Esc close"))
	}

	if len(items) == 0 {
		lines = append(lines, component.GetStyle("dim").Render("No matching actions"))
	} else {
		selected = max(0, min(selected, len(items)-1))
		availableRows := max(1, min(maxVisibleActions, body.Rows-6))
		start := max(0, selected-availableRows+1)
		end := min(len(items), start+availableRows)
		if end-start < availableRows {
			start = max(0, end-availableRows)
		}
		for index := start; index < end; index++ {
			item := items[index]
			marker := "  "
			if index == selected {
				marker = "> "
			}
			keyWidth := min(14, max(7, innerWidth/4))
			key := component.TruncateVisible(item.Key, keyWidth)
			labelWidth := max(1, innerWidth-keyWidth-3)
			line := fmt.Sprintf("%s%-*s %s", marker, keyWidth, key, component.TruncateVisible(item.Label, labelWidth))
			if item.Disabled {
				line = component.GetStyle("dim").Render(line)
			} else if index == selected {
				line = component.GetStyle("selectedRow").Width(innerWidth).Render(line)
			}
			lines = append(lines, line)
		}
		if len(items) > availableRows {
			lines = append(lines, component.GetStyle("dim").Render(fmt.Sprintf("%d-%d / %d", start+1, end, len(items))))
		}
	}

	return component.GetStyle("actionBar").
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(boxWidth - 4).
		Render(strings.Join(lines, "\n"))
}
