package component

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
)

func TestApplyThemeStylesProjectsFixedScopes(t *testing.T) {
	previousStyles := rawStyles
	previousTable := tableCfg
	defer func() {
		rawStyles = previousStyles
		tableCfg = previousTable
	}()

	theme := config.DefaultTheme()
	theme.Header.Label = config.TokenRef(config.ColorTokenPurple)
	theme.Dialog.Title = config.TokenRef(config.ColorTokenOrange)
	theme.Toast.Success = config.TokenRef(config.ColorTokenCyan)
	theme.Main.RowSelected = config.TokenRef(config.ColorTokenSurface)

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
}
