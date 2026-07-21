package action

import (
	"fmt"
	"strings"

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

func Global(app ...*state.AppModel) []Shortcut {
	model := firstModel(app)
	return []Shortcut{
		{bindingLabel(model, keys.ActionDown, keys.KJDown), i18n.T("key.down")},
		{bindingLabel(model, keys.ActionUp, keys.KKUp), i18n.T("key.up")},
		{bindingLabel(model, keys.ActionTabNext, keys.KTab), i18n.T("key.panel")},
		{bindingLabel(model, keys.ActionEnter, keys.KEnter), "Enter"},
		{bindingLabel(model, keys.ActionBack, keys.KEsc), i18n.T("key.back")},
		{bindingLabel(model, keys.ActionFilter, keys.KeySlash), i18n.T("key.filter")},
		{bindingLabel(model, keys.ActionRefresh, keys.KeyR), "Refresh"},
		{bindingLabel(model, keys.ActionCommand, ":"), "Command"},
		{bindingLabel(model, keys.ActionSwitchRuntime, keys.KeyF2), "Runtime"},
		{bindingLabel(model, keys.ActionHelp, keys.KeyQmark), i18n.T("key.help")},
		{keys.KeyHUpper, i18n.T("key.header")},
		{keys.KeyCUpper, i18n.T("key.connect")},
		{bindingLabel(model, keys.ActionQuit, keys.KeyQ), i18n.T("key.quit")},
	}
}

func ForMode(app *state.AppModel) []Shortcut {
	if app == nil {
		return nil
	}
	switch app.Navigation.Mode {
	case state.ModeFilter:
		return shortcuts("Enter", "Keep", "Esc Esc", "Clear & exit", "Ctrl+A/E", "Home/End", "Ctrl+W/U/K", "Edit")
	case state.ModeSearch:
		return shortcuts("Enter", "Search", "Esc", "Cancel", "Ctrl+A/E", "Home/End", "Ctrl+W/U/K", "Edit")
	case state.ModeImagePull:
		return shortcuts("Enter", "Pull", "Esc", "Cancel", "Ctrl+A/E", "Home/End", "Ctrl+W/U/K", "Edit")
	case state.ModeCommand:
		return shortcuts("Enter", "Run", "Tab", "Complete", "Esc", "Exit", "Ctrl+A/E", "Home/End", "Ctrl+W/U/K", "Edit")
	case state.ModeMark:
		return []Shortcut{
			{bindingLabel(app, keys.ActionBack, keys.KEsc), "Cancel"},
			{fmt.Sprintf("Space/%s", bindingLabel(app, keys.ActionEnter, keys.KEnter)), "Toggle"},
			{bindingLabel(app, removeAction(app.Navigation.ActivePanel), keys.KCtrlD), "Delete marked"},
		}
	case state.ModeConfirm:
		return shortcuts("y", "Confirm", "n", "Cancel")
	case state.ModeLogView:
		return shortcuts(bindingLabel(app, keys.ActionBack, keys.KEsc), "Back", navigationLabel(app), "Scroll", "PgUp/Dn", "Page", "g/Ctrl+G", "Top/Bot", "n/Ctrl+N", "Match", "w", "Wrap")
	case state.ModeDetail:
		return shortcuts(bindingLabel(app, keys.ActionBack, keys.KEsc)+"/"+bindingLabel(app, keys.ActionEnter, keys.KEnter), "Back", navigationLabel(app), "Scroll", "Space/PgDn", "Page", "PgUp", "Page up", "g", "Top")
	case state.ModeHelp:
		return shortcuts(bindingLabel(app, keys.ActionHelp, keys.KeyQmark)+"/"+bindingLabel(app, keys.ActionBack, keys.KEsc), "Close")
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
	return ForPanel(app.Navigation.ActivePanel, len(app.Selection.MarkedIDs), app)
}

func ForPanel(panel state.PanelType, marked int, app ...*state.AppModel) []Shortcut {
	model := firstModel(app)
	hasMarked := marked > 0
	switch panel {
	case state.PanelContainers:
		if hasMarked {
			return []Shortcut{
				{fmt.Sprintf("%s [%d]", keys.KSpace, marked), i18n.T("key.toggle")},
				{bindingLabel(model, keys.ActionContainerStart, keys.KeyS), i18n.T("key.batch_start")},
				{bindingLabel(model, keys.ActionContainerStop, keys.KCtrlS), i18n.T("key.batch_stop")},
				{bindingLabel(model, keys.ActionContainerRestart, keys.KCtrlR), i18n.T("key.batch_restart")},
				{bindingLabel(model, keys.ActionContainerRemove, keys.KCtrlD), i18n.T("key.batch_delete")},
			}
		}
		return []Shortcut{
			{keys.KSpace, i18n.T("key.mark")}, {bindingLabel(model, keys.ActionContainerStart, keys.KeyS), i18n.T("key.start")},
			{bindingLabel(model, keys.ActionContainerStop, keys.KCtrlS), i18n.T("key.stop")},
			{bindingLabel(model, keys.ActionContainerRestart, keys.KCtrlR), i18n.T("key.restart")},
			{bindingLabel(model, keys.ActionContainerKill, keys.KCtrlK), i18n.T("key.kill")},
			{bindingLabel(model, keys.ActionContainerLogs, keys.KeyL), i18n.T("key.logs")},
			{bindingLabel(model, keys.ActionDetail, keys.KeyD), i18n.T("key.detail")},
			{bindingLabel(model, keys.ActionContainerExec, keys.KeyE), i18n.T("key.exec")},
			{bindingLabel(model, keys.ActionContainerInspect, keys.KeyI), "Inspect"},
			{bindingLabel(model, keys.ActionContainerStats, keys.KeyM), i18n.T("key.stats")},
			{bindingLabel(model, keys.ActionContainerRemove, keys.KCtrlD), i18n.T("key.delete")},
		}
	case state.PanelImages:
		if hasMarked {
			return []Shortcut{
				{fmt.Sprintf("%s [%d]", keys.KSpace, marked), i18n.T("key.toggle")},
				{bindingLabel(model, keys.ActionImagePull, keys.KCtrlP), i18n.T("key.batch_pull")},
				{bindingLabel(model, keys.ActionImagePrune, keys.KeyP), i18n.T("key.batch_prune")},
				{bindingLabel(model, keys.ActionImageRemove, keys.KCtrlD), i18n.T("key.batch_delete")},
			}
		}
		return []Shortcut{
			{keys.KSpace, i18n.T("key.mark")}, {keys.KRight, i18n.T("key.expand")},
			{keys.KeyY, i18n.T("key.copy")}, {bindingLabel(model, keys.ActionDetail, keys.KeyD), i18n.T("key.detail")},
			{keys.KCtrlB, i18n.T("key.debug")}, {keys.KCtrlE, i18n.T("key.export")},
			{bindingLabel(model, keys.ActionImagePull, keys.KCtrlP), i18n.T("key.pull")},
			{bindingLabel(model, keys.ActionImagePrune, keys.KeyP), i18n.T("key.prune")},
			{keys.KeyO, i18n.T("key.sort")}, {bindingLabel(model, keys.ActionImageRemove, keys.KCtrlD), i18n.T("key.delete")},
		}
	case state.PanelVolumes, state.PanelNetworks:
		if hasMarked {
			return []Shortcut{
				{fmt.Sprintf("%s [%d]", keys.KSpace, marked), i18n.T("key.toggle")},
				{bindingLabel(model, removeAction(panel), keys.KCtrlD), i18n.T("key.batch_delete")},
			}
		}
		return []Shortcut{
			{keys.KSpace, i18n.T("key.mark")}, {bindingLabel(model, keys.ActionEnter, keys.KEnter), i18n.T("key.expand")},
			{bindingLabel(model, removeAction(panel), keys.KCtrlD), i18n.T("key.delete")},
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

func Sections(app ...*state.AppModel) []Section {
	model := firstModel(app)
	return []Section{
		{Title: i18n.T("help.global"), Shortcuts: Global(model)},
		{Title: i18n.T("help.containers"), Shortcuts: ForPanel(state.PanelContainers, 0, model)},
		{Title: i18n.T("help.images"), Shortcuts: ForPanel(state.PanelImages, 0, model)},
		{Title: "Compose", Shortcuts: ForPanel(state.PanelCompose, 0, model)},
		{Title: i18n.T("help.volumes_networks"), Shortcuts: ForPanel(state.PanelVolumes, 0, model)},
	}
}

func bindingLabel(app *state.AppModel, action keys.KeyAction, fallback string) string {
	if app == nil || app.Dependencies.Config == nil {
		return fallback
	}
	bindings := keys.EffectiveKeys(app.Dependencies.Config.Keymap, action)
	if len(bindings) == 0 {
		return fallback
	}
	for index := range bindings {
		bindings[index] = displayKey(bindings[index])
	}
	return strings.Join(bindings, "/")
}

func navigationLabel(app *state.AppModel) string {
	return bindingLabel(app, keys.ActionDown, keys.KJDown) + "/" + bindingLabel(app, keys.ActionUp, keys.KKUp)
}

func displayKey(key string) string {
	parts := strings.Split(key, "+")
	for index, part := range parts {
		switch part {
		case string(keys.ModCtrl), string(keys.ModAlt), string(keys.ModShift):
			parts[index] = strings.ToUpper(part[:1]) + part[1:]
		case keys.KeyTab, keys.KeyEnter, keys.KeyEsc, keys.KeyHome, keys.KeyEnd, keys.KeyDelete:
			parts[index] = strings.ToUpper(part[:1]) + part[1:]
		default:
			if len(part) == 1 || (len(part) > 1 && part[0] == 'f') {
				parts[index] = strings.ToUpper(part)
			}
		}
	}
	return strings.Join(parts, "+")
}

func firstModel(models []*state.AppModel) *state.AppModel {
	if len(models) == 0 {
		return nil
	}
	return models[0]
}

func removeAction(panel state.PanelType) keys.KeyAction {
	switch panel {
	case state.PanelContainers:
		return keys.ActionContainerRemove
	case state.PanelImages:
		return keys.ActionImageRemove
	case state.PanelVolumes:
		return keys.ActionVolumeRemove
	case state.PanelNetworks:
		return keys.ActionNetworkRemove
	default:
		return keys.ActionDelete
	}
}

func shortcuts(values ...string) []Shortcut {
	result := make([]Shortcut, 0, len(values)/2)
	for i := 0; i+1 < len(values); i += 2 {
		result = append(result, Shortcut{Key: values[i], Description: values[i+1]})
	}
	return result
}
