package component

import (
	"github.com/elizabevil/docker-tui/internal/tui/utils"
)

// BreadcrumbItem represents a single navigation segment in the breadcrumb path.
type BreadcrumbItem struct {
	Label string // display name (may be i18n key)
	ID    string // identifier (for keyboard navigation)
}

// RenderBreadcrumb renders a navigation breadcrumb path with truncation.
//
//	items: path segments, root first  (e.g. {"Images", "containers", "detail"})
//	separator: segment separator (e.g. " > ")
//	width: maximum available width in characters
//
// When the rendered path exceeds width, the leftmost segments are truncated:
//
//	"Images > containers > detail"  (full)
//	"Ima... > containers > detail"  (truncated)
func RenderBreadcrumb(items []BreadcrumbItem, separator string, width int) string {
	if len(items) == 0 || width <= 0 {
		return ""
	}
	if separator == "" {
		separator = " > "
	}

	// Calculate full length
	full := ""
	for i, item := range items {
		if i > 0 {
			full += separator
		}
		full += item.Label
	}

	visLen := utils.VisibleLen(full)
	if visLen <= width {
		return GetStyle("breadcrumb").Render(full)
	}

	// Need to truncate: keep rightmost segments readable
	// Start from the right and add segments until we hit the limit
	rightParts := make([]string, 0, len(items))
	rightLen := 0

	for i := len(items) - 1; i >= 0; i-- {
		part := items[i].Label
		partLen := utils.VisibleLen(part)

		if i > 0 {
			partLen += utils.VisibleLen(separator)
		}

		if rightLen > 0 {
			partLen += utils.VisibleLen(separator)
		}

		if rightLen+partLen+3 > width { // 3 = "..."
			break
		}

		if rightLen > 0 {
			rightParts = append(rightParts, part)
		} else {
			rightParts = append(rightParts, part)
		}
		rightLen += partLen
	}

	// Reverse and prepend truncation marker
	result := "..."
	for i := len(rightParts) - 1; i >= 0; i-- {
		if result != "..." {
			result += separator
		}
		result += rightParts[i]
	}

	return GetStyle("breadcrumb").Render(result)
}
