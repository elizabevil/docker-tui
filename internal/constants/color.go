// Package constants defines shared constant values used across internal packages.
package constants

// Palette color names. These are the canonical keys of the themed 12-color
// palette: UI code refers to palette colors by these names, and
// internal/utils.ParseColor resolves them at runtime through the palette
// registered by the style package.
const (
	ColorSuccess          = "success"
	ColorPrimary          = "primary"
	ColorInfo             = "info"
	ColorDanger           = "danger"
	ColorWarning          = "warning"
	ColorAccent           = "accent"
	ColorAccentSecondary  = "accentSecondary"
	ColorForeground       = "foreground"
	ColorForegroundMuted  = "foregroundMuted"
	ColorBackground       = "background"
	ColorBackgroundSubtle = "backgroundSubtle"
	ColorBackgroundDeep   = "backgroundDeep"
	ColorBG               = "bg"
)

// Standard CSS color names supported by color parsing (internal/utils).
const (
	ColorBlack   = "black"
	ColorSilver  = "silver"
	ColorMaroon  = "maroon"
	ColorOlive   = "olive"
	ColorLime    = "lime"
	ColorAqua    = "aqua"
	ColorTeal    = "teal"
	ColorNavy    = "navy"
	ColorFuchsia = "fuchsia"
)
