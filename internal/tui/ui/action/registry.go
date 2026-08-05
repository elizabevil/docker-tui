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
		{bindingLabel(model, keys.ActionEnter, keys.KEnter), i18n.T("key.sym_enter")},
		{bindingLabel(model, keys.ActionBack, keys.KEsc), i18n.T("key.back")},
		{bindingLabel(model, keys.ActionFilter, keys.KeySlash), i18n.T("key.filter")},
		{bindingLabel(model, keys.ActionClearFilters, keys.KeyCtrlI), i18n.T("key.clear_filters")},
		{bindingLabel(model, keys.ActionRefresh, keys.KeyR), i18n.T("key.refresh")},
		{bindingLabel(model, keys.ActionCommand, ":"), i18n.T("key.command")},
		{bindingLabel(model, keys.ActionActionBar, keys.KeySemicolon), i18n.T("key.action_bar")},
		{bindingLabel(model, keys.ActionSwitchRuntime, keys.KeyF2), i18n.T("key.runtime")},
		{bindingLabel(model, keys.ActionEvents, keys.KeyF3), i18n.T("events.title")},
		{bindingLabel(model, keys.ActionHelp, keys.KeyQmark), i18n.T("key.help")},
		{keys.KeyCUpper, i18n.T("key.connect")},
	}
}

func ForMode(app *state.AppModel) []Shortcut {
	if app == nil {
		return nil
	}
	switch app.Navigation.Mode {
	case state.ModeFilter:
		return shortcuts("Enter", i18n.T("key.keep"), "Esc Esc", i18n.T("key.clear_exit"), "Ctrl+A/E", i18n.T("key.home_end"), "Ctrl+W/U/K", i18n.T("key.edit"))
	case state.ModeSearch:
		return shortcuts("Enter", i18n.T("key.search"), "Esc", i18n.T("key.cancel"), "Ctrl+A/E", i18n.T("key.home_end"), "Ctrl+W/U/K", i18n.T("key.edit"))
	case state.ModeImagePull:
		return shortcuts("Enter", i18n.T("key.pull"), "Esc", i18n.T("key.cancel"), "Ctrl+A/E", i18n.T("key.home_end"), "Ctrl+W/U/K", i18n.T("key.edit"))
	case state.ModeImageWorkflow:
		return shortcuts("Enter", i18n.T("key.image_start"), "Esc", i18n.T("key.cancel"), "Ctrl+A/E", i18n.T("key.home_end"), "Ctrl+W/U/K", i18n.T("key.edit"))
	case state.ModeImageTransfer:
		return shortcuts("Esc", i18n.T("key.cancel"))
	case state.ModeCommand:
		return shortcuts("Enter", i18n.T("key.run"), "Tab", i18n.T("key.complete"), "Esc", i18n.T("key.exit"), "Ctrl+A/E", i18n.T("key.home_end"), "Ctrl+W/U/K", i18n.T("key.edit"))
	case state.ModeActionBar:
		return shortcuts("j/k", i18n.T("key.move"), "Enter", i18n.T("key.run"), "/", i18n.T("key.filter"), "1-9", i18n.T("key.jump"), "Esc", i18n.T("key.close"))
	case state.ModeHistory:
		if app.History.Filtering {
			return shortcuts("Enter/Esc", i18n.T("history.filter_done"), "Backspace", i18n.T("history.filter_delete"))
		}
		return shortcuts("Esc", i18n.T("key.back"), "j/k", i18n.T("key.scroll"), "PgUp/Dn", i18n.T("history.page"), "g/G", i18n.T("history.first_last"), "/", i18n.T("key.filter"))
	case state.ModeEvents:
		if app.EventPanel.Filtering {
			return shortcuts("Enter/Esc", i18n.T("events.filter_done"), "Backspace", i18n.T("history.filter_delete"))
		}
		return shortcuts("Esc", i18n.T("key.back"), "j/k", i18n.T("key.scroll"), "PgUp/Dn", i18n.T("history.page"), "Space", i18n.T("events.pause"), "/", i18n.T("key.filter"), "Ctrl+D", i18n.T("events.clear"))
	case state.ModeMark:
		return []Shortcut{
			{bindingLabel(app, keys.ActionBack, keys.KEsc), i18n.T("key.cancel")},
			{fmt.Sprintf("Space/%s", bindingLabel(app, keys.ActionEnter, keys.KEnter)), i18n.T("key.toggle_mark")},
			{bindingLabel(app, removeAction(app.Navigation.ActivePanel), keys.KCtrlD), i18n.T("key.delete_marked")},
		}
	case state.ModeConfirm:
		return shortcuts("Tab", i18n.T("key.tab_switch"), "Enter", i18n.T("key.confirm"), "Esc", i18n.T("key.cancel"))
	case state.ModeLogView:
		return shortcuts(bindingLabel(app, keys.ActionBack, keys.KEsc), i18n.T("key.back"), navigationLabel(app), i18n.T("key.scroll"), "PgUp/Dn", i18n.T("key.page_down"), "g/Ctrl+G", i18n.T("key.top_bot"), "n/Ctrl+N", i18n.T("key.match"), "w", i18n.T("key.wrap"))
	case state.ModeDetail:
		result := shortcuts(bindingLabel(app, keys.ActionBack, keys.KEsc)+"/"+bindingLabel(app, keys.ActionEnter, keys.KEnter), i18n.T("key.back"), navigationLabel(app), i18n.T("key.scroll"), "Space/PgDn", i18n.T("key.page_down"), "PgUp", i18n.T("key.page_up"), "g", i18n.T("key.detail_top"))
		if app.Detail.HasRawSource() {
			result = append(result, Shortcut{Key: keys.KeyS, Description: i18n.T("key.switch")})
		}
		return result
	case state.ModeTop:
		return shortcuts("Esc", i18n.T("key.back"), "j/k", i18n.T("key.move"), "r", i18n.T("key.refresh_short"))
	case state.ModeAuditDetail:
		return shortcuts("Esc/Enter", i18n.T("key.back"))
	case state.ModeRename:
		return shortcuts("Enter", i18n.T("key.rename"), "Esc", i18n.T("key.cancel"))
	case state.ModeResourceCreate:
		return shortcuts("Enter", i18n.T("key.create"), "Esc", i18n.T("key.cancel"))
	case state.ModeHelp:
		return shortcuts(bindingLabel(app, keys.ActionHelp, keys.KeyQmark)+"/"+bindingLabel(app, keys.ActionBack, keys.KEsc), i18n.T("key.close_help"))
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
	if app.Navigation.ActivePanel == state.PanelCompose {
		return Compose(app)
	}
	return ForPanel(app.Navigation.ActivePanel, len(app.Selection.MarkedIDs), app)
}

func Compose(app *state.AppModel) []Shortcut {
	if app == nil {
		return nil
	}
	if app.Compose.ComposeContainerViewID != "" {
		return []Shortcut{
			{keys.KEsc + "/" + keys.KLeft, i18n.T("key.back")},
			{keys.KUp + "/" + keys.KDown, i18n.T("key.scroll")},
		}
	}
	if app.Compose.ComposeFocus == 1 {
		return []Shortcut{
			{keys.KEnter, i18n.T("key.expand")},
			{keys.KLeft, i18n.T("key.projects")},
			{keys.KeyS, i18n.T("key.start")},
			{keys.KCtrlS, i18n.T("key.stop")},
			{keys.KeyL, i18n.T("key.logs")},
			{keys.KCtrlD, i18n.T("key.down")},
		}
	}
	return []Shortcut{
		{keys.KEnter + "/" + keys.KRight, i18n.T("key.services")},
		{keys.KeyD, i18n.T("key.detail")},
		{keys.KeyS, i18n.T("key.start")},
		{keys.KCtrlS, i18n.T("key.stop")},
		{keys.KeyL, i18n.T("key.logs")},
		{keys.KCtrlD, i18n.T("key.down")},
	}
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
				{bindingLabel(model, keys.ActionContainerPause, keys.KeyP), i18n.T("key.pause")},
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
			{bindingLabel(model, keys.ActionContainerInspect, keys.KeyI), i18n.T("key.inspect")},
			{bindingLabel(model, keys.ActionContainerStats, keys.KeyM), i18n.T("key.stats")},
			{bindingLabel(model, keys.ActionContainerPause, keys.KeyP), pauseLabel(model)},
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
			{bindingLabel(model, keys.ActionImageHistory, keys.KeyHUpper), i18n.T("history.title")},
			{keys.KCtrlB, i18n.T("key.debug")}, {bindingLabel(model, keys.ActionImageTag, keys.KeyCtrlT), i18n.T("key.tag")},
			{bindingLabel(model, keys.ActionImagePush, keys.KeyCtrlU), i18n.T("key.push")},
			{bindingLabel(model, keys.ActionImageSave, keys.KeyCtrlE), i18n.T("key.save")}, {bindingLabel(model, keys.ActionImageLoad, keys.KeyCtrlL), i18n.T("key.load")},
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
			{bindingLabel(model, createAction(panel), keys.KeyC), i18n.T("key.create")},
			{bindingLabel(model, pruneAction(panel), keys.KeyP), i18n.T("key.prune")},
			{bindingLabel(model, removeAction(panel), keys.KCtrlD), i18n.T("key.delete")},
		}
	case state.PanelCompose:
		return []Shortcut{
			{keys.KeyS, i18n.T("key.start")}, {keys.KCtrlS, i18n.T("key.stop")},
			{keys.KeyL, i18n.T("key.logs")}, {keys.KCtrlD, i18n.T("key.down")},
			{keys.KRight, i18n.T("key.detail")}, {keys.KEnter, i18n.T("key.expand")},
		}
	case state.PanelAudit:
		return []Shortcut{
			{keys.KEnter, i18n.T("key.audit_detail")}, {keys.KeyE, i18n.T("key.filter_level")},
			{keys.KeySlash, i18n.T("key.filter")},
		}
	default:
		return nil
	}
}

func createAction(panel state.PanelType) keys.KeyAction {
	if panel == state.PanelVolumes {
		return keys.ActionVolumeCreate
	}
	return keys.ActionNetworkCreate
}

func pruneAction(panel state.PanelType) keys.KeyAction {
	if panel == state.PanelVolumes {
		return keys.ActionVolumePrune
	}
	return keys.ActionNetworkPrune
}

func pauseLabel(model *state.AppModel) string {
	if model != nil && model.Resources.Containers != nil {
		if selected := model.Resources.Containers.Selected(); selected != nil && selected.State == state.ContainerStatePaused {
			return i18n.T("key.unpause")
		}
	}
	return i18n.T("key.pause")
}

func Sections(app ...*state.AppModel) []Section {
	model := firstModel(app)
	imageShortcuts := ForPanel(state.PanelImages, 0, model)
	imageShortcuts = append(imageShortcuts, Shortcut{Key: i18n.T("key.action_bar"), Description: i18n.T("history.title")})
	return []Section{
		{Title: i18n.T("help.global"), Shortcuts: Global(model)},
		{Title: i18n.T("help.containers"), Shortcuts: ForPanel(state.PanelContainers, 0, model)},
		{Title: i18n.T("help.images"), Shortcuts: imageShortcuts},
		{Title: i18n.T("compose.title"), Shortcuts: ForPanel(state.PanelCompose, 0, model)},
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
