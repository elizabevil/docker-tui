package box

import (
	"charm.land/lipgloss/v2"
)

// BorderedBox wraps content in a border with an explicit background. Unlike
// the previous ad-hoc `lipgloss.NewStyle().Border(...).Render(content)` calls
// scattered through the codebase, this component carries the background on
// the box itself so the padding characters inside the border pick up the
// intended fill instead of the terminal default.
type BorderedBox struct {
	Content     string
	BorderKind  lipgloss.Border // zero value → defaults to lipgloss.RoundedBorder()
	UseRounded  bool            // when BorderKind is the zero value, fall back to rounded
	BorderColor string          // hex color; "" → no border colour override
	Background  string          // hex color; "" → terminal default
	Padding     [2]int          // vertical, horizontal
	Width       int             // 0 → no width override
	Height      int             // 0 → no height override
}

// Render produces the bordered box with the given background.
func (bb *BorderedBox) Render() string {
	style := lipgloss.NewStyle()
	if c := backgroundColor(bb.Background); c != nil {
		style = style.Background(c)
	}
	if !bb.UseRounded && (bb.BorderKind != lipgloss.Border{}) {
		style = style.Border(bb.BorderKind)
	} else {
		style = style.Border(lipgloss.RoundedBorder())
	}
	if c := backgroundColor(bb.BorderColor); c != nil {
		style = style.BorderForeground(c)
	}
	style = style.Padding(bb.Padding[0], bb.Padding[1])
	if bb.Width > 0 {
		style = style.Width(bb.Width)
	}
	if bb.Height > 0 {
		style = style.Height(bb.Height)
	}
	return style.Render(bb.Content)
}

// GetStyle exposes the resolved style for callers that need to compose
// further (rare; mostly for tests).
func (bb *BorderedBox) GetStyle() lipgloss.Style {
	style := lipgloss.NewStyle()
	if c := backgroundColor(bb.Background); c != nil {
		style = style.Background(c)
	}
	style = style.Padding(bb.Padding[0], bb.Padding[1])
	if bb.Width > 0 {
		style = style.Width(bb.Width)
	}
	return style
}
