package utils

import (
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/constants"
)

// palette maps palette color names to their resolved colors. It is populated
// by the style package: defaults are registered at package init and the
// themed values are re-registered on every theme application.
var palette = make(map[string]color.Color)

// SetPaletteColor registers a named palette color for ParseColor resolution.
// Names are stored lowercased so lookups are case-insensitive.
func SetPaletteColor(name string, c color.Color) {
	palette[strings.ToLower(name)] = c
}

// ParseColor resolves theme palette names and common terminal/CSS color forms:
// palette names (registered via SetPaletteColor), standard CSS names, #hex
// (3/4/6/8 digits), rgb()/rgba() functions, and ANSI 0-255 color indices.
func ParseColor(value string) (color.Color, bool) {
	value = strings.TrimSpace(value)
	name := strings.ToLower(value)
	if c, ok := palette[name]; ok {
		return c, true
	}
	if named, ok := standardNamedColors[name]; ok {
		return named, true
	}
	if parsed, ok := parseHexColor(value); ok {
		return parsed, true
	}
	if parsed, ok := parseFunctionalColor(name); ok {
		return parsed, true
	}
	if index, err := strconv.Atoi(value); err == nil && index >= 0 && index <= 255 {
		return lipgloss.Color(value), true
	}
	return nil, false
}

var standardNamedColors = map[string]color.Color{
	constants.ColorBlack:   color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff},
	constants.ColorSilver:  color.NRGBA{R: 0xc0, G: 0xc0, B: 0xc0, A: 0xff},
	constants.ColorMaroon:  color.NRGBA{R: 0x80, G: 0x00, B: 0x00, A: 0xff},
	constants.ColorOlive:   color.NRGBA{R: 0x80, G: 0x80, B: 0x00, A: 0xff},
	constants.ColorLime:    color.NRGBA{R: 0x00, G: 0xff, B: 0x00, A: 0xff},
	constants.ColorAqua:    color.NRGBA{R: 0x00, G: 0xff, B: 0xff, A: 0xff},
	constants.ColorTeal:    color.NRGBA{R: 0x00, G: 0x80, B: 0x80, A: 0xff},
	constants.ColorNavy:    color.NRGBA{R: 0x00, G: 0x00, B: 0x80, A: 0xff},
	constants.ColorFuchsia: color.NRGBA{R: 0xff, G: 0x00, B: 0xff, A: 0xff},
}

func parseHexColor(value string) (color.Color, bool) {
	if !strings.HasPrefix(value, "#") {
		return nil, false
	}
	hex := value[1:]
	if len(hex) == 3 || len(hex) == 4 {
		var expanded strings.Builder
		for _, digit := range hex {
			expanded.WriteRune(digit)
			expanded.WriteRune(digit)
		}
		hex = expanded.String()
	}
	if len(hex) != 6 && len(hex) != 8 {
		return nil, false
	}
	raw, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return nil, false
	}
	if len(hex) == 6 {
		raw = raw<<8 | 0xff
	}
	return color.NRGBA{R: uint8(raw >> 24), G: uint8(raw >> 16), B: uint8(raw >> 8), A: uint8(raw)}, true
}

func parseFunctionalColor(value string) (color.Color, bool) {
	open := strings.IndexByte(value, '(')
	if open < 0 || !strings.HasSuffix(value, ")") {
		return nil, false
	}
	name := value[:open]
	parts := strings.Split(value[open+1:len(value)-1], ",")
	if (name != "rgb" || len(parts) != 3) && (name != "rgba" || len(parts) != 4) {
		return nil, false
	}
	channels := [4]uint8{0, 0, 0, 0xff}
	for i := range 3 {
		n, err := strconv.Atoi(strings.TrimSpace(parts[i]))
		if err != nil || n < 0 || n > 255 {
			return nil, false
		}
		channels[i] = uint8(n)
	}
	if name == "rgba" {
		alpha, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		if err != nil || alpha < 0 || alpha > 1 {
			return nil, false
		}
		channels[3] = uint8(alpha*255 + 0.5)
	}
	return color.NRGBA{R: channels[0], G: channels[1], B: channels[2], A: channels[3]}, true
}
