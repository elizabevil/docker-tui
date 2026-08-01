package component

import (
	"embed"
	"image/color"

	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
)

//go:embed styles.jsonc
var embeddedConfig embed.FS

func init() {
	loadComponentStyles()
}

// GetColor resolves palette names and common color formats, with a white fallback.
// 简单委托给 style.Color，仅为保持 API 兼容。
func GetColor(name string) color.Color {
	return style.Color(name)
}
