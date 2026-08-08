package view

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/elizabevil/docker-tui/internal/tui/filter"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/mattn/go-runewidth"
)

const (
	headerRailHeight   = 4
	messageRailHeight  = 2
	queryRailHeight    = 3
	footerRailHeight   = 2
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
	if m == nil {
		return ""
	}

	text := ""
	var style lipgloss.Style
	switch {
	case m.Feedback.BatchProgress != nil:
		progress := m.Feedback.BatchProgress
		indicator := component.NewBatchProgressIndicator(progress.Action, progress.Current, progress.Total)
		railStyle := component.GetStyle(component.StyleMessageRail).
			Foreground(component.GetStyle(component.StyleToastInfo).GetForeground())
		return railStyle.Render(wrapMessageRail(indicator.Render(max(1, width)), max(1, width), messageRailHeight))
	case m.ContainerWait.IsActive():
		indicator := component.NewWaitIndicator(m.ContainerWait.TargetID, m.ContainerWait.Generation)
		railStyle := component.GetStyle(component.StyleMessageRail).
			Foreground(component.GetStyle(component.StyleToastWarning).GetForeground())
		return railStyle.Render(wrapMessageRail(indicator.Render(max(1, width)), max(1, width), messageRailHeight))
	case m.Feedback.ErrorMessage != "":
		style = component.GetStyle(component.StyleToastError)
		text = m.Feedback.ErrorMessage
		if m.Feedback.ErrorCount > 1 {
			text += fmt.Sprintf(" [%d]", m.Feedback.ErrorCount)
		}
	case m.Feedback.ToastMessage != "":
		style = component.GetStyle(levelStyle(m.Feedback.ToastLevel))
		text = m.Feedback.ToastMessage
	case m.Feedback.AuditOperationMessage != "":
		style = component.GetStyle(component.StyleToastInfo)
		text = m.Feedback.AuditOperationMessage
	case m.Feedback.InfoMessage != "":
		style = component.GetStyle(component.StyleToastInfo)
		text = m.Feedback.InfoMessage
	default:
		return ""
	}
	railStyle := component.GetStyle(component.StyleMessageRail).Foreground(style.GetForeground())
	return railStyle.Render(wrapMessageRail(text, max(1, width), messageRailHeight))
}

func levelStyle(level state.NotificationLevel) component.StyleName {
	switch level {
	case state.NotificationSuccess:
		return component.StyleToastSuccess
	case state.NotificationError:
		return component.StyleToastError
	case state.NotificationWarning:
		return component.StyleToastWarning
	default:
		return component.StyleToastInfo
	}
}

func wrapMessageRail(text string, width, maxLines int) string {
	if text == "" || maxLines <= 0 {
		return ""
	}
	width = max(1, width)

	wrapped := make([]string, 0, maxLines+1)
	for logicalLine := range strings.SplitSeq(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if logicalLine == "" {
			wrapped = append(wrapped, "")
			continue
		}
		for logicalLine != "" {
			prefix, rest := splitVisiblePrefix(logicalLine, width)
			wrapped = append(wrapped, prefix)
			logicalLine = rest
		}
	}

	if len(wrapped) <= maxLines {
		return strings.Join(wrapped, "\n")
	}
	overflow := strings.Join(wrapped[maxLines-1:], " ")
	wrapped = wrapped[:maxLines]
	wrapped[maxLines-1] = component.TruncateVisible(overflow, width)
	return strings.Join(wrapped, "\n")
}

func splitVisiblePrefix(text string, width int) (string, string) {
	used := 0
	for index, r := range text {
		cellWidth := runewidth.RuneWidth(r)
		if used+cellWidth > width {
			if index == 0 {
				_, size := utf8.DecodeRuneInString(text)
				return text[:size], text[size:]
			}
			return text[:index], text[index:]
		}
		used += cellWidth
	}
	return text, ""
}

func renderQueryRail(m *state.AppModel, width int) string {
	if m == nil {
		return ""
	}
	kind := queryKindFor(m)
	switch kind {
	case queryFilter:
		return renderQueryInput(kind, m.Navigation.FilterInput.Text, m.Navigation.FilterInput.Cursor, width, filter.New(m).MatchCount(), !m.CursorBlinkHidden)
	case querySearch:
		return renderQueryInput(kind, m.Navigation.SearchInput.Text, m.Navigation.SearchInput.Cursor, width, "", !m.CursorBlinkHidden)
	case queryCommand:
		return renderQueryInput(kind, m.Navigation.CommandInput.Text, m.Navigation.CommandInput.Cursor, width, "", !m.CursorBlinkHidden)
	case queryImagePull:
		return renderQueryInput(kind, m.Dialog.Input.Text, m.Dialog.Input.Cursor, width, "", !m.CursorBlinkHidden)
	default:
		return ""
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
	if m.Navigation.Mode == state.ModeCommand {
		return queryCommand
	}
	if m.Navigation.Mode == state.ModeImagePull {
		return queryImagePull
	}
	if m.Navigation.Mode == state.ModeSearch {
		return querySearch
	}
	if m.Navigation.Mode == state.ModeFilter {
		return queryFilter
	}
	return queryNone
}

func renderQueryInput(kind queryKind, text string, cursor int, width int, matchCount string, cursorVisible ...bool) string {
	if kind == queryNone {
		return ""
	}
	boxWidth := max(16, width-4)
	innerWidth := max(8, boxWidth-4)

	var input string
	visible := len(cursorVisible) == 0 || cursorVisible[0]
	switch kind {
	case queryCommand:
		prefix := component.GetStyle(component.StyleCommandPrefix).Render(": ")
		suffix := component.AutocompleteSuffix(text)
		if suffix != "" {
			typed := component.GetStyle(component.StyleSearchBar).Render(text)
			hint := component.GetStyle(component.StyleSearchHint).Render(suffix)
			cursorMark := component.BlockCursor
			if !visible {
				cursorMark = " "
			}
			input = prefix + typed + cursorMark + hint
		} else {
			input = prefix + component.GetStyle(component.StyleSearchBar).Render(insertCursor(text, cursor, visible))
		}
	case querySearch:
		input = component.GetStyle(component.StyleSearchBar).Render("Search: " + insertCursor(text, cursor, visible))
	case queryImagePull:
		input = component.GetStyle(component.StyleSearchBar).Render("Pull: " + insertCursor(text, cursor, visible))
	default:
		countSuffix := ""
		if matchCount != "" {
			countSuffix = " " + component.GetStyle(component.StyleDim).Render(matchCount)
		}
		input = component.GetStyle(component.StyleSearchBar).Render("Filter: " + insertCursor(text, cursor, visible) + countSuffix)
	}

	content := component.PadVisible(input, innerWidth)
	// Bordered box: border top + single content row + border bottom =
	// queryRailHeight rows, so fitRailHeight does not truncate it. The
	// cursor sits on the content row and the OS-level IME candidate
	// window (drawn at cursor_row + 1) lands on the border row, which
	// is still inside the query rail and clear of the panel.
	return component.GetStyle(component.StyleQueryBar).
		Width(boxWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(component.GetStyle(component.StyleDim).GetForeground()).
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
