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
	if theme.Header.Background != transparent || theme.Header.Label != transparent ||
		theme.Header.Value != transparent {
		t.Fatalf("header slots should default to transparent: %#v", theme.Header)
	}
	if theme.Main.PanelBackground != transparent ||
		theme.Footer.StatusBackground != transparent ||
		theme.Footer.ShortcutBackground != transparent {
		t.Fatalf("panel/footer backgrounds should default to transparent: %#v", theme)
	}
	if theme.Main.RowSelected != ValueRef(Color("grey")) {
		t.Fatalf("rowSelected = %+v", theme.Main.RowSelected)
	}
}

func TestThemePatchPreservesOmittedProperties(t *testing.T) {
	theme := DefaultTheme()
	cyan := Color("#00ffff")
	falseColor := TokenRef(ColorTokenAccentSecondary)
	background := TokenRef(ColorTokenBackgroundSubtle)
	patch := ThemePatch{
		Palette: &PalettePatch{Primary: &cyan},
		Main: &MainStylesPatch{
			PanelBackground:     &background,
			ActionBarBackground: &background,
		},
		Dialog: &DialogStylesPatch{Border: &falseColor},
	}
	patch.Apply(theme)
	if theme.Palette.Primary != cyan || theme.Palette.Success == "" {
		t.Fatalf("palette patch failed: %#v", theme.Palette)
	}
	if theme.Dialog.Border != TokenRef(ColorTokenAccentSecondary) || !isColorRef(theme.Dialog.Body) {
		t.Fatalf("dialog patch failed: %#v", theme.Dialog)
	}
	if theme.Main.PanelBackground != background || theme.Main.ActionBarBackground != background ||
		theme.Main.QueryBarBackground != TokenRef(ColorTokenTransparent) {
		t.Fatalf("main background patch failed: %#v", theme.Main)
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

func TestMainBackgroundsDefaultToTransparent(t *testing.T) {
	theme := DefaultTheme()
	refs := []ColorRef{
		theme.Main.ActionBarBackground,
		theme.Main.MessageRailBackground,
		theme.Main.QueryBarBackground,
	}
	for _, ref := range refs {
		if ref != TokenRef(ColorTokenTransparent) {
			t.Fatalf("main background = %+v, want transparent token", ref)
		}
		if !isColorRef(ref) {
			t.Fatalf("main background %+v failed isColorRef", ref)
		}
	}
}

func TestIsColorRefAcceptsTransparent(t *testing.T) {
	ref := TokenRef(ColorTokenTransparent)
	if !isColorRef(ref) {
		t.Fatalf("transparent token should be accepted by isColorRef")
	}
}
