package component

import (
	"image/color"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

func TestLookupCoversAllRegisteredStyleNames(t *testing.T) {
	for _, name := range allStyleNames {
		if _, ok := rawStyles.lookup(name); !ok {
			t.Errorf("registered style %q has no lookup case", name)
		}
	}
}

func TestGetStyleFallsBackForUnknownName(t *testing.T) {
	got := GetStyle(StyleName("unknown")).Render("abc")
	want := safeFallbackRef.Normal.Render("abc")
	if got != want {
		t.Errorf("unknown style name did not fall back to safe fallback: got %q, want %q", got, want)
	}
}

func TestApplyThemeStylesProjectsFixedScopes(t *testing.T) {
	previousStyles := rawStyles
	previousTable := tableCfg
	previousSafe := safeFallbackRef
	defer func() {
		rawStyles = previousStyles
		tableCfg = previousTable
		safeFallbackRef = previousSafe
	}()

	theme := config.DefaultTheme()
	theme.Header.Label = config.TokenRef(config.ColorTokenAccentSecondary)
	theme.Dialog.Title = config.TokenRef(config.ColorTokenAccent)
	theme.Toast.Success = config.TokenRef(config.ColorTokenPrimary)
	theme.Main.RowSelected = config.TokenRef(config.ColorTokenBackgroundDeep)
	theme.SafeFallback.Normal = config.TokenRef(config.ColorTokenForeground)
	theme.SafeFallback.Accent = config.TokenRef(config.ColorTokenPrimary)
	theme.SafeFallback.Error = config.TokenRef(config.ColorTokenDanger)
	theme.Table.MarkedBackground = config.ValueRef(config.FallbackColorTableMarkedBackground)
	theme.Table.NameForeground = config.ValueRef(config.FallbackColorTableNameForeground)
	theme.Table.ColumnForeground = config.ValueRef(config.FallbackColorTableColumnForeground)
	theme.Footer.ShortcutBackground = config.TokenRef(config.ColorTokenBackgroundDeep)
	theme.Main.PanelBackground = config.TokenRef(config.ColorTokenDanger)
	theme.Main.ActionBarBackground = config.TokenRef(config.ColorTokenAccent)
	theme.Main.MessageRailBackground = config.TokenRef(config.ColorTokenBackgroundSubtle)
	theme.Main.QueryBarBackground = config.TokenRef(config.ColorTokenAccentSecondary)

	ApplyThemeStyles(theme)

	if rawStyles.HeaderLabel.Color != string(config.FallbackColorAccentSecondary) {
		t.Fatalf("header label = %q", rawStyles.HeaderLabel.Color)
	}
	if rawStyles.DialogTitle.Color != string(config.FallbackColorAccent) {
		t.Fatalf("dialog title = %q", rawStyles.DialogTitle.Color)
	}
	if rawStyles.ToastSuccess.Color != string(config.FallbackColorPrimary) {
		t.Fatalf("toast success = %q", rawStyles.ToastSuccess.Color)
	}
	if tableCfg.RowStyles.Selected.Background != string(config.FallbackColorBackgroundDeep) {
		t.Fatalf("selected row background = %q", tableCfg.RowStyles.Selected.Background)
	}
	if tableCfg.RowStyles.Marked.Background != string(config.FallbackColorTableMarkedBackground) {
		t.Fatalf("marked row background = %q, want %q", tableCfg.RowStyles.Marked.Background, config.FallbackColorTableMarkedBackground)
	}
	if tableCfg.ColumnStyles.Name.Color != string(config.FallbackColorTableNameForeground) {
		t.Fatalf("name column foreground = %q, want %q", tableCfg.ColumnStyles.Name.Color, config.FallbackColorTableNameForeground)
	}
	if rawStyles.ShortcutBar.Background != string(config.FallbackColorBackgroundDeep) {
		t.Fatalf("shortcut bar background = %q, want %q", rawStyles.ShortcutBar.Background, config.FallbackColorBackgroundDeep)
	}
	if rawStyles.Panel.Background != string(config.FallbackColorDanger) {
		t.Fatalf("panel background = %q", rawStyles.Panel.Background)
	}
	if rawStyles.ActionBar.Background != string(config.FallbackColorAccent) {
		t.Fatalf("action bar background = %q", rawStyles.ActionBar.Background)
	}
	if rawStyles.ActionBar.Color != string(config.FallbackColorForeground) {
		t.Fatalf("action bar color = %q", rawStyles.ActionBar.Color)
	}
	if rawStyles.FormInput.Background != string(config.FallbackColorBackground) {
		t.Fatalf("form input background = %q", rawStyles.FormInput.Background)
	}
	if rawStyles.MessageRail.Background != string(config.FallbackColorBackgroundSubtle) {
		t.Fatalf("message rail background = %q", rawStyles.MessageRail.Background)
	}
	if rawStyles.QueryBar.Background != string(config.FallbackColorAccentSecondary) {
		t.Fatalf("query bar background = %q", rawStyles.QueryBar.Background)
	}
	if rawStyles.DialogBodyBackground.Background == "" {
		t.Fatal("dialog body background is empty after ApplyThemeStyles")
	}
	if safeFallbackRef.Normal.GetForeground() == nil {
		t.Fatal("safe fallback normal style has no foreground")
	}
	if safeFallbackRef.Error.GetForeground() == nil {
		t.Fatal("safe fallback error style has no foreground")
	}
}

func TestRenderBackgroundLayerRestoresBackgroundAfterReset(t *testing.T) {
	rendered := RenderBackgroundLayer("left\033[0mright", color.NRGBA{R: 39, G: 41, B: 44, A: 255})
	if !strings.Contains(rendered, "\033[0m\033[48;2;39;41;44mright") {
		t.Fatalf("background was not restored after reset: %q", rendered)
	}
}

func TestFlattenThemeBackgroundBlendsAlphaOverAppBackground(t *testing.T) {
	theme := config.DefaultTheme()
	theme.Palette.Background = config.Color("#18191b")
	ref := config.ValueRef(config.Color("#2b2d30cc"))
	if got := flattenThemeBackground(theme, ref); got != "#27292c" {
		t.Fatalf("flattened background = %q, want %q", got, "#27292c")
	}
}

func TestHeaderLabelAndValueBackgroundIsTransparentWhenTokenIsTransparent(t *testing.T) {
	previousStyles := rawStyles
	previousSafe := safeFallbackRef
	defer func() {
		rawStyles = previousStyles
		safeFallbackRef = previousSafe
	}()

	theme := config.DefaultTheme()
	theme.Header.Label = config.TokenRef(config.ColorTokenTransparent)
	theme.Header.Value = config.TokenRef(config.ColorTokenTransparent)
	ApplyThemeStyles(theme)

	for _, ref := range []styleRef{rawStyles.HeaderLabel, rawStyles.HeaderBar} {
		if ref.Background != "" {
			t.Fatalf("background should be empty for transparent token, got %q", ref.Background)
		}
	}
}

func TestBuildStyleSkipsBackgroundWhenValueIsTransparent(t *testing.T) {
	s := buildStyle(styleRef{Color: "foregroundMuted", Background: "transparent"})
	rendered := s.Render("abc")
	if strings.Contains(rendered, "48;2") {
		t.Fatalf("transparent label/value rendered a background fill: %q", rendered)
	}
}

func TestFlattenThemeBackgroundReturnsEmptyForTransparent(t *testing.T) {
	theme := config.DefaultTheme()
	ref := config.TokenRef(config.ColorTokenTransparent)
	if got := flattenThemeBackground(theme, ref); got != "" {
		t.Fatalf("flattened background = %q, want empty", got)
	}
}
