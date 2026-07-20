package config

import (
	"embed"
	"io/fs"
)

//go:embed themes/*.jsonc
var embeddedThemes embed.FS

//go:embed default.jsonc
var defaultConfigJSON string

// DefaultConfigJSON returns the embedded default configuration as a string.
// Used by DefaultConfig() to populate struct defaults.
func DefaultConfigJSON() string { return defaultConfigJSON }

func ThemeFS() fs.FS {
	sub, _ := fs.Sub(embeddedThemes, "themes")
	return sub
}
