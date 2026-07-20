package keys

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

func TestCompileBindingsAppliesOverridesAndFallbacks(t *testing.T) {
	bindings := CompileBindings(config.KeymapConfig{Help: []string{"f3"}})
	if got := bindings.ByAction[ActionHelp]; len(got) != 1 || got[0] != "f3" {
		t.Fatalf("help override = %#v", got)
	}
	if got := bindings.ByAction[ActionQuit]; len(got) == 0 || got[0] != "q" {
		t.Fatalf("quit fallback = %#v", got)
	}
}

func TestResolverUsesViewPriorityForConflictingDeleteBinding(t *testing.T) {
	resolver := NewResolver(CompileBindings(config.KeymapConfig{}))
	action, ok := resolver.Resolve(KeyCtrlD, Context{App: "app", Surface: "main", View: "images", Mode: "normal"})
	if !ok || action != ActionImageRemove {
		t.Fatalf("Resolve(ctrl+d, images) = %q, %v", action, ok)
	}
}

func TestResourceDeleteOverrideDisablesLegacyGenericBinding(t *testing.T) {
	resolver := NewResolver(CompileBindings(config.KeymapConfig{ImageRemove: []string{"z"}}))
	context := Context{App: "app", Surface: "main", View: "images", Mode: "normal"}

	if action, ok := resolver.Resolve(KeyCtrlD, context); ok {
		t.Fatalf("legacy delete binding resolved after resource override: %q", action)
	}
	if action, ok := resolver.Resolve("z", context); !ok || action != ActionImageRemove {
		t.Fatalf("resource override = %q, %v", action, ok)
	}
}

func TestResolverRejectsActionOutsideContext(t *testing.T) {
	resolver := NewResolver(CompileBindings(config.KeymapConfig{ContainerStart: []string{"z"}}))
	if action, ok := resolver.Resolve("z", Context{App: "app", Surface: "main", View: "images", Mode: "normal"}); ok {
		t.Fatalf("resolved out-of-context action %q", action)
	}
}

func TestDefaultConfigMatchesRegistryDefaults(t *testing.T) {
	configured := CompileBindings(config.DefaultConfig().Keymap)
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
