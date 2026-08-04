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
	theme.Header.Label = config.TokenRef(config.ColorTokenPurple)
	theme.Dialog.Title = config.TokenRef(config.ColorTokenOrange)
	theme.Toast.Success = config.TokenRef(config.ColorTokenCyan)
	theme.Main.RowSelected = config.TokenRef(config.ColorTokenSurface)
	theme.SafeFallback.Normal = config.TokenRef(config.ColorTokenWhite)
	theme.SafeFallback.Accent = config.TokenRef(config.ColorTokenCyan)
	theme.SafeFallback.Error = config.TokenRef(config.ColorTokenRed)
	theme.Table.MarkedBackground = config.ValueRef(config.FallbackColorTableMarkedBackground)
	theme.Table.NameForeground = config.ValueRef(config.FallbackColorTableNameForeground)
	theme.Table.ColumnForeground = config.ValueRef(config.FallbackColorTableColumnForeground)

	ApplyThemeStyles(theme)

	if rawStyles.HeaderLabel.Color != string(config.FallbackColorPurple) {
		t.Fatalf("header label = %q", rawStyles.HeaderLabel.Color)
	}
	if rawStyles.DialogTitle.Color != string(config.FallbackColorOrange) {
		t.Fatalf("dialog title = %q", rawStyles.DialogTitle.Color)
	}
	if rawStyles.ToastSuccess.Color != string(config.FallbackColorCyan) {
		t.Fatalf("toast success = %q", rawStyles.ToastSuccess.Color)
	}
	if tableCfg.RowStyles.Selected.Background != string(config.FallbackColorSurface) {
		t.Fatalf("selected row background = %q", tableCfg.RowStyles.Selected.Background)
	}
	if tableCfg.RowStyles.Marked.Background != string(config.FallbackColorTableMarkedBackground) {
		t.Fatalf("marked row background = %q, want %q", tableCfg.RowStyles.Marked.Background, config.FallbackColorTableMarkedBackground)
	}
	if tableCfg.ColumnStyles.Name.Color != string(config.FallbackColorTableNameForeground) {
		t.Fatalf("name column foreground = %q, want %q", tableCfg.ColumnStyles.Name.Color, config.FallbackColorTableNameForeground)
	}
	if safeFallbackRef.Normal.GetForeground() == nil {
		t.Fatal("safe fallback normal style has no foreground")
	}
	if safeFallbackRef.Error.GetForeground() == nil {
		t.Fatal("safe fallback error style has no foreground")
	}
}
