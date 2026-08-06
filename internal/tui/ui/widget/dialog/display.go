package dialog

import (
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"

	"charm.land/lipgloss/v2"
)

// ── FormInfo banner icons ───────────────────────────────────────
const (
	// IconInfo is the U+2139 INFORMATION SOURCE prefix rendered by
	// FormDisplayInfo banners.
	IconInfo = "\u2139"
	// IconWarning is the U+26A0 WARNING SIGN prefix rendered by
	// FormDisplayWarning banners.
	IconWarning = "\u26a0"
)

// FormDisplayLevel is the severity of a FormInfo banner.
type FormDisplayLevel int

const (
	// FormDisplayInfo is a neutral informational banner (ℹ).
	FormDisplayInfo FormDisplayLevel = iota
	// FormDisplayWarning is a caution banner (⚠).
	FormDisplayWarning
)

// FormDisplay renders static, non-editable information inside a Form (e.g.
// target ID, current size, status). It has no focus, no state and no
// behavior — it is a pure presentation component.
type FormDisplay struct {
	Label string // optional leading label, e.g. "Target ID"
	Value string
	Dim   bool // render with muted style
}

// NewFormDisplay creates a static read-only display row.
func NewFormDisplay(label, value string) FormDisplay {
	return FormDisplay{Label: label, Value: value}
}

// Render draws the display row: an optional dim label followed by the value.
func (d FormDisplay) Render(width int) string {
	if width <= 0 {
		width = 40
	}
	if d.Dim {
		return component.GetStyle(component.StyleDim).
			Render(lipgloss.NewStyle().Width(width).Render(d.Label + " " + d.Value))
	}
	prefix := ""
	if d.Label != "" {
		prefix = component.GetStyle(component.StyleDim).Render(d.Label) + " "
	}
	return lipgloss.NewStyle().Width(width).Render(prefix + d.Value)
}

// FormInfo renders a warning / info banner inside a Form. It is
// non-interactive and carries an icon prefix. Optionally the rendered text
// may contain a link — the caller is responsible for any click behavior.
type FormInfo struct {
	Level FormDisplayLevel
	Text  string
}

// NewFormInfo creates an informational banner (ℹ icon).
func NewFormInfo(text string) FormInfo {
	return FormInfo{Level: FormDisplayInfo, Text: text}
}

// NewFormWarning creates a warning banner (⚠ icon).
func NewFormWarning(text string) FormInfo {
	return FormInfo{Level: FormDisplayWarning, Text: text}
}

// Render draws the banner with its level icon and style.
func (i FormInfo) Render(width int) string {
	if width <= 0 {
		width = 40
	}
	icon, styleName := IconInfo, component.StyleDialogConfirm
	if i.Level == FormDisplayWarning {
		icon, styleName = IconWarning, component.StyleDialogWarning
	}
	body := lipgloss.NewStyle().Width(width - 2).Render(icon + " " + i.Text)
	return component.GetStyle(styleName).Render(body)
}
