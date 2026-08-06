package config

import (
	"embed"
	"io/fs"
)

//go:embed themes/*.jsonc
var embeddedThemes embed.FS

//go:embed defaults/*.jsonc defaults/operations/scopes/*.jsonc
var embeddedDefaults embed.FS

func DefaultsFS() fs.FS {
	sub, _ := fs.Sub(embeddedDefaults, defaultsFSRoot) //nolint:errcheck // embedded FS path is always valid.
	return sub
}

func ThemeFS() fs.FS {
	sub, _ := fs.Sub(embeddedThemes, themesFSRoot) //nolint:errcheck // embedded FS path is always valid.
	return sub
}
