package keys

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

type Context struct {
	App     string
	Surface string
	View    string
	Mode    string
}

type ActionSpec struct {
	Action      KeyAction
	DefaultKeys []string
	Contexts    []Context
}

type BindingSet struct {
	ByKey    map[string][]KeyAction
	ByAction map[KeyAction][]string
}

type Resolver struct {
	bindings BindingSet
	specs    map[KeyAction]ActionSpec
}

func Registry() []ActionSpec {
	app := []Context{{App: "app"}}
	main := []Context{{Surface: "main"}}
	containers := []Context{{View: "containers"}, {View: "image-containers"}, {View: "compose-containers"}}
	images := []Context{{View: "images"}}
	compose := []Context{{View: "compose"}}
	return []ActionSpec{
		{ActionQuit, []string{KeyCtrlC}, app},
		{ActionHelp, []string{KeyQmark, KeyF1}, app},
		{ActionFilter, []string{KeySlash}, main},
		{ActionRefresh, []string{KeyR}, main},
		{ActionActionBar, []string{KeySemicolon}, main},
		{ActionTabNext, []string{KeyTab, KeyCloseBracket}, main},
		{ActionTabPrev, []string{KeyShiftTab, KeyOpenBracket}, main},
		{ActionUp, []string{KeyUp, KeyK}, main},
		{ActionDown, []string{KeyDown, KeyJ}, main},
		{ActionEnter, []string{KeyEnter}, main},
		{ActionBack, []string{KeyEsc}, main},
		{ActionDelete, []string{KeyCtrlD}, main},
		{ActionCommand, []string{":"}, app},
		{ActionSwitchRuntime, []string{KeyF2}, app},
		{ActionEvents, []string{KeyF3}, app},
		{ActionDetail, []string{KeyD}, []Context{{View: "containers"}, {View: "images"}, {View: "volumes"}, {View: "networks"}, {View: "audit"}}},
		{ActionContainerStart, []string{KeyS}, containers},
		{ActionContainerStop, []string{KeyCtrlS}, containers},
		{ActionContainerRestart, []string{KeyCtrlR}, containers},
		{ActionContainerKill, []string{KeyCtrlK}, containers},
		{ActionContainerRemove, []string{KeyCtrlD}, []Context{{View: "containers"}}},
		{ActionContainerLogs, []string{KeyL}, containers},
		{ActionContainerExec, []string{KeyE}, containers},
		{ActionContainerInspect, []string{KeyI}, []Context{{View: "containers"}}},
		{ActionContainerStats, []string{KeyM}, []Context{{View: "containers"}}},
		{ActionContainerPause, []string{KeyP}, containers},

		// TASK-019 advanced container actions. Keys chosen to avoid
		// collisions with the existing lifecycle bindings above; users can
		// override via the keymap config.
		{ActionContainerUpdate, nil, containers},
		{ActionContainerDiff, nil, containers},
		{ActionContainerExport, nil, containers},
		{ActionContainerCommit, nil, containers},
		{ActionContainerWait, nil, containers},
		{ActionContainerCopy, nil, containers},
		{ActionImagePull, []string{KeyCtrlP}, images},
		{ActionImageRemove, []string{KeyCtrlD}, images},
		{ActionImagePrune, []string{KeyP}, images},
		{ActionImageTag, []string{KeyCtrlT}, images},
		{ActionImagePush, []string{KeyCtrlU}, images},
		{ActionImageSave, []string{KeyCtrlE}, images},
		{ActionImageLoad, []string{KeyCtrlL}, images},
		{ActionImageHistory, []string{KeyH}, images},
		{ActionVolumeCreate, []string{KeyC}, []Context{{View: "volumes"}}},
		{ActionVolumePrune, []string{KeyP}, []Context{{View: "volumes"}}},
		{ActionVolumeRemove, []string{KeyCtrlD}, []Context{{View: "volumes"}}},
		{ActionNetworkCreate, []string{KeyC}, []Context{{View: "networks"}}},
		{ActionNetworkPrune, []string{KeyP}, []Context{{View: "networks"}}},
		{ActionNetworkRemove, []string{KeyCtrlD}, []Context{{View: "networks"}}},

		// R08-12: Compose project-level actions.
		{ActionComposeProjectStart, []string{KeyS}, compose},
		{ActionComposeProjectStop, []string{KeyCtrlS}, compose},
		{ActionComposeProjectRestart, []string{KeyCtrlR}, compose},
		{ActionComposeProjectDown, []string{KeyCtrlD}, compose},
		{ActionComposeProjectLogs, []string{KeyL}, compose},
		{ActionComposeProjectTop, []string{KeyCtrlT}, compose},
		{ActionComposeProjectPort, []string{KeyComma}, compose},
		{ActionComposeProjectStats, nil, compose},
		{ActionComposeProjectScale, nil, compose},
		{ActionComposeProjectPause, nil, compose},
		{ActionComposeProjectUnpause, nil, compose},
		{ActionComposeProjectKill, []string{KeyCtrlK}, compose},
		{ActionComposeProjectRm, nil, compose},
		{ActionComposeProjectPrune, nil, compose},
		{ActionComposeProjectRun, nil, compose},
		{ActionComposeProjectEvents, []string{KeyCtrlF3}, compose},
		{ActionComposeProjectDetail, []string{KeyD}, compose},

		// R08-12: Compose service-level actions.
		{ActionComposeServiceRun, nil, compose},
		{ActionComposeServiceExec, []string{KeyE}, compose},
		{ActionComposeServiceLogs, nil, compose},

		{ActionRefreshConnections, []string{KeyF12, KeyR}, app},
		{ActionClearFilters, []string{KeyCtrlI}, app},
	}
}

func CompileBindings(overrides config.KeymapConfig) BindingSet {
	set := BindingSet{ByKey: make(map[string][]KeyAction), ByAction: make(map[KeyAction][]string)}
	configured := configuredBindings(overrides)
	for _, spec := range Registry() {
		bindings := spec.DefaultKeys
		if override := configured[spec.Action]; len(override) > 0 {
			bindings = override
		}
		for _, binding := range bindings {
			key := Normalize(binding)
			if key == "" || containsAction(set.ByKey[key], spec.Action) {
				continue
			}
			set.ByKey[key] = append(set.ByKey[key], spec.Action)
			set.ByAction[spec.Action] = append(set.ByAction[spec.Action], key)
		}
	}
	return set
}

func NewResolver(bindings BindingSet) *Resolver {
	specs := make(map[KeyAction]ActionSpec)
	for _, spec := range Registry() {
		specs[spec.Action] = spec
	}
	return &Resolver{bindings: bindings, specs: specs}
}

func (r *Resolver) Resolve(key string, context Context) (KeyAction, bool) {
	if r == nil {
		return "", false
	}
	candidates := r.bindings.ByKey[Normalize(key)]
	bestScore := -1
	var best KeyAction
	for _, candidate := range candidates {
		spec, ok := r.specs[candidate]
		if !ok {
			continue
		}
		if candidate == ActionDelete && r.resourceDeleteOverrides(context) {
			continue
		}
		if score, matches := contextScore(spec.Contexts, context); matches && score > bestScore {
			best, bestScore = candidate, score
		}
	}
	return best, bestScore >= 0
}

// resourceDeleteOverrides keeps the legacy generic delete action as a
// fallback without letting it preserve an old key after a resource-specific
// delete binding has been overridden.
func (r *Resolver) resourceDeleteOverrides(context Context) bool {
	var action KeyAction
	switch context.View {
	case "containers":
		action = ActionContainerRemove
	case "images":
		action = ActionImageRemove
	case "volumes":
		action = ActionVolumeRemove
	case "networks":
		action = ActionNetworkRemove
	case "compose", "compose-containers":
		action = ActionComposeProjectDown
	default:
		return false
	}
	return len(r.bindings.ByAction[action]) > 0
}

func EffectiveKeys(overrides config.KeymapConfig, action KeyAction) []string {
	return append([]string(nil), CompileBindings(overrides).ByAction[action]...)
}

func Normalize(key string) string {
	if key == KeySpace {
		return KeySpace
	}
	return strings.ToLower(strings.TrimSpace(key))
}

func configuredBindings(keymap config.KeymapConfig) map[KeyAction][]string {
	cm := keymap.Compose
	return map[KeyAction][]string{
		ActionQuit: keymap.Global.Quit.Values(), ActionActionBar: keymap.Global.ActionBar.Values(), ActionHelp: keymap.Global.Help.Values(), ActionFilter: keymap.Global.Filter.Values(), ActionRefresh: keymap.Global.Refresh.Values(),
		ActionContainerStart: keymap.Container.Start.Values(), ActionContainerStop: keymap.Container.Stop.Values(),
		ActionContainerRestart: keymap.Container.Restart.Values(), ActionContainerKill: keymap.Container.Kill.Values(),
		ActionContainerRemove: keymap.Container.Remove.Values(), ActionContainerLogs: keymap.Container.Logs.Values(),
		ActionContainerExec: keymap.Container.Exec.Values(), ActionContainerInspect: keymap.Container.Inspect.Values(),
		ActionContainerStats: keymap.Container.Stats.Values(), ActionContainerPause: keymap.Container.Pause.Values(),
		// TASK-019 advanced container actions.
		ActionContainerUpdate: keymap.Container.Update.Values(), ActionContainerDiff: keymap.Container.Diff.Values(),
		ActionContainerExport: keymap.Container.Export.Values(), ActionContainerCommit: keymap.Container.Commit.Values(),
		ActionContainerWait: keymap.Container.Wait.Values(), ActionContainerCopy: keymap.Container.Copy.Values(),
		ActionImagePull:   keymap.Image.Pull.Values(),
		ActionImageRemove: keymap.Image.Remove.Values(), ActionImagePrune: keymap.Image.Prune.Values(),
		ActionImageTag: keymap.Image.Tag.Values(), ActionImagePush: keymap.Image.Push.Values(), ActionImageSave: keymap.Image.Save.Values(), ActionImageLoad: keymap.Image.Load.Values(), ActionImageHistory: keymap.Image.History.Values(),
		ActionVolumeCreate: keymap.Volume.Create.Values(), ActionVolumePrune: keymap.Volume.Prune.Values(), ActionVolumeRemove: keymap.Volume.Remove.Values(),
		ActionNetworkCreate: keymap.Network.Create.Values(), ActionNetworkPrune: keymap.Network.Prune.Values(), ActionNetworkRemove: keymap.Network.Remove.Values(),
		ActionTabNext: keymap.Navigation.TabNext.Values(), ActionTabPrev: keymap.Navigation.TabPrev.Values(), ActionUp: keymap.Navigation.Up.Values(),
		ActionDown: keymap.Navigation.Down.Values(), ActionEnter: keymap.Navigation.Enter.Values(), ActionBack: keymap.Navigation.Back.Values(), ActionDelete: keymap.Navigation.Delete.Values(),

		// R08-12: Compose project-level actions.
		ActionComposeProjectStart:   cm.Start.Values(),
		ActionComposeProjectStop:    cm.Stop.Values(),
		ActionComposeProjectRestart: cm.Restart.Values(),
		ActionComposeProjectDown:    cm.Down.Values(),
		ActionComposeProjectLogs:    cm.Logs.Values(),
		ActionComposeProjectTop:     cm.Top.Values(),
		ActionComposeProjectPort:    cm.Port.Values(),
		ActionComposeProjectStats:   cm.Stats.Values(),
		ActionComposeProjectScale:   cm.Scale.Values(),
		ActionComposeProjectPause:   cm.Pause.Values(),
		ActionComposeProjectUnpause: cm.Unpause.Values(),
		ActionComposeProjectKill:    cm.Kill.Values(),
		ActionComposeProjectRm:      cm.Rm.Values(),
		ActionComposeProjectPrune:   cm.Prune.Values(),
		ActionComposeProjectRun:      cm.Run.Values(),
		ActionComposeProjectEvents:  cm.Events.Values(),
		ActionComposeProjectDetail:  cm.Detail.Values(),

		// R08-12: Compose service-level actions.
		ActionComposeServiceRun:  cm.Run.Values(),
		ActionComposeServiceExec: cm.Exec.Values(),
		ActionComposeServiceLogs: cm.ServiceLogs.Values(),
	}
}

func contextScore(patterns []Context, actual Context) (int, bool) {
	best := -1
	for _, pattern := range patterns {
		if pattern.App != "" && pattern.App != actual.App ||
			pattern.Surface != "" && pattern.Surface != actual.Surface ||
			pattern.View != "" && pattern.View != actual.View ||
			pattern.Mode != "" && pattern.Mode != actual.Mode {
			continue
		}
		score := 0
		if pattern.App != "" {
			score += 1
		}
		if pattern.Surface != "" {
			score += 10
		}
		if pattern.View != "" {
			score += 100
		}
		if pattern.Mode != "" {
			score += 1000
		}
		if score > best {
			best = score
		}
	}
	return best, best >= 0
}

func containsAction(actions []KeyAction, target KeyAction) bool {
	for _, action := range actions {
		if action == target {
			return true
		}
	}
	return false
}
