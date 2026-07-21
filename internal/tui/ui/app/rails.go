package view

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

const (
	minimumTerminalWidth  = 80
	minimumTerminalHeight = 20

	headerRailHeight   = 4
	messageRailHeight  = 1
	queryRailHeight    = 3
	footerRailHeight   = 3
	minimumPanelHeight = 4
)

type railHeights struct {
	header  int
	message int
	query   int
	panel   int
	footer  int
}

func calculateRailHeights(usableHeight int) railHeights {
	fixedHeight := headerRailHeight + messageRailHeight + queryRailHeight + footerRailHeight
	return railHeights{
		header:  headerRailHeight,
		message: messageRailHeight,
		query:   queryRailHeight,
		panel:   max(minimumPanelHeight, usableHeight-fixedHeight),
		footer:  footerRailHeight,
	}
}

func (h railHeights) total() int {
	return h.header + h.message + h.query + h.panel + h.footer
}

func panelBodyHeight(panelHeight int) int {
	const panelOverhead = 3 // border(2) + title(1)
	return max(1, panelHeight-panelOverhead)
}

func renderMessageRail(m *state.AppModel, width int) string {
	if m == nil || m.ToastMessage == "" {
		return ""
	}

	var style lipgloss.Style
	switch m.ToastLevel {
	case component.ToastSuccess:
		style = component.GetStyle("toastSuccess")
	case component.ToastError:
		style = component.GetStyle("toastError")
	case component.ToastWarning:
		style = component.GetStyle("toastWarning")
	default:
		style = component.GetStyle("toastInfo")
	}
	return style.Render(component.TruncateVisible(m.ToastMessage, max(1, width)))
}

func renderQueryRail(m *state.AppModel, width int) string {
	if m == nil {
		return ""
	}
	kind := queryKindFor(m)
	switch kind {
	case queryFilter:
		return renderQueryInput(kind, m.FilterInput.Text, m.FilterInput.Cursor, width)
	case querySearch:
		return renderQueryInput(kind, m.SearchInput.Text, m.SearchInput.Cursor, width)
	default:
		return renderQueryInput(kind, m.FilterText, m.FilterCursor, width)
	}
}

type queryKind int

const (
	queryNone queryKind = iota
	queryFilter
	querySearch
	queryCommand
	queryImagePull
)

func queryKindFor(m *state.AppModel) queryKind {
	if m == nil {
		return queryNone
	}
	if m.Mode == state.ModeCommand {
		return queryCommand
	}
	if m.Mode == state.ModeImagePull {
		return queryImagePull
	}
	if m.Mode == state.ModeSearch {
		return querySearch
	}
	if m.Mode == state.ModeFilter {
		return queryFilter
	}
	return queryNone
}

func renderQueryInput(kind queryKind, text string, cursor int, width int) string {
	if kind == queryNone {
		return ""
	}
	boxWidth := max(16, width-4)
	innerWidth := max(8, boxWidth-4)

	var input string
	switch kind {
	case queryCommand:
		prefix := component.GetStyle("commandPrefix").Render(": ")
		suffix := component.AutocompleteSuffix(text)
		if suffix != "" {
			typed := component.GetStyle("searchBar").Render(text)
			hint := component.GetStyle("searchHint").Render(suffix)
			input = prefix + typed + "\u2588" + hint
		} else {
			input = prefix + component.GetStyle("searchBar").Render(insertCursor(text, cursor))
		}
	case querySearch:
		input = component.GetStyle("searchBar").Render("Search: " + insertCursor(text, cursor))
	case queryImagePull:
		input = component.GetStyle("searchBar").Render("Pull: " + insertCursor(text, cursor))
	default:
		input = component.GetStyle("searchBar").Render("Filter: " + insertCursor(text, cursor))
	}

	content := component.PadVisible(input, innerWidth)
	return lipgloss.NewStyle().
		Width(boxWidth-2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(component.GetStyle("dim").GetForeground()).
		Padding(0, 1).
		Render(content)
}

func terminalSizeMessage(width, height int) string {
	return fmt.Sprintf(
		"Terminal too small: %dx%d (minimum %dx%d)",
		width, height, minimumTerminalWidth, minimumTerminalHeight,
	)
}

func fitRailHeight(content string, height int) string {
	if height <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if content == "" {
		lines = []string{""}
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
