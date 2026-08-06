package component

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

// TestLightThemeActionContainerStylesResolve confirms that the explicit
// hex overrides in light.jsonc.action.container.{window.border,
// confirm.foreground, formInput.background} survive the
// ThemePatch → ApplyThemeStyles cascade and end up in rawStyles for the
// five new StyleActionContainer* entries.
//
// This is the R06-01 light-theme acceptance check: if the patch is
// dropped or the wiring is broken, the resolved hex must NOT equal the
// default token value.
func TestLightThemeActionContainerStylesResolve(t *testing.T) {
	previousStyles := rawStyles
	previousTable := tableCfg
	previousSafe := safeFallbackRef
	defer func() {
		rawStyles = previousStyles
		tableCfg = previousTable
		safeFallbackRef = previousSafe
	}()

	resolved, err := config.LoadResolved(config.LoadOptions{ThemeName: config.ThemeName("light")})
	if err != nil {
		t.Fatalf("LoadResolved light: %v", err)
	}
	ApplyThemeStyles(resolved.Theme)

	cases := []struct {
		name     string
		got      string
		wantHex  string
	}{
		{"ActionContainerWindowBorder", rawStyles.ActionContainerWindowBorder.Color, "#3875d7"},
		{"ActionContainerConfirmForeground", rawStyles.ActionContainerConfirm.Color, "#8a3fa0"},
		{"ActionContainerFormInputBackground", rawStyles.ActionContainerFormInput.Background, "#f0f0f0"},
	}
	for _, tc := range cases {
		if tc.got != tc.wantHex {
			t.Errorf("%s = %q, want %q (light theme override)", tc.name, tc.got, tc.wantHex)
		}
	}
}

// TestLightThemeActionImageScopeDefaultsApplied confirms the image scope
// (R06-01 follow-up seat) still picks up sensible defaults even though no
// theme overrides it. Without DefaultTheme().Action.Image this test
// fails, proving the seed exists.
func TestLightThemeActionImageScopeDefaultsApplied(t *testing.T) {
	previousStyles := rawStyles
	previousTable := tableCfg
	previousSafe := safeFallbackRef
	defer func() {
		rawStyles = previousStyles
		tableCfg = previousTable
		safeFallbackRef = previousSafe
	}()

	resolved, err := config.LoadResolved(config.LoadOptions{ThemeName: config.ThemeName("light")})
	if err != nil {
		t.Fatalf("LoadResolved light: %v", err)
	}
	ApplyThemeStyles(resolved.Theme)

	if rawStyles.ActionImageWindowBorder.Color == "" || rawStyles.ActionImageWindowBorder.Color == "transparent" {
		t.Errorf("ActionImageWindowBorder.Color must be a concrete hex, got %q", rawStyles.ActionImageWindowBorder.Color)
	}
	if rawStyles.ActionImageConfirm.Color == "" || rawStyles.ActionImageConfirm.Color == "transparent" {
		t.Errorf("ActionImageConfirm.Color must be a concrete hex, got %q", rawStyles.ActionImageConfirm.Color)
	}
}