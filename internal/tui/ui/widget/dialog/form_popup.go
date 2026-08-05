package dialog

import (
	"fmt"
	"strings"
	"time"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"

	"charm.land/lipgloss/v2"
)

// renderFormPopup dispatches to the popup-kind-specific renderer. Path fields
// use renderPathPopup (eza-l style listing with breadcrumb + detail box);
// Select/MultiSelect use the flat option list.
func renderFormPopup(form state.FormState, box string, dialogW, dialogH int) string {
	switch form.Popup.Kind {
	case state.PopupPath:
		return renderPathPopup(form, box, dialogW, dialogH)
	default:
		return renderSelectPopup(form, box, dialogW, dialogH)
	}
}

// renderPathPopup draws the directory/file picker popup for a FormPath field.
// Layout: header (label + breadcrumb), column headers (T MODE OWNER GROUP
// SIZE DATE NAME), candidate rows (./ ../ pinned at top, others below),
// selection detail box (full info of cursor entry, name wraps).
func renderPathPopup(form state.FormState, _ string, dialogW, _ int) string {
	field := form.PopupField()
	if field == nil {
		return ""
	}
	entries := pathEntries(field)
	popupW := dialogW - 6
	if popupW < 16 {
		popupW = 16
	}
	innerW := popupW - 4
	bodyW := innerW - 2 // reserve 2 cells for "> " cursor prefix

	if len(entries) == 0 {
		var message string
		switch {
		case field.PathError != "":
			message = lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleDialogError).GetForeground()).Render("⚠ " + field.PathError)
		case field.PathLoading:
			message = i18n.T("form.path.loading")
		default:
			message = i18n.T("form.path.no_matches")
		}
		lines := []string{
			component.FormRow("", 0, innerW, field.Label),
			component.FormRow("", 0, innerW, lipgloss.NewStyle().Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).Render(field.Input.Text)),
			component.FormRow("", 0, innerW, ""),
			component.FormRow("", 0, innerW, ""),
			component.FormRow("", 0, innerW, message),
		}
		for len(lines) < state.FormPopupVisibleRows+3 {
			lines = append(lines, component.FormRow("", 0, innerW, ""))
		}
		return DialogBox(DialogStyle{Width: popupW, TitleColor: component.GetStyle(component.StylePanelTitle).GetForeground(), LeftAligned: true}, lines...)
	}

	rows, idxMap := renderPathRows(entries, bodyW)
	visStart, visEnd := visiblePopupIndices(len(rows), form.Popup.Cursor, state.FormPopupVisibleRows)
	lines := shadePathRows(rows, idxMap, visStart, visEnd, form.Popup.Cursor, innerW, bodyW)

	header := []string{
		component.FormRow("", 0, innerW, field.Label),
		component.FormRow("", 0, innerW, lipgloss.NewStyle().Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).Render(utils.FitVisible(field.Input.Text, innerW))),
		component.FormRow("", 0, innerW, pathColumnHeader()),
	}
	lines = append(header, lines...)

	if detail := renderPathDetail(selectedEntry(entries, form.Popup.Cursor, idxMap), innerW); len(detail) > 0 {
		lines = append(lines, component.FormRow("", 0, innerW, strings.Repeat(component.BorderLineHorizontal, innerW)))
		lines = append(lines, detail...)
	}

	return DialogBox(DialogStyle{Width: popupW, TitleColor: component.GetStyle(component.StylePanelTitle).GetForeground(), LeftAligned: true}, lines...)
}

// pathEntries returns the entry list shown in the popup: ./ and ../ pinned at
// top (when not root) followed by the field's real suggestions.
func pathEntries(field *state.FormField) []state.PathEntry {
	out := make([]state.PathEntry, 0, len(field.Suggestions)+2)
	out = append(out, state.PathEntry{Name: ".", Path: "", IsDir: true, Type: state.PathEntryDir, Mode: "drwxr-xr-x"})
	out = append(out, state.PathEntry{Name: "..", Path: "", IsDir: true, Type: state.PathEntryDir, Mode: "drwxr-xr-x"})
	out = append(out, field.Suggestions...)
	return out
}

// pathColumnHeader is the header row for the eza-l column layout.
func pathColumnHeader() string {
	return "T MODE        OWNER   GROUP   SIZE    DATE         NAME"
}

// renderPathRows builds the formatted rows for each entry plus an index map
// (row index -> entry index in field.Suggestions; negative for ./ and ../ which
// are virtual and have no Suggestions slot).
func renderPathRows(entries []state.PathEntry, bodyW int) ([]string, []int) {
	rows := make([]string, 0, len(entries))
	idxMap := make([]int, 0, len(entries))
	for i, e := range entries {
		rows = append(rows, formatPathRow(e, bodyW))
		if i < 2 {
			idxMap = append(idxMap, -1) // virtual ./ and .. entries
		} else {
			idxMap = append(idxMap, i-2)
		}
	}
	return rows, idxMap
}

// formatPathRow builds one eza-l style row: T MODE OWNER GROUP SIZE DATE NAME.
func formatPathRow(e state.PathEntry, bodyW int) string {
	typeRune := "F"
	switch e.Type {
	case state.PathEntryDir:
		typeRune = "D"
	case state.PathEntryLink:
		typeRune = "L"
	}
	// Date column is fixed-width so empty Mtime does not shift the name
	// column left (BR-043 §3.3 layout consistency).
	date := formatDate(e.Mtime)
	if date == "" {
		date = strings.Repeat(" ", 12)
	}
	mode := e.Mode
	if mode == "" {
		mode = "----------"
	}
	owner := utils.PadVisible(utils.TruncateVisible(e.Owner, 5), 5)
	group := utils.PadVisible(utils.TruncateVisible(e.Group, 5), 5)
	size := utils.PadVisible(humanSize(e.Size), 7)
	name := e.Name
	if e.Type == state.PathEntryDir && name != "." && name != ".." && !strings.HasSuffix(name, "/") {
		name += "/"
	}
	prefix := fmt.Sprintf("%s %s %s %s %s %s ", typeRune, mode, owner, group, size, date)
	return component.FormRow("", 0, bodyW, prefix+name)
}

// humanSize renders sizes like ls -h: B/K/M/G/T. Empty for zero.
func humanSize(n int64) string { return utils.HumanSizeBytes(n) }

// formatDate renders a short date like "Jul 01 14:30" matching eza -l.
func formatDate(t time.Time) string { return utils.FormatShortDate(t) }

// visiblePopupIndices returns the window [start, end) of rows to show, with
// cursor centered.
func visiblePopupIndices(total, cursor, limit int) (int, int) {
	if limit <= 0 || total <= limit {
		return 0, total
	}
	start := cursor - limit/2
	if start < 0 {
		start = 0
	}
	if start+limit > total {
		start = total - limit
	}
	return start, start + limit
}

func shadePathRows(rows []string, idxMap []int, start, end, cursor, innerW, bodyW int) []string {
	highlight := lipgloss.NewStyle().Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).Bold(true).Background(component.GetStyle(component.StyleHeaderBar).GetBackground()).Width(innerW)
	normal := lipgloss.NewStyle().Width(innerW).Align(lipgloss.Left)
	out := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}
		cell := prefix + rows[i]
		cell = component.FormRow("", 0, bodyW, cell)
		cell = utils.PadVisible(cell, innerW)
		if i == cursor {
			cell = highlight.Render(cell)
		} else {
			cell = normal.Render(cell)
		}
		out = append(out, cell)
	}
	for len(out) < state.FormPopupVisibleRows {
		out = append(out, normal.Render(component.FormRow("", 0, bodyW, "")))
	}
	return out
}

// selectedEntry returns the entry the popup cursor currently points at,
// resolving virtual ./ and ../ to the Suggestions slice.
func selectedEntry(entries []state.PathEntry, cursor int, idxMap []int) *state.PathEntry {
	if cursor < 0 || cursor >= len(idxMap) {
		return nil
	}
	entryIdx := idxMap[cursor]
	if entryIdx < 0 {
		if cursor < len(entries) {
			return &entries[cursor]
		}
		return nil
	}
	return &entries[cursor]
}

// renderPathDetail renders the dedicated selection detail box. Two lines:
// Mode Owner:Group Size Date and Name (wraps if too long).
func renderPathDetail(e *state.PathEntry, innerW int) []string {
	if e == nil {
		return nil
	}
	mode := e.Mode
	if mode == "" {
		mode = "----------"
	}
	owner := e.Owner
	if owner == "" {
		owner = "-"
	}
	group := e.Group
	if group == "" {
		group = "-"
	}
	size := humanSize(e.Size)
	date := "-"
	if !e.Mtime.IsZero() {
		date = e.Mtime.Format("Jan 02 2006 15:04")
	}
	header := fmt.Sprintf("%s %s:%s %s %s", mode, owner, group, size, date)
	name := e.Name
	if e.Type == state.PathEntryDir && name != "." && name != ".." && !strings.HasSuffix(name, "/") {
		name += "/"
	}
	nameLine := "Name: " + name
	if e.Type == state.PathEntryLink && e.LinkTarget != "" {
		nameLine += " → " + e.LinkTarget
	}
	return []string{
		component.FormRow("", 0, innerW, header),
		component.FormRow("", 0, innerW, nameLine),
	}
}

// renderSelectPopup draws the flat single/multi-select option list popup.
func renderSelectPopup(form state.FormState, _ string, dialogW, _ int) string {
	field := form.PopupField()
	if field == nil {
		return ""
	}
	rows := popupRows(field)
	if len(rows) == 0 {
		return ""
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

	return DialogBox(DialogStyle{Width: popupW, TitleColor: component.GetStyle(component.StylePanelTitle).GetForeground(), LeftAligned: true}, lines...)
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
	highlight := lipgloss.NewStyle().Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).Bold(true).Background(component.GetStyle(component.StyleHeaderBar).GetBackground()).Width(innerW)
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
