package action

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

type Shortcut struct {
	Key         string
	Description string
}

type Section struct {
	Title     string
	Shortcuts []Shortcut
}

func Global() []Shortcut {
	return []Shortcut{
		{keys.KJDown, i18n.T("key.down")},
		{keys.KKUp, i18n.T("key.up")},
		{keys.KTab, i18n.T("key.panel")},
		{keys.KeySlash, i18n.T("key.filter")},
		{keys.KeyQmark, i18n.T("key.help")},
		{keys.KeyHUpper, i18n.T("key.header")},
		{keys.KeyCUpper, i18n.T("key.connect")},
		{keys.KeyQ, i18n.T("key.quit")},
	}
}

func ForMode(app *state.AppModel) []Shortcut {
	if app == nil {
		return nil
	}
	switch app.Mode {
	case state.ModeFilter:
		return shortcuts("Enter", "Apply", "Esc", "Exit")
	case state.ModeCommand:
		return shortcuts("Enter", "Run", "Tab", "Complete", "Esc", "Exit")
	case state.ModeMark:
		return shortcuts("Esc", "Cancel", "Space/Enter", "Toggle", "Ctrl+D", "Delete marked")
	case state.ModeConfirm:
		return shortcuts("y", "Confirm", "n", "Cancel")
	case state.ModeLogView:
		return shortcuts("Esc", "Back", "j/k", "Scroll", "PgUp/Dn", "Page", "g/G", "Top/Bot")
	case state.ModeDetail:
		return shortcuts("Esc/Enter", "Back", "j/k", "Scroll", "Space/PgDn", "Page")
	case state.ModeHelp:
		return shortcuts("?/Esc", "Close")
	default:
		return nil
	}
}

func Context(app *state.AppModel) []Shortcut {
	if mode := ForMode(app); len(mode) > 0 {
		return mode
	}
	if app == nil {
		return nil
	}
	return ForPanel(app.ActivePanel, len(app.MarkedIDs))
}

func ForPanel(panel state.PanelType, marked int) []Shortcut {
	hasMarked := marked > 0
	switch panel {
	case state.PanelContainers:
		if hasMarked {
			return []Shortcut{
				{fmt.Sprintf("%s [%d]", keys.KSpace, marked), i18n.T("key.toggle")},
				{keys.KeyS, i18n.T("key.batch_start")}, {keys.KCtrlS, i18n.T("key.batch_stop")},
				{keys.KCtrlR, i18n.T("key.batch_restart")}, {keys.KCtrlD, i18n.T("key.batch_delete")},
			}
		}
		return []Shortcut{
			{keys.KSpace, i18n.T("key.mark")}, {keys.KeyS, i18n.T("key.start")},
			{keys.KCtrlS, i18n.T("key.stop")}, {keys.KCtrlR, i18n.T("key.restart")},
			{keys.KeyL, i18n.T("key.logs")}, {keys.KeyD, i18n.T("key.detail")},
			{keys.KCtrlD, i18n.T("key.delete")},
		}
	case state.PanelImages:
		if hasMarked {
			return []Shortcut{
				{fmt.Sprintf("%s [%d]", keys.KSpace, marked), i18n.T("key.toggle")},
				{keys.KCtrlP, i18n.T("key.batch_pull")}, {keys.KeyP, i18n.T("key.batch_prune")},
				{keys.KCtrlD, i18n.T("key.batch_delete")},
			}
		}
		return []Shortcut{
			{keys.KSpace, i18n.T("key.mark")}, {keys.KRight, i18n.T("key.expand")},
			{keys.KeyY, i18n.T("key.copy")}, {keys.KeyD, i18n.T("key.detail")},
			{keys.KCtrlB, i18n.T("key.debug")}, {keys.KCtrlE, i18n.T("key.export")},
			{keys.KCtrlP, i18n.T("key.pull")}, {keys.KeyP, i18n.T("key.prune")},
			{keys.KeyO, i18n.T("key.sort")}, {keys.KCtrlD, i18n.T("key.delete")},
		}
	case state.PanelVolumes, state.PanelNetworks:
		if hasMarked {
			return []Shortcut{
				{fmt.Sprintf("%s [%d]", keys.KSpace, marked), i18n.T("key.toggle")},
				{keys.KCtrlD, i18n.T("key.batch_delete")},
			}
		}
		return []Shortcut{
			{keys.KSpace, i18n.T("key.mark")}, {keys.KEnter, i18n.T("key.expand")},
			{keys.KCtrlD, i18n.T("key.delete")},
		}
	case state.PanelCompose:
		return []Shortcut{
			{keys.KeyS, i18n.T("key.start")}, {keys.KCtrlS, i18n.T("key.stop")},
			{keys.KeyL, i18n.T("key.logs")}, {keys.KCtrlD, i18n.T("key.down")},
			{keys.KRight, i18n.T("key.detail")}, {keys.KEnter, i18n.T("key.expand")},
		}
	default:
		return nil
	}
}

func Sections() []Section {
	return []Section{
		{Title: i18n.T("help.global"), Shortcuts: Global()},
		{Title: i18n.T("help.containers"), Shortcuts: ForPanel(state.PanelContainers, 0)},
		{Title: i18n.T("help.images"), Shortcuts: ForPanel(state.PanelImages, 0)},
		{Title: "Compose", Shortcuts: ForPanel(state.PanelCompose, 0)},
		{Title: i18n.T("help.volumes_networks"), Shortcuts: ForPanel(state.PanelVolumes, 0)},
	}
}

func shortcuts(values ...string) []Shortcut {
	result := make([]Shortcut, 0, len(values)/2)
	for i := 0; i+1 < len(values); i += 2 {
		result = append(result, Shortcut{Key: values[i], Description: values[i+1]})
	}
	return result
}
