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
// The 16 basic HTML color names plus the "grey" spelling alias.
const (
	ColorBlack   = "black"
	ColorSilver  = "silver"
	ColorGray    = "gray"
	ColorGrey    = "grey"
	ColorWhite   = "white"
	ColorMaroon  = "maroon"
	ColorRed     = "red"
	ColorPurple  = "purple"
	ColorFuchsia = "fuchsia"
	ColorGreen   = "green"
	ColorLime    = "lime"
	ColorOlive   = "olive"
	ColorYellow  = "yellow"
	ColorNavy    = "navy"
	ColorBlue    = "blue"
	ColorTeal    = "teal"
	ColorAqua    = "aqua"
)
