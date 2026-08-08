package footer

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/action"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func Shortcuts(app *state.AppModel) string {
	globalRow := renderShortcuts(action.Global(app))
	contextRow := renderShortcuts(action.Context(app))
	return lipgloss.JoinVertical(lipgloss.Top, globalRow, contextRow)
}

// Render returns the fixed two-row footer rail: global actions and
// context actions. The query input (when active) is rendered in the
// layout layer, not here, so this package stays decoupled from
// the input component.
func Render(app *state.AppModel, width int) string {
	rows := strings.Split(Shortcuts(app), "\n")
	for len(rows) < 2 {
		rows = append(rows, "")
	}
	rows = rows[:2]
	for i := range rows {
		row := component.PadVisible(component.TruncateVisible(rows[i], width), width)
		rows[i] = component.GetStyle(component.StyleShortcutBar).Render(row)
	}
	return strings.Join(rows, "\n")
}

func renderShortcuts(shortcuts []action.Shortcut) string {
	if len(shortcuts) == 0 {
		return ""
	}
	var parts []string
	for _, shortcut := range shortcuts {
		parts = append(parts, component.GetStyle(component.StyleHintKey).Render(shortcut.Key)+component.GetStyle(component.StyleHintDescription).Render(" "+shortcut.Description))
	}
	return strings.Join(parts, component.GetStyle(component.StyleHintSeparator).Render(" "+component.BorderLineVertical+" "))
}
