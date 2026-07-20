package component

import (
	_ "embed"

	"charm.land/lipgloss/v2"
)

//go:embed dialog.jsonc
var dialogDefaultData []byte

// dialogConfig maps the JSONC structure for dialog default values.
type dialogConfig struct {
	OverlayColor string `json:"overlayColor"`
}

func (c *dialogConfig) normalize() {
	if c.OverlayColor == "" {
		c.OverlayColor = "#0d1117cc"
	}
}

// DefaultDialogConfig returns default values parsed from the embedded dialog.jsonc.
// If parsing fails or overlayColor is empty, a sensible fallback is returned.
func DefaultDialogConfig() dialogConfig {
	loader := ConfigLoader[dialogConfig]{
		RawData:  dialogDefaultData,
		Fallback: dialogConfig{OverlayColor: "#0d1117cc"},
		Normalize: func(c *dialogConfig) {
			c.normalize()
		},
	}
	return loader.Load()
}

// RenderConfirmMsg returns a centered dialog string for confirm/cancel prompts.
//
//	message: main prompt text (shown with warning icon)
//	target:  secondary info line (can be "")
//	width, height: terminal dimensions for centering
//	overlayColor: hex color for the dimmed overlay background;
//	  if empty, the default from embedded dialog.jsonc is used.
//
// The dialog has a dimmed overlay background, Cancel on left (default),
// Confirm on right.
func RenderConfirmMsg(msg, target string, termW, termH int, overlayColor string) string {
	if overlayColor == "" {
		overlayColor = DefaultDialogConfig().OverlayColor
	}
	dialogW := termW * 25 / 100
	if dialogW < 40 {
		dialogW = 40
	}
	if dialogW > 60 {
		dialogW = 60
	}
	inner := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Foreground(lipgloss.Color("yellow")).Render("\u26a0 "+msg),
		"",
		lipgloss.NewStyle().Faint(true).Render("  Target: "+target),
		"",
		lipgloss.NewStyle().Faint(true).Render("  [n] Cancel  [y] Confirm"),
	)
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Foreground(lipgloss.Color("yellow")).
		Background(lipgloss.Color(overlayColor)).
		Padding(1, 2).
		Width(dialogW).
		Align(lipgloss.Center).
		Render(inner)

	return PlaceOverlay(termW, termH, dialog, overlayColor)
}

// RenderShellDialog shows the exec shell prompt as an overlay dialog.
// If overlayColor is empty, the default from embedded dialog.jsonc is used.
func RenderShellDialog(shell string, termW, termH int, overlayColor string) string {
	if overlayColor == "" {
		overlayColor = DefaultDialogConfig().OverlayColor
	}
	dialogW := termW * 25 / 100
	if dialogW < 40 {
		dialogW = 40
	}
	if dialogW > 60 {
		dialogW = 60
	}
	if shell == "" {
		shell = "/bin/sh"
	}
	inner := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Foreground(lipgloss.Color("yellow")).Render("Enter container shell"),
		"",
		lipgloss.NewStyle().Faint(true).Render("  Shell: "+shell+"\u2588"),
		"",
		lipgloss.NewStyle().Faint(true).Render("  [Enter] Confirm  [Esc] Cancel"),
	)
	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Foreground(lipgloss.Color("yellow")).
		Background(lipgloss.Color(overlayColor)).
		Padding(1, 2).
		Width(dialogW).
		Align(lipgloss.Center).
		Render(inner)

	return PlaceOverlay(termW, termH, dialog, overlayColor)
}

// PlaceOverlay centers a dialog over the terminal area WITHOUT a full-screen
// backdrop. The dialog's own background provides contrast; surrounding content
// remains visible through the terminal background.
func PlaceOverlay(termW, termH int, dialog string, overlayColor string) string {
	return lipgloss.Place(termW, termH,
		lipgloss.Center, lipgloss.Center, dialog,
	)
}

// ── Unified Dialog API ──────────────────────────────────────────

// DialogType enumerates the supported dialog modes.
type DialogType string

const (
	DialogConfirm   DialogType = "confirm"
	DialogShell     DialogType = "shell"
	DialogSelection DialogType = "selection"
	DialogExec      DialogType = "exec"
	DialogExport    DialogType = "export"
	DialogDebug     DialogType = "debug"
)

// DialogConfig holds all parameters for rendering any dialog type.
type DialogConfig struct {
	Type         DialogType
	Title        string
	Body         string
	Preview      string
	Focus        int
	Options      []string
	ConfirmKey   string
	CancelKey    string
	ConfirmLabel string
	CancelLabel  string
	TermW        int
	TermH        int
}

// RenderDialog dispatches to the correct dialog renderer based on Type.
// Returns the rendered dialog string ready for overlay placement.
func RenderDialog(cfg DialogConfig) string {
	switch cfg.Type {
	case DialogShell:
		return RenderShellDialog(cfg.Body, cfg.TermW, cfg.TermH, "")
	default:
		return RenderConfirmMsg(cfg.Title, cfg.Body, cfg.TermW, cfg.TermH, "")
	}
}
