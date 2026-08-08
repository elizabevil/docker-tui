package box

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

// Badge renders a single keystroke badge like "[R] Restart" with its own
// background. Used by the header keystroke column to show recent input.
type Badge struct {
	Key         string
	Description string
	StyleName   StyleName // typically StyleHeaderKey or StyleHeaderLast
	Background  string    // hex color; "" → terminal default
	Bold        bool
}

// Render produces the formatted badge "[Key] Description" with the configured
// background applied to the entire string.
func (b *Badge) Render() string {
	style := styleWithBackground(component.GetStyle(b.StyleName), b.Background)
	if b.Bold {
		style = style.Bold(true)
	}
	text := "[" + b.Key + "] " + b.Description
	return style.Render(text)
}

// BadgeRow joins multiple badges with the configured separator.
// The separator inherits the background so the row reads as a single box.
type BadgeRow struct {
	Badges     []Badge
	Separator  string // typically "  " or " │ "
	Background string
}

// Render produces the joined badge row.
func (br *BadgeRow) Render() string {
	if len(br.Badges) == 0 {
		return ""
	}
	parts := make([]string, len(br.Badges))
	for i, b := range br.Badges {
		parts[i] = b.Render()
	}
	sep := br.Separator
	if sep == "" {
		sep = "  "
	}
	// Color the separator with the same background so the row reads as one box.
	sepStyle := styleWithBackground(component.GetStyle(StyleDim), br.Background)
	return parts[0] + sepStyle.Render(sep) + joinWithSep(parts[1:], sepStyle.Render(sep))
}

func joinWithSep(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString(parts[0])
	for _, p := range parts[1:] {
		out.WriteString(sep + p)
	}
	return out.String()
}
