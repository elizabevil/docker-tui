package keys

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

func TestCompileBindingsAppliesOverridesAndFallbacks(t *testing.T) {
	bindings := CompileBindings(config.KeymapConfig{Global: config.GlobalKeymap{Help: config.KeyBinding{Primary: "f3"}}})
	if got := bindings.ByAction[ActionHelp]; len(got) != 1 || got[0] != "f3" {
		t.Fatalf("help override = %#v", got)
	}
	if got := bindings.ByAction[ActionQuit]; len(got) == 0 || got[0] != KeyCtrlC {
		t.Fatalf("quit fallback = %#v", got)
	}
	if got := bindings.ByAction[ActionActionBar]; len(got) == 0 || got[0] != KeySemicolon {
		t.Fatalf("action bar fallback = %#v", got)
	}
}

func TestResolverUsesViewPriorityForConflictingDeleteBinding(t *testing.T) {
	resolver := NewResolver(CompileBindings(config.KeymapConfig{}))
	action, ok := resolver.Resolve(KeyCtrlD, Context{App: "app", Surface: "main", View: "images", Mode: "normal"})
	if !ok || action != ActionImageRemove {
		t.Fatalf("Resolve(ctrl+d, images) = %q, %v", action, ok)
	}
}

func TestImageHistoryUsesHOnlyInImages(t *testing.T) {
	resolver := NewResolver(CompileBindings(config.KeymapConfig{}))
	images := Context{App: "app", Surface: "main", View: "images", Mode: "normal"}
	if action, ok := resolver.Resolve(KeyH, images); !ok || action != ActionImageHistory {
		t.Fatalf("Resolve(h, images) = %q, %v", action, ok)
	}
	containers := Context{App: "app", Surface: "main", View: "containers", Mode: "normal"}
	if action, ok := resolver.Resolve(KeyH, containers); ok {
		t.Fatalf("Resolve(h, containers) unexpectedly resolved %q", action)
	}
}

func TestResourceDeleteOverrideDisablesLegacyGenericBinding(t *testing.T) {
	resolver := NewResolver(CompileBindings(config.KeymapConfig{Image: config.ImageKeymap{Remove: config.KeyBinding{Primary: "z"}}}))
	context := Context{App: "app", Surface: "main", View: "images", Mode: "normal"}

	if action, ok := resolver.Resolve(KeyCtrlD, context); ok {
		t.Fatalf("legacy delete binding resolved after resource override: %q", action)
	}
	if action, ok := resolver.Resolve("z", context); !ok || action != ActionImageRemove {
		t.Fatalf("resource override = %q, %v", action, ok)
	}
}

func TestResolverRejectsActionOutsideContext(t *testing.T) {
	resolver := NewResolver(CompileBindings(config.KeymapConfig{Container: config.ContainerKeymap{Start: config.KeyBinding{Primary: "z"}}}))
	if action, ok := resolver.Resolve("z", Context{App: "app", Surface: "main", View: "images", Mode: "normal"}); ok {
		t.Fatalf("resolved out-of-context action %q", action)
	}
}

func TestDefaultAppConfigMatchesRegistryDefaults(t *testing.T) {
	configured := CompileBindings(config.DefaultAppConfig().Keymap)
	defaults := CompileBindings(config.KeymapConfig{})
	for _, spec := range Registry() {
		got := configured.ByAction[spec.Action]
		want := defaults.ByAction[spec.Action]
		if len(got) != len(want) {
			t.Fatalf("%s bindings differ: got=%v want=%v", spec.Action, got, want)
		}
		for index := range want {
			if got[index] != want[index] {
				t.Fatalf("%s bindings differ: got=%v want=%v", spec.Action, got, want)
			}
		}
	}
}

// TestComposeCtrlDRoutesToComposeProjectDown verifies R08-12 F1 —
// Ctrl+D in the compose view resolves to the project down action and
// the legacy ActionDelete does NOT shadow it.
func TestComposeCtrlDRoutesToComposeProjectDown(t *testing.T) {
	resolver := NewResolver(CompileBindings(config.DefaultAppConfig().Keymap))
	compose := Context{App: "app", Surface: "main", View: "compose", Mode: "normal"}

	action, ok := resolver.Resolve(KeyCtrlD, compose)
	if !ok || action != ActionComposeProjectDown {
		t.Fatalf("Resolve(ctrl+d, compose) = %q, %v; want %q", action, ok, ActionComposeProjectDown)
	}

	if _, ok := resolver.Resolve(KeyCtrlD, compose); ok && action == ActionDelete {
		t.Fatalf("ActionDelete leaked into compose view after override")
	}
}

// TestComposeContainerActionsReachableInSubview verifies R08-12 F2 —
// when the user is inside compose-containers, container-scoped actions
// still resolve to the same handlers as the regular containers view.
func TestComposeContainerActionsReachableInSubview(t *testing.T) {
	resolver := NewResolver(CompileBindings(config.DefaultAppConfig().Keymap))
	subview := Context{App: "app", Surface: "main", View: "compose-containers", Mode: "normal"}

	cases := []struct {
		key  string
		want KeyAction
	}{
		{KeyS, ActionContainerStart},
		{KeyCtrlS, ActionContainerStop},
		{KeyCtrlR, ActionContainerRestart},
		{KeyCtrlK, ActionContainerKill},
		{KeyL, ActionContainerLogs},
		{KeyE, ActionContainerExec},
		{KeyP, ActionContainerPause},
	}
	for _, tc := range cases {
		action, ok := resolver.Resolve(tc.key, subview)
		if !ok || action != tc.want {
			t.Fatalf("Resolve(%s, compose-containers) = %q, %v; want %q", tc.key, action, ok, tc.want)
		}
	}
}

// TestComposeProjectActionsResolveInComposeView verifies R08-12 F3/F4 —
// the 22 project-level + 3 service-level + 3 group actions resolve to
// their expected default key bindings in the compose view.
func TestComposeProjectActionsResolveInComposeView(t *testing.T) {
	resolver := NewResolver(CompileBindings(config.DefaultAppConfig().Keymap))
	compose := Context{App: "app", Surface: "main", View: "compose", Mode: "normal"}

	cases := []struct {
		key  string
		want KeyAction
	}{
		{KeyS, ActionComposeProjectStart},
		{KeyCtrlS, ActionComposeProjectStop},
		{KeyCtrlR, ActionComposeProjectRestart},
		{KeyCtrlD, ActionComposeProjectDown},
		{KeyL, ActionComposeProjectLogs},
		{KeyCtrlT, ActionComposeProjectTop},
		{KeyComma, ActionComposeProjectPort},
		{KeyCtrlK, ActionComposeProjectKill},
		{KeyCtrlF3, ActionComposeProjectEvents},
		{KeyD, ActionComposeProjectDetail},
		{KeyE, ActionComposeServiceExec},
	}
	for _, tc := range cases {
		action, ok := resolver.Resolve(tc.key, compose)
		if !ok || action != tc.want {
			t.Fatalf("Resolve(%s, compose) = %q, %v; want %q", tc.key, action, ok, tc.want)
		}
	}
}

// TestComposeProjectActionsRespectKeymapOverride verifies R08-12 F6 —
// a user-supplied keymap.compose.* override beats the registry default.
func TestComposeProjectActionsRespectKeymapOverride(t *testing.T) {
	custom := config.DefaultAppConfig().Keymap
	custom.Compose.Start = config.KeyBinding{Primary: "z"}
	custom.Compose.Down = config.KeyBinding{Primary: "shift+ctrl+x"}

	resolver := NewResolver(CompileBindings(custom))
	compose := Context{App: "app", Surface: "main", View: "compose", Mode: "normal"}

	if action, ok := resolver.Resolve("z", compose); !ok || action != ActionComposeProjectStart {
		t.Fatalf("override z → start = %q, %v; want %q", action, ok, ActionComposeProjectStart)
	}
	if action, ok := resolver.Resolve("shift+ctrl+x", compose); !ok || action != ActionComposeProjectDown {
		t.Fatalf("override shift+ctrl+x → down = %q, %v; want %q", action, ok, ActionComposeProjectDown)
	}
}

// TestComposeActionsIsolatedFromContainersView verifies R08-12 F4
// (context isolation) — pressing s on the containers panel still starts
// a single container, not the compose project.
func TestComposeActionsIsolatedFromContainersView(t *testing.T) {
	resolver := NewResolver(CompileBindings(config.DefaultAppConfig().Keymap))
	containers := Context{App: "app", Surface: "main", View: "containers", Mode: "normal"}

	if action, ok := resolver.Resolve(KeyS, containers); !ok || action != ActionContainerStart {
		t.Fatalf("Resolve(s, containers) = %q, %v; want %q", action, ok, ActionContainerStart)
	}
}
