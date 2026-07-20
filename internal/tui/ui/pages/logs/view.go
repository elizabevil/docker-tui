package logs

import (
	_ "embed"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

//go:embed logs.jsonc
var logsDefaultData []byte

// logsConfig maps the JSONC structure for log view defaults.
type logsConfig struct {
	DefaultSince   string `json:"defaultSince"`
	DefaultTail    string `json:"defaultTail"`
	ShowTimestamps bool   `json:"showTimestamps"`
}

func (c *logsConfig) normalize() {
	if c.DefaultSince == "" {
		c.DefaultSince = "1h"
	}
	if c.DefaultTail == "" {
		c.DefaultTail = "200"
	}
}

// DefaultLogsConfig returns default values parsed from the embedded logs.jsonc.
func DefaultLogsConfig() logsConfig {
	loader := component.ConfigLoader[logsConfig]{
		RawData: logsDefaultData,
		Fallback: logsConfig{
			DefaultSince:   "1h",
			DefaultTail:    "200",
			ShowTimestamps: false,
		},
		Normalize: func(c *logsConfig) {
			c.normalize()
		},
	}
	return loader.Load()
}

// RenderView renders the log streaming view with search highlight and word wrap.
func RenderView(m *state.AppModel, panelHeight, panelWidth int) string {
	lines := m.LogContent
	headerExtras := fmt.Sprintf("  %d lines", len(lines))
	if m.LogSearchText != "" {
		headerExtras += fmt.Sprintf(" │ search: \"%s\"", m.LogSearchText)
	}
	if m.LogWrapEnabled {
		headerExtras += " │ wrap"
	}
	header := component.GetStyle("dim").Render(headerExtras)

	rowHeight := panelHeight - 2
	if rowHeight < 1 {
		rowHeight = 1
	}

	if m.LogContainerID == "" {
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

	if m.LogViewOffset < 0 {
		m.LogViewOffset = 0
	}
	maxOffset := len(lines) - rowHeight
	if m.LogWrapEnabled {
		maxOffset = len(lines) - 1
	}
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.LogViewOffset > maxOffset {
		m.LogViewOffset = maxOffset
	}

	lineNumStyle := component.GetStyle("dim")
	timestampStyle := component.GetStyle("logTimestamp")
	logTextStyle := component.GetStyle("logText")
	stderrStyle := component.GetStyle("logStderr")
	searchHighlight := component.GetStyle("logHighlightBg")

	const prefixWidth = 38 // line number + timestamp + separators
	logWidth := max(8, panelWidth-prefixWidth)

	rendered := make([]string, 0, rowHeight)
	consumed := 0
	for lineIndex := m.LogViewOffset; lineIndex < len(lines) && len(rendered) < rowHeight; lineIndex++ {
		raw := lines[lineIndex]
		lineNum := lineIndex + 1
		numStr := fmt.Sprintf("%4d", lineNum)

		timestamp, text := splitLogLine(raw)
		segments := []string{text}
		if m.LogWrapEnabled {
			segments = wrapLine(text, logWidth)
		}

		textStyle := logTextStyle
		if strings.Contains(strings.ToLower(text), "stderr") {
			textStyle = stderrStyle
		}
		matched := m.LogSearchText != "" && strings.Contains(strings.ToLower(raw), strings.ToLower(m.LogSearchText))
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

	footer := fmt.Sprintf(" %d-%d/%d │ j/k scroll │ / search │ n/N next │ w wrap │ Esc back",
		m.LogViewOffset+1,
		m.LogViewOffset+consumed,
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
	body := lipgloss.NewStyle().Height(bodyHeight).MaxHeight(bodyHeight).Render(strings.Join(bodyLines, "\n"))
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
