package box

import (
	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// LabeledValue renders a label-value pair as a self-contained box.
//
// The label and value share one background. Because the box-level background
// is applied to both halves independently, the two halves render correctly even
// when the internal style resets (\x1b[0m) clear the parent's SGR state —
// the next half sets the background again. No reset-barrier trick is needed.
type LabeledValue struct {
	Label      string
	Value      string
	LabelStyle StyleName
	ValueStyle StyleName
	Background string // hex color; "" → terminal default (no fill)
	LabelWidth int    // 0 → no padding; >0 → pad label to this visible width
	ValueAlign lipgloss.Position
	TotalWidth int    // 0 → no width; >0 → wrap entire pair to this width
	FocusBold  bool   // if true, label is bolded (mirrors form field focus)
	FocusColor string // optional foreground override applied to the label
}

// Render produces the label + value pair with a shared background.
func (lv *LabeledValue) Render() string {
	labelStyle := styleWithBackground(component.GetStyle(lv.LabelStyle), lv.Background)
	valueStyle := styleWithBackground(component.GetStyle(lv.ValueStyle), lv.Background)

	if lv.FocusColor != "" {
		if c := backgroundColor(lv.FocusColor); c != nil {
			labelStyle = labelStyle.Foreground(c)
		}
	}
	if lv.FocusBold {
		labelStyle = labelStyle.Bold(true)
	}

	labelText := lv.Label
	if lv.LabelWidth > 0 {
		labelText = utils.PadVisible(labelText, lv.LabelWidth)
	}
	label := labelStyle.Render(labelText)
	value := valueStyle.Render(lv.Value)

	out := label + value
	if lv.TotalWidth > 0 {
		return lipgloss.NewStyle().
			Width(lv.TotalWidth).
			Align(lv.ValueAlign).
			Render(out)
	}
	return out
}
