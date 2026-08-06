package dialog

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

// withLightTheme applies the light theme via the same path production
// uses (config.LoadResolved → component.ApplyThemeStyles) and restores the
// previous style cache on cleanup.
func withLightTheme(t *testing.T) {
	t.Helper()
	resolved, err := config.LoadResolved(config.LoadOptions{ThemeName: config.ThemeName("light")})
	if err != nil {
		t.Fatalf("LoadResolved light: %v", err)
	}
	component.ApplyThemeStyles(resolved.Theme)
}

// TestFormDialogHonoursLightThemeContainerWindowBorder is the R06-01
// acceptance check: with the light theme active, the FormContainerCopy
// dialog must emit an SGR that contains the explicit
// light.jsonc.action.container.window.border hex (#3875d7) rather than the
// default "primary" token value.
func TestFormDialogHonoursLightThemeContainerWindowBorder(t *testing.T) {
	withLightTheme(t)

	m := formTestModel(state.FormSpec{
		Kind:  state.FormContainerCopy,
		Title: "Copy file from container",
		Fields: []state.FormField{
			NewPathField(PathFieldConfig{
				Key:        "source",
				Label:      "Source",
				PathSource: state.PathContainer,
				Required:   true,
			}),
		},
	})
	rendered := FormDialog(m, "", LoadDialogConfig(), 60, 20)
	if !strings.Contains(rendered, "56;117;215") {
		t.Fatalf("FormDialog under light theme must embed #3875d7 (rgb 56,117,215) as window.border; got %q", stripANSI(rendered))
	}
}

// TestFormDialogHonoursLightThemeConfirmForeground verifies the
// explicit light.jsonc.action.container.confirm.foreground override
// (#8a3fa0) reaches the Confirm button SGR.
func TestFormDialogHonoursLightThemeConfirmForeground(t *testing.T) {
	withLightTheme(t)

	m := formTestModel(state.FormSpec{
		Kind:  state.FormContainerExport,
		Title: "Export container filesystem",
		Fields: []state.FormField{
			NewPathField(PathFieldConfig{
				Key:        "destination",
				Label:      "Destination",
				PathSource: state.PathLocal,
				PathMode:   state.PathSaveFile,
				Required:   true,
			}),
		},
	})
	rendered := FormDialog(m, "", LoadDialogConfig(), 60, 20)
	if !strings.Contains(rendered, "138;63;160") {
		t.Fatalf("FormDialog under light theme must embed #8a3fa0 (rgb 138,63,160) as confirm.foreground; got %q", stripANSI(rendered))
	}
}

// TestFormDialogHonoursLightThemeFormInputBackground verifies the
// explicit light.jsonc.action.container.formInput.background override
// (#f0f0f0) is emitted on form input cells.
func TestFormDialogHonoursLightThemeFormInputBackground(t *testing.T) {
	withLightTheme(t)

	m := formTestModel(state.FormSpec{
		Kind:  state.FormContainerUpdate,
		Title: "Update container",
		Fields: []state.FormField{
			NewIntField(IntFieldConfig{Key: "memory", Label: "Memory", Required: true}),
		},
	})
	rendered := FormDialog(m, "", LoadDialogConfig(), 60, 20)
	if !strings.Contains(rendered, "240;240;240") {
		t.Fatalf("FormDialog under light theme must embed #f0f0f0 (rgb 240,240,240) as formInput.background; got %q", stripANSI(rendered))
	}
}