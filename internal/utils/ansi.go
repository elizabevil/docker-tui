// Package utils provides shared utility functions for the TUI.
package utils

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
)

// StripANSI removes ANSI escape sequences from a string using a fast state machine.
// Replaces regex: `\033\[[\d;]*[a-zA-Z]` with ~10x faster byte-scanning.
func StripANSI(s string) string {
	n := len(s)
	if n == 0 {
		return ""
	}
	// Pre-allocate result buffer (most strings don't have ANSI codes)
	var buf strings.Builder
	buf.Grow(n)
	i := 0
	for i < n {
		if s[i] == '\033' && i+1 < n && s[i+1] == '[' {
			i += 2 // skip \033[
			for i < n {
				c := s[i]
				if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
					i++ // skip the terminator
					break
				}
				i++
			}
		} else {
			buf.WriteByte(s[i])
			i++
		}
	}
	return buf.String()
}

// DisplayWidth returns terminal cell width after removing ANSI sequences.
// East Asian wide characters occupy two cells; combining marks occupy zero.
func DisplayWidth(s string) int {
	return runewidth.StringWidth(StripANSI(s))
}

// VisibleLen is kept as a compatibility alias for DisplayWidth.
func VisibleLen(s string) int {
	return DisplayWidth(s)
}

// ansiPrefix extracts the leading ANSI escape sequence from s, if any.
func ansiPrefix(s string) string {
	if len(s) >= 2 && s[0] == '\033' && s[1] == '[' {
		for i := 2; i < len(s); i++ {
			c := s[i]
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
				return s[:i+1]
			}
		}
	}
	return ""
}

// TruncateVisible truncates a string to maxVisible visible runes,
// preserving any ANSI escape prefix that was in the original.
// Always closes ANSI codes after truncation to prevent color leaks.
func TruncateVisible(s string, maxVisible int) string {
	clean := StripANSI(s)
	if DisplayWidth(clean) <= maxVisible {
		return s
	}
	if maxVisible <= 3 {
		return truncateCells(clean, maxVisible)
	}
	prefix := ansiPrefix(s)
	truncated := truncateCells(clean, maxVisible-3) + "..."
	if prefix != "" {
		return prefix + truncated + "\033[0m" // 关闭 ANSI，防止颜色泄漏
	}
	return truncated
}

func truncateCells(s string, width int) string {
	if width <= 0 {
		return ""
	}
	used := 0
	var out strings.Builder
	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if used+w > width {
			break
		}
		out.WriteRune(r)
		used += w
	}
	return out.String()
}

// WrapCells splits text into lines no wider than width terminal cells.
func WrapCells(s string, width int) []string {
	if width <= 0 || DisplayWidth(s) <= width {
		return []string{s}
	}
	lines := make([]string, 0, DisplayWidth(s)/width+1)
	var line strings.Builder
	used := 0
	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if used > 0 && used+w > width {
			lines = append(lines, line.String())
			line.Reset()
			used = 0
		}
		line.WriteRune(r)
		used += w
	}
	lines = append(lines, line.String())
	return lines
}

// PadVisible pads a string to exactly width visible characters by appending spaces.
func PadVisible(s string, width int) string {
	vis := DisplayWidth(s)
	if vis >= width {
		return s
	}
	return s + strings.Repeat(" ", width-vis)
}

// HexToRGB parses a hex color string (#RRGGBB) into RGB components.
func HexToRGB(hex string) (int, int, int) {
	if len(hex) < 6 {
		return 0, 0, 0
	}
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) < 6 {
		return 0, 0, 0
	}
	r := int(hex[0]&0x0f)<<4 + int(hex[1]&0x0f)
	g := int(hex[2]&0x0f)<<4 + int(hex[3]&0x0f)
	b := int(hex[4]&0x0f)<<4 + int(hex[5]&0x0f)
	return r, g, b
}

// RGBToHex converts RGB components to a hex color string.
func RGBToHex(r, g, b int) string {
	r = clamp(r, 0, 255)
	g = clamp(g, 0, 255)
	b = clamp(b, 0, 255)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// InterpolateColor returns the hex color at position t (0.0–1.0) between start and end.
func InterpolateColor(start, end string, t float64) string {
	sr, sg, sb := HexToRGB(start)
	er, eg, eb := HexToRGB(end)
	r := clamp(int(float64(sr)+float64(er-sr)*t), 0, 255)
	g := clamp(int(float64(sg)+float64(eg-sg)*t), 0, 255)
	b := clamp(int(float64(sb)+float64(eb-sb)*t), 0, 255)
	return RGBToHex(r, g, b)
}

// BlendColors blends two hex colors: result = base + (top - base) * topPct/100.
func BlendColors(base, top string, topPct int) string {
	if topPct <= 0 {
		return base
	}
	if topPct >= 100 {
		return top
	}
	br, bg, bb := HexToRGB(base)
	tr, tg, tb := HexToRGB(top)
	t := float64(topPct) / 100
	r := clamp(int(float64(br)+float64(tr-br)*t), 0, 255)
	g := clamp(int(float64(bg)+float64(tg-bg)*t), 0, 255)
	b := clamp(int(float64(bb)+float64(tb-bb)*t), 0, 255)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
