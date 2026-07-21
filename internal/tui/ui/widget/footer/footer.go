package footer

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/action"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func StatusBar(app *state.AppModel) string {
	if app.Navigation.Mode == state.ModeMark {
		status := fmt.Sprintf("%s Mark mode", component.GetStyle("toastWarning").Render("●"))
		return component.GetStyle("statusBar").Render(status)
	}
	engineLabel := app.Connection.RuntimeType
	if engineLabel == "" {
		engineLabel = app.Connection.ConnectionTarget
	}
	if engineLabel == "" {
		engineLabel = "no runtime"
	}
	hostStr := ""
	if app.Connection.Docker != nil {
		hostStr = app.Connection.Docker.Host
	}
	status := fmt.Sprintf("%s %s", component.GetStyle("toastSuccess").Render("●"), engineLabel)
	if app.Connection.Connecting {
		status = fmt.Sprintf("%s connecting...", component.GetStyle("toastWarning").Render("○"))
	} else if !app.Connection.Connected {
		status = fmt.Sprintf("%s disconnected (%s)", component.GetStyle("toastError").Render("○"), engineLabel)
	}
	if hostStr != "" {
		status += " │ " + hostStr
	}
	if op := operationLogStatus(app); op != "" {
		status += " │ " + op
	}
	return component.GetStyle("statusBar").Render(status)
}

func Shortcuts(app *state.AppModel) string {
	globalRow := renderShortcuts(action.Global(app))
	contextRow := renderShortcuts(action.Context(app))
	return lipgloss.JoinVertical(lipgloss.Top, globalRow, contextRow)
}

// Render returns the fixed three-row footer rail: global actions, context
// actions, and runtime/operation status.
func Render(app *state.AppModel, width int) string {
	rows := strings.Split(Shortcuts(app), "\n")
	for len(rows) < 2 {
		rows = append(rows, "")
	}
	rows = rows[:2]
	rows = append(rows, StatusBar(app))
	for i := range rows {
		rows[i] = component.PadVisible(component.TruncateVisible(rows[i], width), width)
	}
	return strings.Join(rows, "\n")
}

func renderShortcuts(shortcuts []action.Shortcut) string {
	if len(shortcuts) == 0 {
		return ""
	}
	var parts []string
	for _, shortcut := range shortcuts {
		parts = append(parts, component.GetStyle("hintKey").Render(shortcut.Key)+component.GetStyle("hintDesc").Render(" "+shortcut.Description))
	}
	return component.GetStyle("shortcutBar").Render(strings.Join(parts, component.GetStyle("hintSep").Render(" │ ")))
}
