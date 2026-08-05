package box

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

// Stat renders a single percentage (e.g. CPU%, memory%) with a color derived
// from the value and an optional background. The bar is intentionally not part
// of this component — a separate component can wrap Stat when a bar is needed.
type Stat struct {
	Label       string // optional leading text (e.g. "CPU ")
	LabelStyle  StyleName
	Value       float64
	Unit        string // typically "%"
	ValueWidth  int    // visible width reserved for the value+unit string
	Suffix      string // optional unit after value (e.g. "/15.6GB")
	SuffixStyle StyleName
	Color       string  // hex color for the value text; "" → no override
	WarningAt   float64 // threshold for the warning color (e.g. 50)
	DangerAt    float64 // threshold for the danger color (e.g. 80)
	Background  string  // hex color; "" → terminal default
}

// Render returns the label + coloured value (+ optional suffix) inside a
// self-contained box.
func (s *Stat) Render() string {
	out := ""
	if s.Label != "" {
		labelStyle := styleWithBackground(component.GetStyle(s.LabelStyle), s.Background)
		out += labelStyle.Render(s.Label)
	}
	valueText := formatStatValue(s.Value, s.Unit, s.ValueWidth)
	valueStyle := styleWithBackground(component.GetStyle(StyleDim), s.Background)
	if c := pickStatColor(s.Value, s.Color, s.WarningAt, s.DangerAt); c != nil {
		valueStyle = valueStyle.Foreground(c)
	}
	out += valueStyle.Render(valueText)
	if s.Suffix != "" {
		suffixStyle := styleWithBackground(component.GetStyle(s.SuffixStyle), s.Background)
		out += suffixStyle.Render(s.Suffix)
	}
	return out
}

func formatStatValue(v float64, unit string, width int) string {
	// Format like " 6%" or "73%" — pad-left to width when set.
	text := formatFloat(v) + unit
	if width > 0 {
		return padLeftVisible(text, width)
	}
	return text
}

func formatFloat(v float64) string {
	if v == float64(int(v)) {
		// Render integers without decimals to keep the column compact.
		return intStr(int(v))
	}
	return floatStr(v)
}

func intStr(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	digits := []byte{}
	for v > 0 {
		digits = append([]byte{byte('0' + v%10)}, digits...)
		v /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}

func floatStr(v float64) string {
	// One decimal place is enough for the header stat use-case.
	intPart := int(v)
	frac := int((v - float64(intPart)) * 10)
	if frac < 0 {
		frac = -frac
	}
	return intStr(intPart) + "." + string([]byte{byte('0' + frac)})
}

func padLeftVisible(text string, width int) string {
	if width <= 0 {
		return text
	}
	vis := lipgloss.Width(text)
	if vis >= width {
		return text
	}
	return repeatSpace(width-vis) + text
}

func repeatSpace(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

// pickStatColor selects a foreground color based on the value thresholds.
// If explicitColor is non-empty, it wins.
func pickStatColor(v float64, explicitColor string, warningAt, dangerAt float64) color.Color {
	if explicitColor != "" {
		return backgroundColor(explicitColor)
	}
	if dangerAt > 0 && v >= dangerAt {
		return backgroundColor("danger")
	}
	if warningAt > 0 && v >= warningAt {
		return backgroundColor("warning")
	}
	return backgroundColor("success")
}
