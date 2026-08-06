package dialog

import (
	"fmt"
	"image/color"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

type DialogConfig = config.DialogLayoutConfig

func LoadDialogConfig() DialogConfig {
	return config.DefaultAppConfig().UI.Dialog
}

// dialogWidth computes dialog width from terminal width using config percentages.
func dialogWidth(termW int, cfg DialogConfig) int {
	pct := cfg.Width.Percent
	if pct <= 0 {
		pct = 25
	}
	w := termW * pct / 100
	if w < cfg.Width.Min {
		w = cfg.Width.Min
	}
	if w > cfg.Width.Max {
		w = cfg.Width.Max
	}
	return w
}

// dialogHeight computes dialog height from terminal height using config percentages.
func dialogHeight(termH int, cfg DialogConfig) int {
	pct := cfg.Height.Percent
	if pct <= 0 {
		pct = 30
	}
	h := termH * pct / 100
	if h < cfg.Height.Min {
		h = cfg.Height.Min
	}
	if h > cfg.Height.Max {
		h = cfg.Height.Max
	}
	return h
}

// panelDialogWidth computes the dialog width from a panel body using
// cfg.PanelSize.WidthPercent. A zero WidthPercent falls back to
// cfg.Width.Percent so the legacy percentage still applies. The result
// is upper-bounded by cfg.PanelSize.MaxWidth, or cfg.Width.Max when
// MaxWidth is unset.
func panelDialogWidth(bodyW int, cfg DialogConfig) int {
	pct := cfg.PanelSize.WidthPercent
	if pct <= 0 {
		pct = cfg.Width.Percent
	}
	w := bodyW * pct / 100
	if cfg.PanelSize.MaxWidth > 0 && w > cfg.PanelSize.MaxWidth {
		w = cfg.PanelSize.MaxWidth
	} else if cfg.Width.Max > 0 && w > cfg.Width.Max {
		w = cfg.Width.Max
	}
	return w
}

// panelDialogHeight computes the dialog height from a panel body using
// cfg.PanelSize.HeightPercent. A zero HeightPercent falls back to
// cfg.Height.Percent. The result is upper-bounded by
// cfg.PanelSize.MaxHeight, or cfg.Height.Max when MaxHeight is unset.
func panelDialogHeight(bodyH int, cfg DialogConfig) int {
	pct := cfg.PanelSize.HeightPercent
	if pct <= 0 {
		pct = cfg.Height.Percent
	}
	h := bodyH * pct / 100
	if cfg.PanelSize.MaxHeight > 0 && h > cfg.PanelSize.MaxHeight {
		h = cfg.PanelSize.MaxHeight
	} else if cfg.Height.Max > 0 && h > cfg.Height.Max {
		h = cfg.Height.Max
	}
	return h
}

// dialogPosition computes the (x, y) top-left position for a dialog of size
// (dlgW, dlgH) within a terminal of size (termW, termH) using config percentages.
func dialogPosition(termW, termH, dlgW, dlgH int, cfg DialogConfig) (x, y int) {
	switch cfg.Position.Horizontal {
	case "left":
		x = 0
	case "right":
		x = termW - dlgW
	default:
		x = (termW - dlgW) / 2
	}
	switch cfg.Position.Vertical {
	case "top":
		y = 0
	case "bottom":
		y = termH - dlgH
	default:
		y = (termH - dlgH) / 2
	}
	x += cfg.Position.OffsetX
	y += cfg.Position.OffsetY
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return x, y
}

func resolveOverlay(theme *config.Theme) string {
	if theme == nil {
		theme = config.DefaultTheme()
	}
	base := theme.ResolveColor(theme.Dialog.Overlay)
	parsed, ok := utils.ParseColor(base)
	if !ok {
		fallback := config.DefaultTheme()
		base = fallback.ResolveColor(fallback.Dialog.Overlay)
		parsed, _ = utils.ParseColor(base)
	}
	rgba := color.NRGBAModel.Convert(parsed).(color.NRGBA) //nolint:errcheck // NRGBAModel.Convert always yields NRGBA.
	alpha := uint8(int(theme.Dialog.OverlayOpacity) * 255 / 100)
	return fmt.Sprintf("#%02x%02x%02x%02x", rgba.R, rgba.G, rgba.B, alpha)
}

func OverlayColor(m *state.AppModel) string {
	if m == nil {
		return resolveOverlay(nil)
	}
	return resolveOverlay(m.Dependencies.Theme)
}

func titleColorForKind(kind state.DialogKind) color.Color {
	switch kind {
	case state.DialogImageDebug:
		return component.GetStyle(component.StyleDialogWarning).GetForeground()
	case state.DialogExec:
		return component.GetStyle(component.StyleDialogConfirm).GetForeground()
	default:
		return component.GetStyle(component.StylePanelTitle).GetForeground()
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
	dlgCfg := m.Dependencies.Config.UI.Dialog
	oc := resolveOverlay(m.Dependencies.Theme)

	dialogBox := SelectionDialog(
		m.Dialog.Title, m.Dialog.Body, m.Dialog.Preview,
		actionLabelForKind(m.Dialog.Kind), m.Dialog.Focus, m.Viewport.Width, m.Viewport.Height, 0, 0, tc, oc, dlgCfg,
	)
	return lipgloss.Place(m.Viewport.Width, m.Viewport.Height,
		lipgloss.Center, lipgloss.Center, dialogBox,
	)
}

// RenderOverlay renders a scrim (full-screen dim) with a selection dialog
// centered on top. Background content remains visible but muted behind the dialog.
func RenderOverlay(content string, m *state.AppModel) string {
	tc := titleColorForKind(m.Dialog.Kind)
	dlgCfg := m.Dependencies.Config.UI.Dialog
	oc := resolveOverlay(m.Dependencies.Theme)

	dialogBox := SelectionDialog(
		m.Dialog.Title, m.Dialog.Body, m.Dialog.Preview,
		actionLabelForKind(m.Dialog.Kind), m.Dialog.Focus, m.Viewport.Width, m.Viewport.Height, 0, 0, tc, oc, dlgCfg,
	)

	return PlaceDialog(content, dialogBox, m.Viewport.Width, m.Viewport.Height, oc, dlgCfg)
}

// RenderChoiceOverlay renders the shared confirm/choice window without a
// full-screen scrim. The underlying page remains visible outside the box.
func RenderChoiceOverlay(content string, m *state.AppModel) string {
	cfg := m.Dependencies.Config.UI.Dialog
	overlay := resolveOverlay(m.Dependencies.Theme)
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
		component.GetStyle(component.StyleDialogWarning).GetForeground(),
		overlay,
		cfg,
	)
	return PlaceDialog(content, dialogBox, m.Viewport.Width, m.Viewport.Height, overlay, cfg)
}

// RenderExecOverlay renders a full-screen scrim with the exec shell dialog
// (3 shell options + custom input + confirm/cancel) centered on top.
func RenderExecOverlay(content string, m *state.AppModel) string {
	dlgCfg := m.Dependencies.Config.UI.Dialog
	oc := resolveOverlay(m.Dependencies.Theme)

	dialogBox := ExecDialog(m, oc, dlgCfg, 0, 0)
	return PlaceDialog(content, dialogBox, m.Viewport.Width, m.Viewport.Height, oc, dlgCfg)
}

// RenderChoiceOverlayInPanel splices the confirm/choice dialog box into
// content, centering it within body (per BR-040). Sizing uses body
// dimensions so the box fits the panel.
func RenderChoiceOverlayInPanel(content string, m *state.AppModel, body PanelBody) string {
	cfg := m.Dependencies.Config.UI.Dialog
	overlay := resolveOverlay(m.Dependencies.Theme)
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
		component.GetStyle(component.StyleDialogWarning).GetForeground(),
		overlay,
		cfg,
	)
	return PlaceDialogInPanel(content, dialogBox, body, cfg)
}

// RenderOverlayInPanel splices the selection dialog (used for image
// export / image debug kinds) into content, centering within body.
func RenderOverlayInPanel(content string, m *state.AppModel, body PanelBody) string {
	tc := titleColorForKind(m.Dialog.Kind)
	dlgCfg := m.Dependencies.Config.UI.Dialog
	oc := resolveOverlay(m.Dependencies.Theme)

	dialogBox := SelectionDialog(
		m.Dialog.Title, m.Dialog.Body, m.Dialog.Preview,
		actionLabelForKind(m.Dialog.Kind), m.Dialog.Focus, body.Width, body.Rows, body.Width, body.Rows, tc, oc, dlgCfg,
	)
	return PlaceDialogInPanel(content, dialogBox, body, dlgCfg)
}

// RenderExecOverlayInPanel splices the exec shell dialog into content,
// centering within body.
func RenderExecOverlayInPanel(content string, m *state.AppModel, body PanelBody) string {
	dlgCfg := m.Dependencies.Config.UI.Dialog
	oc := resolveOverlay(m.Dependencies.Theme)

	dialogBox := ExecDialog(m, oc, dlgCfg, body.Width, body.Rows)
	return PlaceDialogInPanel(content, dialogBox, body, dlgCfg)
}

// RenderContainerFormOverlayInPanel splices the container-action form dialog
// (Copy / Update / Export / Commit) into content, centering within body.
func RenderContainerFormOverlayInPanel(content string, m *state.AppModel, body PanelBody) string {
	dlgCfg := m.Dependencies.Config.UI.Dialog
	oc := resolveOverlay(m.Dependencies.Theme)

	dialogBox := FormDialog(m, oc, dlgCfg, body.Width, body.Rows)
	return PlaceDialogInPanel(content, dialogBox, body, dlgCfg)
}
