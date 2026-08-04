package dialog

import (
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

// NotificationDialog renders a simple informational dialog with a dismiss hint.
func NotificationDialog(title, body string, termW, termH int, overlayColor string, cfg DialogConfig) string {
	if overlayColor == "" {
		overlayColor = OverlayColor(nil)
	}
	dialogW := dialogWidth(termW, cfg)
	dialogH := dialogHeight(termH, cfg)

	var parts []string
	parts = append(parts, component.GetStyle("panelTitle").Render(title))
	parts = append(parts, "")
	parts = append(parts, body)
	parts = append(parts, "")
	parts = append(parts, component.GetStyle("dim").Render(i18n.T("hint.esc_cancel")))

	return DialogBox(DialogStyle{Width: dialogW, Height: dialogH, TitleColor: component.GetStyle("panelTitle").GetForeground(), OverlayColor: overlayColor}, parts...)
}
