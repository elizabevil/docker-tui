package config

import (
	"testing"
)

// TestLightThemeBackgroundLoaded verifies that the --theme light path
// resolves Palette.Background to the hex value declared in light.jsonc.
// If the dark theme renders correctly but --theme light does not, the
// background hex is the single most likely culprit.
func TestLightThemeBackgroundLoaded(t *testing.T) {
	resolved, err := LoadResolved(LoadOptions{ThemeName: ThemeName("light")})
	if err != nil {
		t.Fatalf("LoadResolved light: %v", err)
	}
	if got, want := string(resolved.Theme.Palette.Background), "#fefefe"; got != want {
		t.Fatalf("light theme Palette.Background = %q, want %q", got, want)
	}
}

func TestDefaultThemeBackgroundLoaded(t *testing.T) {
	resolved, err := LoadResolved(LoadOptions{ThemeName: ThemeName("default")})
	if err != nil {
		t.Fatalf("LoadResolved default: %v", err)
	}
	if got := string(resolved.Theme.Palette.Background); got == "" || got == "transparent" {
		t.Fatalf("default theme Palette.Background = %q, expected a concrete hex", got)
	}
}

func TestLightHeaderBackgroundResolvesToPalette(t *testing.T) {
	resolved, err := LoadResolved(LoadOptions{ThemeName: ThemeName("light")})
	if err != nil {
		t.Fatalf("LoadResolved light: %v", err)
	}
	got := resolved.Theme.ResolveColor(resolved.Theme.Header.Background)
	if want := "#fefefe"; got != want {
		t.Fatalf("light Header.Background ResolveColor = %q, want %q", got, want)
	}
}

func TestLightMainPanelBackgroundResolvesToPalette(t *testing.T) {
	resolved, err := LoadResolved(LoadOptions{ThemeName: ThemeName("light")})
	if err != nil {
		t.Fatalf("LoadResolved light: %v", err)
	}
	got := resolved.Theme.ResolveColor(resolved.Theme.Main.PanelBackground)
	if want := "#fefefe"; got != want {
		t.Fatalf("light Main.PanelBackground ResolveColor = %q, want %q", got, want)
	}
}

// TestLightActionContainerOverridesApply verifies the R06-01 theme contract:
// light.jsonc.action.container.{window.border, confirm.foreground,
// formInput.background} are explicit hex overrides and must survive the
// ThemePatch cascade unchanged.
func TestLightActionContainerOverridesApply(t *testing.T) {
	resolved, err := LoadResolved(LoadOptions{ThemeName: ThemeName("light")})
	if err != nil {
		t.Fatalf("LoadResolved light: %v", err)
	}
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"window.border", resolved.Theme.ResolveColor(resolved.Theme.Action.Container.Window.Border), "#3875d7"},
		{"confirm.foreground", resolved.Theme.ResolveColor(resolved.Theme.Action.Container.Confirm.Foreground), "#8a3fa0"},
		{"formInput.background", resolved.Theme.ResolveColor(resolved.Theme.Action.Container.FormInput.Background), "#f0f0f0"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("light action.container.%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

// TestLightActionImageScopeSeeded verifies the image scope carries
// defaults even though no Operation currently consumes it — this guards
// the R06-01 invariant that future image-scope Operations find a valid
// theme contract without additional wiring.
func TestLightActionImageScopeSeeded(t *testing.T) {
	resolved, err := LoadResolved(LoadOptions{ThemeName: ThemeName("light")})
	if err != nil {
		t.Fatalf("LoadResolved light: %v", err)
	}
	if got := resolved.Theme.ResolveColor(resolved.Theme.Action.Image.Window.Border); got == "" || got == "transparent" {
		t.Errorf("light action.image.window.border must resolve to a concrete colour, got %q", got)
	}
	if got := resolved.Theme.ResolveColor(resolved.Theme.Action.Image.Confirm.Foreground); got == "" || got == "transparent" {
		t.Errorf("light action.image.confirm.foreground must resolve to a concrete colour, got %q", got)
	}
}
