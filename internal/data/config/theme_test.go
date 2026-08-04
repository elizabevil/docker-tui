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
	falseColor := TokenRef(ColorTokenPurple)
	patch := ThemePatch{Palette: &PalettePatch{Cyan: &cyan}, Dialog: &DialogStylesPatch{Border: &falseColor}}
	patch.Apply(theme)
	if theme.Palette.Cyan != cyan || theme.Palette.Green == "" {
		t.Fatalf("palette patch failed: %#v", theme.Palette)
	}
	if theme.Dialog.Border != TokenRef(ColorTokenPurple) || !isColorRef(theme.Dialog.Body) {
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
	theme.Dialog.Border = ColorRef{Token: ColorTokenRed, Value: FallbackColorRed}
	if err := ValidateTheme(theme); err == nil {
		t.Fatal("expected token plus value to fail")
	}
}
