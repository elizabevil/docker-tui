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

// renderFormPopup dispatches to the popup-kind-specific renderer.
func renderFormPopup(form state.FormState, box string, dialogW, dialogH int) string {
	switch form.Popup.Kind {
	case state.PopupPath:
		return renderPathPopup(form, box, dialogW, dialogH)
	default:
		return renderSelectPopup(form, box, dialogW, dialogH)
	}
}

// renderPathPopup draws the directory/file picker popup for a FormPath field.
// Path-specific state is read via type assertion to *PathField.
func renderPathPopup(form state.FormState, _ string, dialogW, _ int) string {
	field := form.PopupField()
	if field == nil {
		return ""
	}
	pf, _ := field.(*PathField)
	if pf == nil {
		return ""
	}
	entries := pathEntries(pf)
	popupW := dialogW - 6
	if popupW < 16 {
		popupW = 16
	}
	innerW := popupW - 4
	bodyW := innerW - 2

	if len(entries) == 0 {
		var message string
		switch {
		case pf.PathError() != "":
			message = lipgloss.NewStyle().Foreground(component.GetStyle(component.StyleDialogError).GetForeground()).Render("⚠ " + pf.PathError())
		case pf.PathLoading():
			message = i18n.T("form.path.loading")
		default:
			message = i18n.T("form.path.no_matches")
		}
		df, _ := field.(FormField)
		label := ""
		if df != nil {
			label = df.Label()
		}
		lines := []string{
			component.FormRow("", 0, innerW, label),
			component.FormRow("", 0, innerW, lipgloss.NewStyle().Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).Render(pf.TextRaw())),
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

	df, _ := field.(FormField)
	label := ""
	if df != nil {
		label = df.Label()
	}
	header := []string{
		component.FormRow("", 0, innerW, label),
		component.FormRow("", 0, innerW, lipgloss.NewStyle().Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).Render(utils.FitVisible(pf.TextRaw(), innerW))),
		component.FormRow("", 0, innerW, pathColumnHeader()),
	}
	lines = append(header, lines...)

	if detail := renderPathDetail(selectedEntry(entries, form.Popup.Cursor, idxMap), innerW); len(detail) > 0 {
		lines = append(lines, component.FormRow("", 0, innerW, strings.Repeat(component.BorderLineHorizontal, innerW)))
		lines = append(lines, detail...)
	}

	return DialogBox(DialogStyle{Width: popupW, TitleColor: component.GetStyle(component.StylePanelTitle).GetForeground(), LeftAligned: true}, lines...)
}

func pathEntries(pf *PathField) []state.PathEntry {
	out := make([]state.PathEntry, 0, len(pf.Suggestions())+2)
	out = append(out, state.PathEntry{Name: ".", Path: "", IsDir: true, Type: state.PathEntryDir, Mode: "drwxr-xr-x"})
	out = append(out, state.PathEntry{Name: "..", Path: "", IsDir: true, Type: state.PathEntryDir, Mode: "drwxr-xr-x"})
	out = append(out, pf.Suggestions()...)
	return out
}

func pathColumnHeader() string {
	return "T MODE        OWNER   GROUP   SIZE    DATE         NAME"
}

func renderPathRows(entries []state.PathEntry, bodyW int) ([]string, []int) {
	rows := make([]string, 0, len(entries))
	idxMap := make([]int, 0, len(entries))
	for i, e := range entries {
		rows = append(rows, formatPathRow(e, bodyW))
		if i < 2 {
			idxMap = append(idxMap, -1)
		} else {
			idxMap = append(idxMap, i-2)
		}
	}
	return rows, idxMap
}

func formatPathRow(e state.PathEntry, bodyW int) string {
	typeRune := "F"
	switch e.Type {
	case state.PathEntryDir:
		typeRune = "D"
	case state.PathEntryLink:
		typeRune = "L"
	}
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

func humanSize(n int64) string { return utils.HumanSizeBytes(n) }
func formatDate(t time.Time) string { return utils.FormatShortDate(t) }

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

func renderSelectPopup(form state.FormState, _ string, dialogW, _ int) string {
	field := form.PopupField()
	if field == nil {
		return ""
	}
	df, _ := field.(FormField)
	if df == nil {
		return ""
	}
	rows := popupRows(df)
	if len(rows) == 0 {
		return ""
	}

	popupW := dialogW - 6
	if popupW < 16 {
		popupW = 16
	}
	innerW := popupW - 4
	lines := visiblePopupRows(form, df, rows, innerW, state.FormPopupVisibleRows)
	for len(lines) < state.FormPopupVisibleRows {
		lines = append(lines, component.FormRow("", 0, innerW, ""))
	}
	label := df.Label()
	lines = append([]string{component.FormRow("", 0, innerW, label), component.FormRow("", 0, innerW, "")}, lines...)

	return DialogBox(DialogStyle{Width: popupW, TitleColor: component.GetStyle(component.StylePanelTitle).GetForeground(), LeftAligned: true}, lines...)
}

func visiblePopupRows(form state.FormState, field FormField, rows []string, innerW, limit int) []string {
	if limit <= 0 || len(rows) <= limit {
		return shadePopup(form, field, rows, innerW, 0)
	}
	start := min(max(0, form.Popup.Cursor-limit/2), len(rows)-limit)
	return shadePopup(form, field, rows[start:start+limit], innerW, start)
}

func popupRows(field FormField) []string {
	switch field.Kind() {
	case state.FormMultiSelect, state.FormSelect:
		return field.Options()
	case state.FormPath:
		pf, _ := field.(*PathField)
		if pf == nil {
			return nil
		}
		out := make([]string, 0, len(pf.Suggestions()))
		for i := range pf.Suggestions() {
			out = append(out, pathRowLabel(&pf.Suggestions()[i]))
		}
		return out
	default:
		return nil
	}
}

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

func shadePopup(form state.FormState, field FormField, rows []string, innerW, offset int) []string {
	out := make([]string, 0, len(rows))
	cursor := form.Popup.Cursor
	highlight := lipgloss.NewStyle().Foreground(component.GetStyle(component.StylePanelTitle).GetForeground()).Bold(true).Background(component.GetStyle(component.StyleHeaderBar).GetBackground()).Width(innerW)
	normal := lipgloss.NewStyle().Width(innerW).Align(lipgloss.Left)
	options := field.Options()
	for i, row := range rows {
		index := i + offset
		cursorCol := "  "
		if index == cursor {
			cursorCol = "> "
		}
		if field.Kind() == state.FormMultiSelect {
			checkbox := "[ ] "
			if form.Popup.PendingSelected[options[index]] {
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