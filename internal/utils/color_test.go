package utils

import (
	"image/color"
	"testing"

	"github.com/elizabevil/docker-tui/internal/constants"
)

func TestParseColorCommonFormats(t *testing.T) {
	tests := map[string]color.NRGBA{
		"#FAF0E6":              {R: 0xfa, G: 0xf0, B: 0xe6, A: 0xff},
		"#abc":                 {R: 0xaa, G: 0xbb, B: 0xcc, A: 0xff},
		"#FAF0E680":            {R: 0xfa, G: 0xf0, B: 0xe6, A: 0x80},
		"rgb(250, 240, 230)":   {R: 0xfa, G: 0xf0, B: 0xe6, A: 0xff},
		"rgba(250,240,230,.5)": {R: 0xfa, G: 0xf0, B: 0xe6, A: 0x80},
		"black":                {R: 0x00, G: 0x00, B: 0x00, A: 0xff},
		"grey":                 {R: 0x80, G: 0x80, B: 0x80, A: 0xff},
		"gray":                 {R: 0x80, G: 0x80, B: 0x80, A: 0xff},
		"white":                {R: 0xff, G: 0xff, B: 0xff, A: 0xff},
		"red":                  {R: 0xff, G: 0x00, B: 0x00, A: 0xff},
		"blue":                 {R: 0x00, G: 0x00, B: 0xff, A: 0xff},
	}
	for input, want := range tests {
		parsed, ok := ParseColor(input)
		if !ok {
			t.Errorf("ParseColor(%q) rejected a supported format", input)
			continue
		}
		got := color.NRGBAModel.Convert(parsed).(color.NRGBA)
		if got != want {
			t.Errorf("ParseColor(%q) = %#v, want %#v", input, got, want)
		}
	}
}

func TestParseColorTransparent(t *testing.T) {
	// ParseColor must return NRGBA{A:0} so consumers can read .A == 0
	// and skip the lipgloss call — lipgloss v2 ignores alpha and would
	// otherwise render A:0 as opaque black. See lipgloss-transparent-probe.
	sentinel := color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00}
	for _, input := range []string{"transparent", "TRANSPARENT", "Transparent", "  transparent  ", "\tTRANSPARENT\n"} {
		parsed, ok := ParseColor(input)
		if !ok {
			t.Errorf("ParseColor(%q) returned ok=false; want true", input)
			continue
		}
		got, isNRGBA := color.NRGBAModel.Convert(parsed).(color.NRGBA)
		if !isNRGBA {
			t.Errorf("ParseColor(%q) returned non-NRGBA color: %T", input, parsed)
			continue
		}
		if got != sentinel {
			t.Errorf("ParseColor(%q) = %#v, want %#v", input, got, sentinel)
		}
		if got.A != 0 {
			t.Errorf("ParseColor(%q).A = %d, want 0 (consumer-side discriminator)", input, got.A)
		}
	}
}

func TestParseColorPaletteAndANSI(t *testing.T) {
	// Palette names only resolve once registered (the style package registers
	// them at startup; tests register the names they exercise here).
	SetPaletteColor(constants.ColorForeground, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
	SetPaletteColor(constants.ColorForegroundMuted, color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff})
	for _, input := range []string{"foreground", " FOREGROUND ", "foregroundMuted", "63", "255"} {
		if parsed, ok := ParseColor(input); !ok || parsed == nil {
			t.Errorf("ParseColor(%q) failed", input)
		}
	}
	for _, input := range []string{"", "#12", "#GGGGGG", "rgb(256,0,0)", "rgba(0,0,0,2)", "256", "unknown"} {
		if _, ok := ParseColor(input); ok {
			t.Errorf("ParseColor(%q) accepted an invalid color", input)
		}
	}
}
