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
