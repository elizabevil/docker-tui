package config

import (
	"embed"
	"io/fs"
)

//go:embed themes/*.jsonc
var embeddedThemes embed.FS

//go:embed default.jsonc
var defaultConfigJSON string

//go:embed styles/*.jsonc
var embeddedStyles embed.FS

// DefaultConfigJSON returns the embedded default configuration as a string.
// Used by DefaultConfig() to populate struct defaults.
func DefaultConfigJSON() string { return defaultConfigJSON }

// StylesFS returns the embedded split-style config files (BR-043 §3.4). Each
// file contributes one top-level section to the merged Config (e.g.
// styles/general.jsonc holds {"general": {...}}).
func StylesFS() fs.FS {
	sub, _ := fs.Sub(embeddedStyles, "styles") //nolint:errcheck // embedded FS path is always valid.
	return sub
}

func ThemeFS() fs.FS {
	sub, _ := fs.Sub(embeddedThemes, "themes") //nolint:errcheck // embedded FS path is always valid.
	return sub
}
