package dialog

import (
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"

	"charm.land/lipgloss/v2"
)

// renderFormPopup draws a single-column list popup for Select / MultiSelect /
// Path fields (BR-041 §4.2-§4.4). The popup is rendered as its own bordered
// modal box replacing the base form while selection is active. This keeps the
// popup bounded by the panel instead of growing the form vertically.
func renderFormPopup(form state.FormState, box string, dialogW, _ int) string {
	field := form.PopupField()
	if field == nil {
		return box
	}
	rows := popupRows(field)
	if len(rows) == 0 {
		message := i18n.T("form.path.no_matches")
		if field.PathLoading {
			message = i18n.T("form.path.loading")
		}
		rows = []string{message}
	}

	popupW := dialogW - 6
	if popupW < 16 {
		popupW = 16
	}
	innerW := popupW - 4
	lines := visiblePopupRows(form, field, rows, innerW, state.FormPopupVisibleRows)
	for len(lines) < state.FormPopupVisibleRows {
		lines = append(lines, component.FormRow("", 0, innerW, ""))
	}
	lines = append([]string{component.FormRow("", 0, innerW, field.Label), component.FormRow("", 0, innerW, "")}, lines...)

	return DialogBox(DialogStyle{Width: popupW, TitleColor: style.Colors.Cyan, LeftAligned: true}, lines...)
}

func visiblePopupRows(form state.FormState, field *state.FormField, rows []string, innerW, limit int) []string {
	if limit <= 0 || len(rows) <= limit {
		return shadePopup(form, field, rows, innerW, 0)
	}
	start := min(max(0, form.Popup.Cursor-limit/2), len(rows)-limit)
	return shadePopup(form, field, rows[start:start+limit], innerW, start)
}

// popupRows returns the rendered cells for the popup-owning field. Each row
// is the full prefixed string (cursor marker + checkbox/type + name) so that
// the column edges are identical across rows (BR-041 §4.2/§4.3/§4.4).
func popupRows(field *state.FormField) []string {
	switch field.Kind {
	case state.FormMultiSelect, state.FormSelect:
		// Select uses plain rows; the shadePopup adds "> "/"  ".
		return field.Options
	case state.FormPath:
		out := make([]string, 0, len(field.Suggestions))
		for i := range field.Suggestions {
			out = append(out, pathRowLabel(&field.Suggestions[i]))
		}
		return out
	default:
		return nil
	}
}

// pathRowLabel renders one path candidate as a fixed-width type column
// followed by the basename; directories get a trailing separator and a
// [DIR] type label, files get [FILE]. The type column is always 7 visible
// cells (e.g. "[DIR]  " or "[FILE] ") so basenames line up.
func pathRowLabel(e *state.PathEntry) string {
	typeCol := "[FILE] "
	if e.IsDir {
		typeCol = "[DIR]  "
		name := e.Name
		if e.IsDir && !strings.HasSuffix(name, "/") {
			name += "/"
		}
		return typeCol + name
	}
	return typeCol + e.Name
}

// entryLabel is retained for callers that need the basename only.
func entryLabel(e *state.PathEntry) string {
	if e.IsDir {
		return e.Name + "/"
	}
	return e.Name
}

// shadePopup renders each option row. The first column is fixed at 2 cells
// ("> " for the cursor row, "  " for the rest) so checkbox/type/text columns
// line up regardless of the cursor (BR-041 §4.2/§4.3/§4.4). The cursor row
// also gets a background highlight so the selection is visible without
// relying on foreground colour alone.
func shadePopup(form state.FormState, field *state.FormField, rows []string, innerW, offset int) []string {
	out := make([]string, 0, len(rows))
	cursor := form.Popup.Cursor
	highlight := lipgloss.NewStyle().Foreground(style.Colors.Cyan).Bold(true).Background(style.Colors.Surface).Width(innerW)
	normal := lipgloss.NewStyle().Width(innerW).Align(lipgloss.Left)
	for i, row := range rows {
		index := i + offset
		cursorCol := "  "
		if index == cursor {
			cursorCol = "> "
		}
		if field.Kind == state.FormMultiSelect {
			checkbox := "[ ] "
			if form.Popup.PendingSelected[field.Options[index]] {
				checkbox = "[x] "
			}
			prefix := cursorCol + checkbox
			cell := prefix + component.TruncateVisible(row, max(0, innerW-len(prefix)))
			cell = component.FormRow("", 0, innerW, cell)
			if index == cursor {
				cell = highlight.Render(cell)
			} else {
				cell = normal.Render(cell)
			}
			out = append(out, cell)
			continue
		} else {
			row = cursorCol + component.TruncateVisible(row, max(0, innerW-2))
		}
		cell := component.FormRow("", 0, innerW, row)
		if index == cursor {
			cell = highlight.Render(cell)
		} else {
			cell = normal.Render(cell)
		}
		out = append(out, cell)
	}
	return out
}
