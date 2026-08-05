// Package constants defines shared constant values used across internal packages.
package constants

// Shared Unicode display glyphs that are used by both the data and UI layers
// (kept out of internal/tui/ui/component to avoid forcing the data layer to
// depend on the UI layer).
const (
	// EmDash (—, U+2014) is used as a "no value" placeholder by both data
	// adapters (image OS/Arch summaries) and UI rendering.
	EmDash = "—"
)
