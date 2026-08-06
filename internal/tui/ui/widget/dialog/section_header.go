package dialog

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"

	"charm.land/lipgloss/v2"
)

// FormSectionHeader renders a non-interactive visual group separator inside a
// Form, e.g. "Image Options" between field groups. It carries no state and no
// behavior; the Form owns layout placement.
type FormSectionHeader struct {
	Title string
}

// NewFormSectionHeader creates a section header with the given title.
func NewFormSectionHeader(title string) FormSectionHeader {
	return FormSectionHeader{Title: title}
}

// Render draws the title followed by a horizontal rule filling the remaining
// width, matching the detail-page section delimiter style.
func (h FormSectionHeader) Render(width int) string {
	if width <= 0 {
		width = 40
	}
	title := component.GetStyle(component.StyleDetailSection).Render(h.Title)
	if h.Title == "" {
		return component.GetStyle(component.StyleDim).
			Render(strings.Repeat(component.BoxHorizontal, width))
	}
	fill := width - lipgloss.Width(title)
	if fill < 1 {
		fill = 1
	}
	rule := component.GetStyle(component.StyleDim).
		Render(strings.Repeat(component.BoxHorizontal, fill))
	return title + rule
}
