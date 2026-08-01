// Package constants defines shared constant values used across internal packages.
package constants

// Palette color names. These are the canonical keys of the themed 12-color
// palette: UI code refers to palette colors by these names, and
// internal/utils.ParseColor resolves them at runtime through the palette
// registered by the style package.
const (
	ColorGreen   = "green"
	ColorCyan    = "cyan"
	ColorBlue    = "blue"
	ColorRed     = "red"
	ColorYellow  = "yellow"
	ColorOrange  = "orange"
	ColorPurple  = "purple"
	ColorWhite   = "white"
	ColorGray    = "gray"
	ColorGrey    = "grey"
	ColorDark    = "dark"
	ColorSurface = "surface"
	ColorBG      = "bg"
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
