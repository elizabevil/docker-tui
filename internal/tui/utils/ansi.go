// Package utils provides shared utility functions for the TUI.
package utils

import (
	"fmt"
	"strings"
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

// VisibleLen returns the number of visible runes (excluding ANSI codes).
func VisibleLen(s string) int {
	return len([]rune(StripANSI(s)))
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
	runes := []rune(clean)
	if len(runes) <= maxVisible {
		return s
	}
	if maxVisible <= 3 {
		return string(runes[:maxVisible])
	}
	prefix := ansiPrefix(s)
	truncated := string(runes[:maxVisible-3]) + "..."
	if prefix != "" {
		return prefix + truncated + "\033[0m" // 关闭 ANSI，防止颜色泄漏
	}
	return truncated
}

// PadVisible pads a string to exactly width visible characters by appending spaces.
func PadVisible(s string, width int) string {
	vis := VisibleLen(s)
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
