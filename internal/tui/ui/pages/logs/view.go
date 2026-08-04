package logs

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/ui/style"
)

// RenderView renders the log streaming view with search highlight and word wrap.
func RenderView(m *state.AppModel, panelHeight, panelWidth int) string {
	lines := m.Log.LogContent
	headerExtras := fmt.Sprintf("  %d lines", len(lines))
	if m.Log.LogSearchText != "" {
		headerExtras += fmt.Sprintf(" " + component.BorderLineVertical + " search: \"%s\"", m.Log.LogSearchText)
	}
	if m.Log.LogWrapEnabled {
		headerExtras += " " + component.BorderLineVertical + " wrap"
	}
	header := component.GetStyle("dim").Render(headerExtras)

	rowHeight := panelHeight - 2
	if rowHeight < 1 {
		rowHeight = 1
	}

	if m.Log.LogContainerID == "" {
		return renderLogPanel(header, []string{
			component.GetStyle("dim").Render("  Select a container and press 'l' to view logs"),
			component.GetStyle("dim").Render("  Press Esc to return to container list"),
		}, component.GetStyle("dim").Render(" Esc back"), rowHeight)
	}

	if len(lines) == 0 {
		return renderLogPanel(header, []string{
			component.GetStyle("dim").Render("  No log output from container"),
		}, component.GetStyle("dim").Render(" Esc back"), rowHeight)
	}

	visibleRows := rowHeight
	if m.Log.LogWrapEnabled {
		visibleRows = 1
	}
	offset := m.Log.VisibleOffset(len(lines), visibleRows)

	lineNumStyle := component.GetStyle("dim")
	timestampStyle := component.GetStyle("logTimestamp")
	logTextStyle := component.GetStyle("logText")
	stderrStyle := component.GetStyle("logStderr")
	searchHighlight := component.GetStyle("logHighlightBg")

	const prefixWidth = 38 // line number + timestamp + separators
	logWidth := max(8, panelWidth-prefixWidth)

	rendered := make([]string, 0, rowHeight)
	consumed := 0
	for lineIndex := offset; lineIndex < len(lines) && len(rendered) < rowHeight; lineIndex++ {
		raw := lines[lineIndex]
		lineNum := lineIndex + 1
		numStr := fmt.Sprintf("%4d", lineNum)

		timestamp, text := splitLogLine(raw)
		segments := []string{text}
		if m.Log.LogWrapEnabled {
			segments = wrapLine(text, logWidth)
		}

		textStyle := logTextStyle
		if strings.Contains(strings.ToLower(text), "stderr") {
			textStyle = stderrStyle
		}
		matched := m.Log.LogSearchText != "" && strings.Contains(strings.ToLower(raw), strings.ToLower(m.Log.LogSearchText))
		for segmentIndex, segment := range segments {
			if len(rendered) >= rowHeight {
				break
			}
			lineNumber, lineTimestamp := numStr, timestamp
			if segmentIndex > 0 {
				lineNumber, lineTimestamp = "    ", ""
			}
			display := fmt.Sprintf("%s %s %s",
				lineNumStyle.Render(lineNumber),
				timestampStyle.Render(lineTimestamp),
				textStyle.Render(segment),
			)
			if matched {
				display = searchHighlight.Render(fmt.Sprintf("%s %s %s", lineNumber, lineTimestamp, segment))
			}
			rendered = append(rendered, display)
		}
		consumed++
	}

	footer := fmt.Sprintf(" %d-%d/%d " + component.BorderLineVertical + " j/k scroll " + component.BorderLineVertical + " / search " + component.BorderLineVertical + " n/N next " + component.BorderLineVertical + " w wrap " + component.BorderLineVertical + " Esc back",
		offset+1,
		offset+consumed,
		len(lines),
	)

	return renderLogPanel(header, rendered, component.GetStyle("dim").Render(footer), rowHeight)
}

func renderLogPanel(header string, bodyLines []string, footer string, bodyHeight int) string {
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	for len(bodyLines) < bodyHeight {
		bodyLines = append(bodyLines, "")
	}
	if len(bodyLines) > bodyHeight {
		bodyLines = bodyLines[:bodyHeight]
	}
	body := lipgloss.NewStyle().Height(bodyHeight).MaxHeight(bodyHeight).Background(style.Colors.BG).Render(strings.Join(bodyLines, "\n"))
	return lipgloss.JoinVertical(lipgloss.Top, header, body, footer)
}

func splitLogLine(raw string) (timestamp, text string) {
	if len(raw) >= 30 && raw[0] == '2' && raw[4] == '-' && raw[10] == 'T' {
		return raw[:30], raw[30:]
	}
	if len(raw) >= 8 && raw[2] == ':' && raw[5] == ':' {
		return raw[:8], raw[8:]
	}
	return "", raw
}

// wrapLine returns display segments without processing content outside the viewport.
func wrapLine(text string, maxWidth int) []string {
	if maxWidth < 1 {
		maxWidth = 1
	}
	runes := []rune(text)
	if len(runes) == 0 {
		return []string{""}
	}
	segments := make([]string, 0, (len(runes)+maxWidth-1)/maxWidth)
	for len(runes) > 0 {
		end := min(maxWidth, len(runes))
		segments = append(segments, string(runes[:end]))
		runes = runes[end:]
	}
	return segments
}
