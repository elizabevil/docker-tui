package dialog

import (
	"fmt"
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/ui/component"

	"charm.land/lipgloss/v2"
)

// FormListPane renders a read-only list of items inside a Form (e.g. the
// targets of a batch operation). It supports optional pagination and a
// header line. The component is purely presentational: paging state and
// navigation are owned by the caller.
type FormListPane struct {
	Title    string   // optional header line above the list
	Items    []string // pre-formatted item rows
	Page     int      // zero-based current page
	PageSize int      // rows per page; <=0 means show all
}

// NewFormListPane creates a list pane with the given items.
func NewFormListPane(items []string) FormListPane {
	return FormListPane{Items: items, PageSize: 0}
}

// TotalRows returns the number of rows on the current page.
func (p FormListPane) TotalRows() int {
	if p.PageSize <= 0 || len(p.Items) <= p.PageSize {
		return len(p.Items)
	}
	start := p.Page * p.PageSize
	if start >= len(p.Items) {
		return 0
	}
	remaining := len(p.Items) - start
	if remaining < p.PageSize {
		return remaining
	}
	return p.PageSize
}

// PageCount returns the total number of pages (>= 1).
func (p FormListPane) PageCount() int {
	if p.PageSize <= 0 || len(p.Items) == 0 {
		return 1
	}
	pages := (len(p.Items) + p.PageSize - 1) / p.PageSize
	if pages < 1 {
		return 1
	}
	return pages
}

// Render draws the pane: optional title, page slice of items, and a page
// indicator when paginated.
func (p FormListPane) Render(width int) string {
	if width <= 0 {
		width = 40
	}
	var parts []string
	if p.Title != "" {
		parts = append(parts, component.GetStyle(component.StyleDetailSection).Render(p.Title))
	}
	items := p.Items
	if p.PageSize > 0 {
		start := p.Page * p.PageSize
		if start >= len(p.Items) {
			items = nil
		} else {
			end := min(start+p.PageSize, len(p.Items))
			items = p.Items[start:end]
		}
	}
	for _, item := range items {
		parts = append(parts, component.GetStyle(component.StyleDim).
			Render(lipgloss.NewStyle().Width(width).Render("  "+item)))
	}
	if p.PageSize > 0 && p.PageCount() > 1 {
		pageInfo := fmt.Sprintf("%d/%d", p.Page+1, p.PageCount())
		parts = append(parts, component.GetStyle(component.StyleDim).
			Render(lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Render(pageInfo)))
	}
	return strings.Join(parts, "\n")
}
