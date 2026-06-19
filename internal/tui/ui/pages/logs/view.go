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
func RenderView(m *state.AppModel, panelHeight int) string {
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
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.LogViewOffset > maxOffset {
		m.LogViewOffset = maxOffset
	}

	visible := lines[m.LogViewOffset:]
	if len(visible) > rowHeight {
		visible = visible[:rowHeight]
	}

	lineNumStyle := component.GetStyle("dim")
	timestampStyle := component.GetStyle("logTimestamp")
	logTextStyle := component.GetStyle("logText")
	stderrStyle := component.GetStyle("logStderr")
	searchHighlight := component.GetStyle("logHighlightBg")

	logW := m.Width - 10 // available width for log content

	var rendered []string
	for i, raw := range visible {
		lineNum := m.LogViewOffset + i + 1
		numStr := fmt.Sprintf("%4d", lineNum)

		timestamp, text := splitLogLine(raw)

		// Word wrap
		if m.LogWrapEnabled && len(text) > logW {
			text = wrapLine(text, logW)
		}

		// Determine style
		textStyle := logTextStyle
		if strings.Contains(strings.ToLower(text), "stderr") {
			textStyle = stderrStyle
		}

		// Search highlight
		display := fmt.Sprintf("%s %s %s",
			lineNumStyle.Render(numStr),
			timestampStyle.Render(timestamp),
			textStyle.Render(text),
		)
		if m.LogSearchText != "" && strings.Contains(raw, m.LogSearchText) {
			display = searchHighlight.Render(fmt.Sprintf("%s %s %s", numStr, timestamp, text))
		}

		rendered = append(rendered, display)
	}

	footer := fmt.Sprintf(" %d-%d/%d │ j/k scroll │ / search │ n/N next │ w wrap │ Esc back",
		m.LogViewOffset+1,
		m.LogViewOffset+len(visible),
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

// wrapLine wraps text at maxWidth, inserting newlines.
func wrapLine(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	var result strings.Builder
	for len(text) > 0 {
		if len(text) <= maxWidth {
			result.WriteString(text)
			break
		}
		result.WriteString(text[:maxWidth])
		result.WriteByte('\n')
		text = text[maxWidth:]
	}
	return result.String()
}
