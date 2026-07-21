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
	containers := []Context{{View: "containers"}, {View: "image-containers"}}
	images := []Context{{View: "images"}}
	return []ActionSpec{
		{ActionQuit, []string{KeyQ, KeyCtrlC}, app},
		{ActionHelp, []string{KeyQmark, KeyF1}, app},
		{ActionFilter, []string{KeySlash}, main},
		{ActionRefresh, []string{KeyR}, main},
		{ActionTabNext, []string{KeyTab}, main},
		{ActionTabPrev, []string{KeyShiftTab}, main},
		{ActionUp, []string{KeyUp, KeyK}, main},
		{ActionDown, []string{KeyDown, KeyJ}, main},
		{ActionEnter, []string{KeyEnter}, main},
		{ActionBack, []string{KeyEsc}, main},
		{ActionDelete, []string{KeyCtrlD}, main},
		{ActionCommand, []string{":"}, app},
		{ActionSwitchRuntime, []string{KeyF2}, app},
		{ActionDetail, []string{KeyD}, []Context{{View: "containers"}, {View: "images"}, {View: "volumes"}, {View: "networks"}}},
		{ActionContainerStart, []string{KeyS}, containers},
		{ActionContainerStop, []string{KeyCtrlS}, containers},
		{ActionContainerRestart, []string{KeyCtrlR}, containers},
		{ActionContainerKill, []string{KeyCtrlK}, containers},
		{ActionContainerRemove, []string{KeyCtrlD}, []Context{{View: "containers"}}},
		{ActionContainerLogs, []string{KeyL}, containers},
		{ActionContainerExec, []string{KeyE}, containers},
		{ActionContainerInspect, []string{KeyI}, []Context{{View: "containers"}}},
		{ActionContainerStats, []string{KeyM}, []Context{{View: "containers"}}},
		{ActionContainerPause, []string{KeyP}, []Context{{View: "containers"}}},
		{ActionImagePull, []string{KeyCtrlP}, images},
		{ActionImageRemove, []string{KeyCtrlD}, images},
		{ActionImagePrune, []string{KeyP}, images},
		{ActionVolumeRemove, []string{KeyCtrlD}, []Context{{View: "volumes"}}},
		{ActionNetworkRemove, []string{KeyCtrlD}, []Context{{View: "networks"}}},
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
	return map[KeyAction][]string{
		ActionQuit: keymap.Quit, ActionHelp: keymap.Help, ActionFilter: keymap.Filter, ActionRefresh: keymap.Refresh,
		ActionContainerStart: keymap.ContainerStart, ActionContainerStop: keymap.ContainerStop,
		ActionContainerRestart: keymap.ContainerRestart, ActionContainerKill: keymap.ContainerKill,
		ActionContainerRemove: keymap.ContainerRemove, ActionContainerLogs: keymap.ContainerLogs,
		ActionContainerExec: keymap.ContainerExec, ActionContainerInspect: keymap.ContainerInspect,
		ActionContainerStats: keymap.ContainerStats, ActionContainerPause: keymap.ContainerPause, ActionImagePull: keymap.ImagePull,
		ActionImageRemove: keymap.ImageRemove, ActionImagePrune: keymap.ImagePrune,
		ActionVolumeRemove: keymap.VolumeRemove, ActionNetworkRemove: keymap.NetworkRemove,
		ActionTabNext: keymap.TabNext, ActionTabPrev: keymap.TabPrev, ActionUp: keymap.Up,
		ActionDown: keymap.Down, ActionEnter: keymap.Enter, ActionBack: keymap.Back, ActionDelete: keymap.Delete,
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
