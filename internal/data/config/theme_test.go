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
			t.Errorf("validate %s: %v", name, err)
		}
	}
}

func TestThemePatchPreservesOmittedProperties(t *testing.T) {
	theme := DefaultTheme()
	cyan := Color("#00ffff")
	falseColor := TokenRef(ColorTokenAccentSecondary)
	patch := ThemePatch{Palette: &PalettePatch{Primary: &cyan}, Dialog: &DialogStylesPatch{Border: &falseColor}}
	patch.Apply(theme)
	if theme.Palette.Primary != cyan || theme.Palette.Success == "" {
		t.Fatalf("palette patch failed: %#v", theme.Palette)
	}
	if theme.Dialog.Border != TokenRef(ColorTokenAccentSecondary) || !isColorRef(theme.Dialog.Body) {
		t.Fatalf("dialog patch failed: %#v", theme.Dialog)
	}
}

func TestValidateThemeRejectsUnknownColorReference(t *testing.T) {
	theme := DefaultTheme()
	theme.Dialog.Border = ColorRef{Token: ColorToken("missing-token")}
	if err := ValidateTheme(theme); err == nil {
		t.Fatal("expected invalid color reference to fail")
	}
}

func TestValidateThemeRejectsAmbiguousColorReference(t *testing.T) {
	theme := DefaultTheme()
	theme.Dialog.Border = ColorRef{Token: ColorTokenDanger, Value: FallbackColorDanger}
	if err := ValidateTheme(theme); err == nil {
		t.Fatal("expected token plus value to fail")
	}
}

func TestResolveColorTransparentReturnsLiteralTransparent(t *testing.T) {
	theme := DefaultTheme()
	got := theme.ResolveColor(TokenRef(ColorTokenTransparent))
	if got != "transparent" {
		t.Fatalf("transparent token resolved to %q, want %q", got, "transparent")
	}
}

func TestDialogStylesHasBodyBackground(t *testing.T) {
	theme := DefaultTheme()
	if theme.Dialog.BodyBackground.Token == "" {
		t.Fatal("DefaultTheme().Dialog.BodyBackground is empty; want TokenRef(ColorTokenBackground)")
	}
	if !isColorRef(theme.Dialog.BodyBackground) {
		t.Fatalf("BodyBackground %+v failed isColorRef", theme.Dialog.BodyBackground)
	}
}

func TestIsColorRefAcceptsTransparent(t *testing.T) {
	ref := TokenRef(ColorTokenTransparent)
	if !isColorRef(ref) {
		t.Fatalf("transparent token should be accepted by isColorRef")
	}
}
