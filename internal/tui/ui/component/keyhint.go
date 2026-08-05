package component

import (
	"strings"
)

// KeyHint represents a single keyboard shortcut hint shown in the footer bar.
type KeyHint struct {
	Key  string // display key symbol (e.g. "j↓", "Tab", "Space")
	Desc string // description text (e.g. "Down", "Mark")
}

// RenderKeyHints renders a slice of KeyHint into a single formatted string.
// Each hint is rendered as "Key Desc" with appropriate styles, separated by "│".
func RenderKeyHints(hints []KeyHint) string {
	if len(hints) == 0 {
		return ""
	}
	var parts []string
	for _, h := range hints {
		parts = append(parts,
			GetStyle(StyleHintKey).Render(h.Key)+
				GetStyle(StyleHintDescription).Render(" "+h.Desc))
	}
	return GetStyle(StyleShortcutBar).Render(strings.Join(parts, GetStyle(StyleHintSeparator).Render(" "+BorderLineVertical+" ")))
}
