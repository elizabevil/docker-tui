package dialog

import (
	_ "embed"
	"fmt"
	"image/color"

	"github.com/elizabevil/docker-tui/internal/tui/ui/style"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

//go:embed dialog.jsonc
var dialogDefaultData []byte

// dialogConfig maps the JSONC structure for dialog appearance defaults.
type dialogConfig struct {
	WidthPercent   int    `json:"widthPercent"`
	MinWidth       int    `json:"minWidth"`
	MaxWidth       int    `json:"maxWidth"`
	HeightPercent  int    `json:"heightPercent"`
	MinHeight      int    `json:"minHeight"`
	MaxHeight      int    `json:"maxHeight"`
	XOffset        int    `json:"xOffset"`
	YOffset        int    `json:"yOffset"`
	OverlayColor   string `json:"overlayColor"`
	OverlayOpacity int    `json:"overlayOpacity"`
	ConfirmKey     string `json:"confirmKey"`
	CancelKey      string `json:"cancelKey"`
}

func defaultDialogConfig() dialogConfig {
	return dialogConfig{
		WidthPercent:   25,
		MinWidth:       40,
		MaxWidth:       80,
		HeightPercent:  30,
		MinHeight:      10,
		MaxHeight:      40,
		XOffset:        50,
		YOffset:        50,
		OverlayColor:   "#0d1117",
		OverlayOpacity: 80,
		ConfirmKey:     "Enter",
		CancelKey:      "Esc",
	}
}

func (c *dialogConfig) normalize() {
	def := defaultDialogConfig()
	if c.WidthPercent <= 0 {
		c.WidthPercent = def.WidthPercent
	}
	if c.MinWidth <= 0 {
		c.MinWidth = def.MinWidth
	}
	if c.MaxWidth <= 0 {
		c.MaxWidth = def.MaxWidth
	}
	if c.HeightPercent <= 0 {
		c.HeightPercent = def.HeightPercent
	}
	if c.MinHeight <= 0 {
		c.MinHeight = def.MinHeight
	}
	if c.MaxHeight <= 0 {
		c.MaxHeight = def.MaxHeight
	}
	if c.XOffset < 0 {
		c.XOffset = def.XOffset
	}
	if c.YOffset < 0 {
		c.YOffset = def.YOffset
	}
	if c.OverlayColor == "" {
		c.OverlayColor = def.OverlayColor
	}
	if c.OverlayOpacity < 0 {
		c.OverlayOpacity = def.OverlayOpacity
	}
	if c.ConfirmKey == "" {
		c.ConfirmKey = def.ConfirmKey
	}
	if c.CancelKey == "" {
		c.CancelKey = def.CancelKey
	}
}

// DefaultDialogViewConfig returns default values parsed from the embedded dialog.jsonc.
func DefaultDialogViewConfig() dialogConfig {
	loader := component.ConfigLoader[dialogConfig]{
		RawData:  dialogDefaultData,
		Fallback: defaultDialogConfig(),
		Normalize: func(c *dialogConfig) {
			c.normalize()
		},
	}
	return loader.Load()
}

// LoadDialogConfig loads the embedded dialog config.
func LoadDialogConfig() dialogConfig {
	return DefaultDialogViewConfig()
}

// dialogWidth computes dialog width from terminal width using config percentages.
func dialogWidth(termW int, cfg dialogConfig) int {
	pct := cfg.WidthPercent
	if pct <= 0 {
		pct = 25
	}
	w := termW * pct / 100
	if w < cfg.MinWidth {
		w = cfg.MinWidth
	}
	if w > cfg.MaxWidth {
		w = cfg.MaxWidth
	}
	return w
}

// dialogHeight computes dialog height from terminal height using config percentages.
func dialogHeight(termH int, cfg dialogConfig) int {
	pct := cfg.HeightPercent
	if pct <= 0 {
		pct = 30
	}
	h := termH * pct / 100
	if h < cfg.MinHeight {
		h = cfg.MinHeight
	}
	if h > cfg.MaxHeight {
		h = cfg.MaxHeight
	}
	return h
}

// dialogPosition computes the (x, y) top-left position for a dialog of size
// (dlgW, dlgH) within a terminal of size (termW, termH) using config percentages.
func dialogPosition(termW, termH, dlgW, dlgH int, cfg dialogConfig) (x, y int) {
	xOff := cfg.XOffset
	if xOff <= 0 {
		xOff = 50
	}
	yOff := cfg.YOffset
	if yOff <= 0 {
		yOff = 50
	}
	x = termW*xOff/100 - dlgW/2
	y = termH*yOff/100 - dlgH/2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return x, y
}

// resolveOverlay produces a hex-with-alpha color string by combining the
// app-level DialogOverlayColor (highest priority) with embed config defaults.
func resolveOverlay(appColor string, cfg dialogConfig) string {
	if appColor != "" {
		return appColor
	}
	base := cfg.OverlayColor
	if base == "" {
		base = "#0d1117"
	}
	// If the embed config already provides a full 9-char hex with alpha, use it.
	if len(base) == 9 && base[0] == '#' {
		return base
	}
	opacity := cfg.OverlayOpacity
	if opacity <= 0 {
		opacity = 80
	}
	if opacity > 100 {
		opacity = 100
	}
	alpha := fmt.Sprintf("%02x", opacity*255/100)
	return base + alpha
}

func titleColorForKind(kind state.DialogKind) color.Color {
	switch kind {
	case state.DialogImageDebug:
		return style.Colors.Orange
	case state.DialogExec:
		return style.Colors.Green
	default:
		return style.Colors.Cyan
	}
}

func actionLabelForKind(kind state.DialogKind) string {
	switch kind {
	case state.DialogImageDebug:
		return i18n.T("hint.enter_run")
	case state.DialogExec:
		return i18n.T("hint.enter_shell")
	default:
		return i18n.T("hint.enter_confirm")
	}
}

// Render builds a full-screen modal dialog for export/debug/exec modes.
func Render(m *state.AppModel) string {
	tc := titleColorForKind(m.Dialog.Kind)
	dlgCfg := LoadDialogConfig()
	oc := resolveOverlay(m.Dependencies.Config.UI.DialogOverlayColor, dlgCfg)

	dialogBox := SelectionDialog(
		m.Dialog.Title, m.Dialog.Body, m.Dialog.Preview,
		actionLabelForKind(m.Dialog.Kind), m.Dialog.Focus, m.Viewport.Width, m.Viewport.Height, tc, oc, dlgCfg,
	)
	return lipgloss.Place(m.Viewport.Width, m.Viewport.Height,
		lipgloss.Center, lipgloss.Center, dialogBox,
	)
}

// RenderOverlay renders a scrim (full-screen dim) with a selection dialog
// centered on top. Background content remains visible but muted behind the dialog.
func RenderOverlay(content string, m *state.AppModel) string {
	tc := titleColorForKind(m.Dialog.Kind)
	dlgCfg := LoadDialogConfig()
	oc := resolveOverlay(m.Dependencies.Config.UI.DialogOverlayColor, dlgCfg)

	dialogBox := SelectionDialog(
		m.Dialog.Title, m.Dialog.Body, m.Dialog.Preview,
		actionLabelForKind(m.Dialog.Kind), m.Dialog.Focus, m.Viewport.Width, m.Viewport.Height, tc, oc, dlgCfg,
	)

	return PlaceDialog(content, dialogBox, m.Viewport.Width, m.Viewport.Height, oc, dlgCfg)
}

// RenderChoiceOverlay renders the shared confirm/choice window without a
// full-screen scrim. The underlying page remains visible outside the box.
func RenderChoiceOverlay(content string, m *state.AppModel) string {
	cfg := LoadDialogConfig()
	overlay := resolveOverlay(m.Dependencies.Config.UI.DialogOverlayColor, cfg)
	options := make([]ChoiceOption, 0, len(m.Confirm.Options))
	for _, option := range m.Confirm.Options {
		options = append(options, ChoiceOption{
			Label:       option.Label,
			Description: option.Description,
			Disabled:    option.Disabled,
		})
	}
	dialogBox := ChoiceDialog(
		i18n.T("key.confirm"),
		m.Confirm.ConfirmMessage+"\nTarget: "+m.Confirm.ConfirmTarget,
		options,
		m.Confirm.Focus,
		m.Viewport.Width,
		m.Viewport.Height,
		style.Colors.Yellow,
		overlay,
		cfg,
	)
	return PlaceDialog(content, dialogBox, m.Viewport.Width, m.Viewport.Height, overlay, cfg)
}

// RenderExecOverlay renders a full-screen scrim with the exec shell dialog
// (3 shell options + custom input + confirm/cancel) centered on top.
func RenderExecOverlay(content string, m *state.AppModel) string {
	dlgCfg := LoadDialogConfig()
	oc := resolveOverlay(m.Dependencies.Config.UI.DialogOverlayColor, dlgCfg)

	dialogBox := ExecDialog(m, oc, dlgCfg)
	return PlaceDialog(content, dialogBox, m.Viewport.Width, m.Viewport.Height, oc, dlgCfg)
}

// RenderChoiceOverlayInPanel splices the confirm/choice dialog box into
// content, centering it within body (per BR-040). Sizing uses body
// dimensions so the box fits the panel.
func RenderChoiceOverlayInPanel(content string, m *state.AppModel, body PanelBody) string {
	cfg := LoadDialogConfig()
	overlay := resolveOverlay(m.Dependencies.Config.UI.DialogOverlayColor, cfg)
	options := make([]ChoiceOption, 0, len(m.Confirm.Options))
	for _, option := range m.Confirm.Options {
		options = append(options, ChoiceOption{
			Label:       option.Label,
			Description: option.Description,
			Disabled:    option.Disabled,
		})
	}
	dialogBox := ChoiceDialog(
		i18n.T("key.confirm"),
		m.Confirm.ConfirmMessage+"\nTarget: "+m.Confirm.ConfirmTarget,
		options,
		m.Confirm.Focus,
		body.Width,
		body.Rows,
		style.Colors.Yellow,
		overlay,
		cfg,
	)
	return PlaceDialogInPanel(content, dialogBox, body, cfg)
}

// RenderOverlayInPanel splices the selection dialog (used for image
// export / image debug kinds) into content, centering within body.
func RenderOverlayInPanel(content string, m *state.AppModel, body PanelBody) string {
	tc := titleColorForKind(m.Dialog.Kind)
	dlgCfg := LoadDialogConfig()
	oc := resolveOverlay(m.Dependencies.Config.UI.DialogOverlayColor, dlgCfg)

	dialogBox := SelectionDialog(
		m.Dialog.Title, m.Dialog.Body, m.Dialog.Preview,
		actionLabelForKind(m.Dialog.Kind), m.Dialog.Focus, body.Width, body.Rows, tc, oc, dlgCfg,
	)
	return PlaceDialogInPanel(content, dialogBox, body, dlgCfg)
}

// RenderExecOverlayInPanel splices the exec shell dialog into content,
// centering within body.
func RenderExecOverlayInPanel(content string, m *state.AppModel, body PanelBody) string {
	dlgCfg := LoadDialogConfig()
	oc := resolveOverlay(m.Dependencies.Config.UI.DialogOverlayColor, dlgCfg)

	dialogBox := ExecDialog(m, oc, dlgCfg)
	return PlaceDialogInPanel(content, dialogBox, body, dlgCfg)
}
