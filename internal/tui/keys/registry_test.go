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
