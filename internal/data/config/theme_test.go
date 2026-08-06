package config

import "testing"

func TestEmbeddedThemesCascadeFromDefault(t *testing.T) {
	base, err := loadNamedTheme(ThemeName(themeDefaultName), emptyValue, DefaultTheme())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range ListThemes() {
		loaded, loadErr := loadNamedTheme(ThemeName(name), emptyValue, base.Theme)
		if loadErr != nil {
			t.Errorf("load %s: %v", name, loadErr)
			continue
		}
		if err := ValidateTheme(loaded.Theme); err != nil {
			t.Errorf("validate %s: %v", name, loadErr)
		}
	}
}

func TestEmbeddedDefaultThemeBackgroundExperiment(t *testing.T) {
	loaded, err := loadNamedTheme(ThemeName(themeDefaultName), emptyValue, DefaultTheme())
	if err != nil {
		t.Fatal(err)
	}
	theme := loaded.Theme
	transparent := TokenRef(ColorTokenTransparent)
	if theme.Palette.Background != Color("#18191b") {
		t.Fatalf("app background = %q", theme.Palette.Background)
	}
	if theme.Surfaces.Header != transparent ||
		theme.Surfaces.Panel != transparent ||
		theme.Surfaces.ActionBar != transparent ||
		theme.Surfaces.MessageRail != transparent ||
		theme.Surfaces.QueryBar != transparent ||
		theme.Surfaces.Footer != transparent ||
		theme.Surfaces.Toast != transparent {
		t.Fatalf("surfaces should default to transparent: %#v", theme.Surfaces)
	}
	if theme.Surfaces.RowSelected != "info" {
		t.Fatalf("rowSelected = %q, want info token", theme.Surfaces.RowSelected)
	}
}

func TestThemePatchPreservesOmittedProperties(t *testing.T) {
	theme := DefaultTheme()
	cyan := Color("#00ffff")
	falseColor := TokenRef(ColorTokenAccentSecondary)
	background := TokenRef(ColorTokenBackgroundSubtle)
	patch := ThemePatch{
		Palette: &PalettePatch{Primary: &cyan},
		Surfaces: &SurfacesStylesPatch{
			Panel:     &background,
			ActionBar: &background,
		},
		Chrome: &ChromeStylesPatch{DialogBorder: &falseColor},
	}
	patch.Apply(theme)
	if theme.Palette.Primary != cyan || theme.Palette.Success == "" {
		t.Fatalf("palette patch failed: %#v", theme.Palette)
	}
	if theme.Chrome.DialogBorder != TokenRef(ColorTokenAccentSecondary) || !isColorRef(theme.Text.DialogBody) {
		t.Fatalf("dialog patch failed: %#v", theme.Chrome)
	}
	if theme.Surfaces.Panel != background || theme.Surfaces.ActionBar != background ||
		theme.Surfaces.QueryBar != TokenRef(ColorTokenTransparent) {
		t.Fatalf("surfaces patch failed: %#v", theme.Surfaces)
	}
}

func TestValidateThemeRejectsUnknownColorReference(t *testing.T) {
	theme := DefaultTheme()
	theme.Chrome.DialogBorder = ColorRef("missing-token")
	if err := ValidateTheme(theme); err == nil {
		t.Fatal("expected invalid color reference to fail")
	}
}

func TestValidateThemeRejectsAmbiguousColorReference(t *testing.T) {
	theme := DefaultTheme()
	// A literal colour is accepted on its own; the ambiguity only
	// arises if both Token and Value were set in the old struct form,
	// which is no longer expressible. This test now exercises the
	// "invalid token name" branch instead of the now-impossible
	// ambiguity branch.
	theme.Chrome.DialogBorder = ColorRef("#not-a-color")
	if err := ValidateTheme(theme); err == nil {
		t.Fatal("expected invalid colour value to fail")
	}
}

func TestResolveColorTransparentReturnsLiteralTransparent(t *testing.T) {
	theme := DefaultTheme()
	got := theme.ResolveColor(TokenRef(ColorTokenTransparent))
	if got != "transparent" {
		t.Fatalf("transparent token resolved to %q, want %q", got, "transparent")
	}
}

func TestSurfacesDialogBodyHasBackground(t *testing.T) {
	theme := DefaultTheme()
	if theme.Surfaces.DialogBody == "" {
		t.Fatal("DefaultTheme().Surfaces.DialogBody is empty; want TokenRef(ColorTokenBackground)")
	}
	if !isColorRef(theme.Surfaces.DialogBody) {
		t.Fatalf("DialogBody %q failed isColorRef", theme.Surfaces.DialogBody)
	}
}

func TestSurfacesBackgroundsDefaultToTransparent(t *testing.T) {
	theme := DefaultTheme()
	refs := []ColorRef{
		theme.Surfaces.ActionBar,
		theme.Surfaces.MessageRail,
		theme.Surfaces.QueryBar,
		theme.Surfaces.Footer,
		theme.Surfaces.Header,
		theme.Surfaces.Toast,
	}
	for _, ref := range refs {
		if ref != TokenRef(ColorTokenTransparent) {
			t.Fatalf("surfaces background = %+v, want transparent token", ref)
		}
		if !isColorRef(ref) {
			t.Fatalf("surfaces background %+v failed isColorRef", ref)
		}
	}
}

func TestIsColorRefAcceptsTransparent(t *testing.T) {
	ref := TokenRef(ColorTokenTransparent)
	if !isColorRef(ref) {
		t.Fatalf("transparent token should be accepted by isColorRef")
	}
}