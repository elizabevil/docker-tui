package component

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

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
