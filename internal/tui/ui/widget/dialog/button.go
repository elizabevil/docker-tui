package dialog

import (
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"

	"charm.land/lipgloss/v2"
)

// ButtonState describes the visual / interactive state of a Button.
type ButtonState int

const (
	// ButtonNormal is the default unfocused state.
	ButtonNormal ButtonState = iota
	// ButtonFocused is the state of the button currently selected in the
	// Tab loop.
	ButtonFocused
	// ButtonDanger highlights a destructive action (e.g. Force Remove) even
	// while focused.
	ButtonDanger
	// ButtonDisabled renders the button faint and non-interactive.
	ButtonDisabled
)

// Button is a business-agnostic form action button (e.g. Cancel / Confirm).
// It renders a key hint, an indicator triangle, and a label, matching the
// existing dialog language. The Button itself carries no behavior: the caller
// decides what a focus / activation means.
type Button struct {
	Key   string // key hint shown before the indicator (e.g. "esc", "enter")
	Label string // human-readable action label
	State ButtonState
}

// NewButton creates a Button with the given key hint and label.
func NewButton(key, label string) Button {
	return Button{Key: key, Label: label}
}

// Render draws the button with styling derived from its state.
func (b Button) Render() string {
	style := lipgloss.NewStyle().
		Foreground(component.GetStyle(component.StyleDim).GetForeground())
	switch b.State {
	case ButtonFocused:
		style = style.
			Foreground(component.GetStyle(component.StyleDialogConfirm).GetForeground()).
			Bold(true)
	case ButtonDanger:
		style = style.
			Foreground(component.GetStyle(component.StyleDialogError).GetForeground()).
			Bold(true)
	case ButtonDisabled:
		style = style.Faint(true)
	}
	label := b.Label
	if b.State == ButtonFocused || b.State == ButtonDanger {
		label = component.ButtonIndicator + " " + label
	}
	return style.Render(b.Key + " " + label)
}
