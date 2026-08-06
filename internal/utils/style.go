package utils

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// StyleRef 定义单一样式条目的属性，对应 styles.jsonc 中的 "styles"
// 对象。Color / Background 既可以是命名调色板（见 ParseColor），也可以
// 是 hex / rgb() 等通用形式；α=0（"transparent"）在 BuildStyle 中被
// 跳过，避免 lipgloss 输出无效的 SGR。
type StyleRef struct {
	Color      string `json:"color,omitempty"`
	Background string `json:"background,omitempty"`
	Bold       bool   `json:"bold,omitempty"`
	Faint      bool   `json:"faint,omitempty"`
}

// StyleOption mutates a lipgloss.Style under construction. Options run
// before StyleRef fields apply, so ref fields overlay any seeded state
// (e.g. via FromStyle).
type StyleOption func(*lipgloss.Style)

// FromStyle seeds the new style from an existing one before ref fields
// apply. Used by component.WrapStyle to layer a StyleRef onto a
// pre-built style.
func FromStyle(existing lipgloss.Style) StyleOption {
	return func(s *lipgloss.Style) { *s = existing }
}

// WithWidth sets the Width field before ref fields apply.
func WithWidth(w int) StyleOption {
	return func(s *lipgloss.Style) { *s = s.Width(w) }
}

// BuildStyle converts a StyleRef to a lipgloss.Style, resolving color
// names through ParseColor and skipping α=0 (transparent) channels to
// avoid lipgloss emitting invalid SGR sequences. Options run before
// ref fields so seeded state (FromStyle) is overlaid by Color /
// Background / Bold / Faint.
func (r StyleRef) BuildStyle(opts ...StyleOption) lipgloss.Style {
	s := lipgloss.NewStyle()
	for _, opt := range opts {
		opt(&s)
	}
	applyChannel := func(value string, set func(c color.Color) lipgloss.Style) {
		if value == "" {
			return
		}
		parsed, ok := ParseColor(value)
		if !ok {
			return
		}
		nrgba, ok := color.NRGBAModel.Convert(parsed).(color.NRGBA)
		if !ok || nrgba.A != 0 {
			s = set(parsed)
		}
	}
	applyChannel(r.Color, s.Foreground)
	applyChannel(r.Background, s.Background)
	if r.Bold {
		s = s.Bold(true)
	}
	if r.Faint {
		s = s.Faint(true)
	}
	return s
}