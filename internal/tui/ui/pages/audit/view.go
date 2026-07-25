package audit

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

// RenderList renders the audit history table view.
func RenderList(auditState *state.AuditState, width int, panelHeight int) string {
	if auditState == nil {
		return i18n.T("msg.loading")
	}

	records := auditState.AuditListModel()
	total := len(records)
	if total == 0 {
		if auditState.FilterText() != "" {
			return component.GetStyle("dim").Render("No audit records match filter")
		}
		return component.GetStyle("dim").Render("No audit records")
	}

	// Ensure cursor is in bounds
	if auditState.Cursor >= total {
		auditState.Cursor = total - 1
	}
	if auditState.Cursor < 0 {
		auditState.Cursor = 0
	}

	// Column definitions
	timeW := 19
	actionW := 14
	targetW := 12
	resultW := 12
	levelW := 8
	availW := width - 8
	if availW < 50 {
		availW = 50
	}

	// Distribute remaining width to message column
	usedW := timeW + actionW + targetW + resultW + levelW
	msgW := availW - usedW
	if msgW < 10 {
		msgW = 10
	}

	cols := []auditColumn{
		{key: "time", width: timeW},
		{key: "action", width: actionW},
		{key: "target", width: targetW},
		{key: "result", width: resultW},
		{key: "level", width: levelW},
		{key: "message", width: msgW},
	}

	// Header
	header := auditHeader(cols)

	// Rows
	rowHeight := component.CalcRowHeight(panelHeight)
	viewOffset := auditState.ViewOffset
	component.EnsureVisible(&viewOffset, auditState.Cursor, rowHeight, total)

	var rows []string
	for i := viewOffset; i < total && len(rows) < rowHeight; i++ {
		r := records[i]
		row := auditRow(cols, &r, i == auditState.Cursor)
		rows = append(rows, row)
	}

	// Build hint
	levelHint := ""
	switch auditState.FilterLevel {
	case state.AuditFilterErrors:
		levelHint = " [Errors]"
	case state.AuditFilterWarnings:
		levelHint = " [Warnings]"
	}
	hint := fmt.Sprintf("%d records%s", total, levelHint)

	// Assemble
	var sb strings.Builder
	if auditState.FilterText() != "" {
		sb.WriteString(component.GetStyle("dim").Render("Filter: " + auditState.FilterText()))
		sb.WriteString("\n")
	}
	sb.WriteString(header)
	sb.WriteString("\n")
	for _, row := range rows {
		sb.WriteString(row)
		sb.WriteString("\n")
	}
	// Pad remaining rows
	for len(rows) < rowHeight {
		sb.WriteString(strings.Repeat(" ", availW))
		sb.WriteString("\n")
		rows = append(rows, "")
	}
	sb.WriteString(component.GetStyle("dim").Render(hint))

	return sb.String()
}

// RenderDetail renders a single audit record in detail view.
func RenderDetail(record *audit.Record, width int, height int) string {
	if record == nil {
		return component.GetStyle("dim").Render("No record selected")
	}

	var sb strings.Builder
	w := width - 4
	if w < 40 {
		w = 40
	}

	sb.WriteString(component.GetStyle("header").Render("Audit Record Detail"))
	sb.WriteString("\n\n")

	// Time
	sb.WriteString(detailRow("Time", record.Time.Format(time.RFC3339), w))
	sb.WriteString("\n")

	// Trace ID (truncated)
	traceID := record.TraceID
	if len(traceID) > 16 {
		traceID = traceID[:16] + "..."
	}
	sb.WriteString(detailRow("Trace ID", traceID, w))
	sb.WriteString("\n")

	// Action
	sb.WriteString(detailRow("Action", record.Action, w))
	sb.WriteString("\n")

	// Result
	resultSt := resolveResultStyle(record.Result)
	sb.WriteString(detailRowStyled("Result", string(record.Result), resultSt, w))
	sb.WriteString("\n")

	// Level
	levelSt := resolveLevelStyle(record.Level)
	sb.WriteString(detailRowStyled("Level", string(record.Level), levelSt, w))
	sb.WriteString("\n")

	// Target
	sb.WriteString(detailRow("Target Type", record.Target.Type, w))
	sb.WriteString("\n")
	if record.Target.ID != "" {
		sb.WriteString(detailRow("Target ID", record.Target.ID, w))
		sb.WriteString("\n")
	}
	if record.Target.Name != "" {
		sb.WriteString(detailRow("Target Name", record.Target.Name, w))
		sb.WriteString("\n")
	}

	// Message
	if record.Message != "" {
		sb.WriteString("\n")
		sb.WriteString(component.GetStyle("detailSection").Render("Message"))
		sb.WriteString("\n")
		sb.WriteString(component.GetStyle("detailValue").Render("  " + record.Message))
		sb.WriteString("\n")
	}

	// Runtime
	if record.Runtime.Type != "" {
		sb.WriteString("\n")
		sb.WriteString(component.GetStyle("detailSection").Render("Runtime"))
		sb.WriteString("\n")
		sb.WriteString(detailRow("Type", record.Runtime.Type, w))
		sb.WriteString("\n")
		if record.Runtime.Host != "" {
			sb.WriteString(detailRow("Host", record.Runtime.Host, w))
			sb.WriteString("\n")
		}
	}

	// UI context
	if record.UI.Surface != "" {
		sb.WriteString("\n")
		sb.WriteString(component.GetStyle("detailSection").Render("UI Context"))
		sb.WriteString("\n")
		sb.WriteString(detailRow("Surface", record.UI.Surface, w))
		sb.WriteString("\n")
		if record.UI.View != "" {
			sb.WriteString(detailRow("View", record.UI.View, w))
			sb.WriteString("\n")
		}
	}

	// Details
	if record.Details.DurationMs > 0 || record.Details.Error != "" {
		sb.WriteString("\n")
		sb.WriteString(component.GetStyle("detailSection").Render("Details"))
		sb.WriteString("\n")
		if record.Details.DurationMs > 0 {
			sb.WriteString(detailRow("Duration", fmt.Sprintf("%dms", record.Details.DurationMs), w))
			sb.WriteString("\n")
		}
		if record.Details.Error != "" {
			sb.WriteString(detailRow("Error", record.Details.Error, w))
			sb.WriteString("\n")
		}
		if record.Details.ExitCode != nil {
			sb.WriteString(detailRow("Exit Code", fmt.Sprintf("%d", *record.Details.ExitCode), w))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func detailRow(label, value string, width int) string {
	return detailRowStyled(label, value, component.GetStyle("detailValue"), width)
}

func detailRowStyled(label, value string, valueStyle lipgloss.Style, width int) string {
	labelStr := component.GetStyle("detailLabel").Render(fmt.Sprintf("%-14s", label))
	valStr := valueStyle.Render(value)
	return "  " + labelStr + "  " + valStr
}

type auditColumn struct {
	key   string
	width int
}

func auditHeader(cols []auditColumn) string {
	var parts []string
	for _, c := range cols {
		title := strings.ToUpper(c.key)
		if len(title) > c.width {
			title = title[:c.width]
		}
		parts = append(parts, component.GetStyle("header").Render(fmt.Sprintf("%-*s", c.width, title)))
	}
	return strings.Join(parts, " ")
}

func auditRow(cols []auditColumn, r *audit.Record, selected bool) string {
	var parts []string
	for _, c := range cols {
		val := auditCellValue(c.key, r)
		if len(val) > c.width {
			val = val[:c.width-1] + "\u2026"
		}
		if selected {
			parts = append(parts, component.GetStyle("rowSelected").Render(fmt.Sprintf("%-*s", c.width, val)))
		} else {
			parts = append(parts, fmt.Sprintf("%-*s", c.width, val))
		}
	}
	return strings.Join(parts, " ")
}

func auditCellValue(key string, r *audit.Record) string {
	switch key {
	case "time":
		return r.Time.Format("2006-01-02 15:04:05")
	case "action":
		return r.Action
	case "target":
		if r.Target.Name != "" {
			return r.Target.Name
		}
		if r.Target.ID != "" {
			if len(r.Target.ID) > 12 {
				return r.Target.ID[:12]
			}
			return r.Target.ID
		}
		return r.Target.Type
	case "result":
		return string(r.Result)
	case "level":
		return string(r.Level)
	case "message":
		return r.Message
	default:
		return ""
	}
}

func resolveResultStyle(result audit.Result) lipgloss.Style {
	switch result {
	case audit.ResultSucceeded:
		return component.GetStyle("toastSuccess")
	case audit.ResultFailed:
		return component.GetStyle("toastError")
	case audit.ResultCancelled:
		return component.GetStyle("toastWarning")
	default:
		return component.GetStyle("dim")
	}
}

func resolveLevelStyle(level audit.Level) lipgloss.Style {
	switch level {
	case audit.LevelError:
		return component.GetStyle("toastError")
	case audit.LevelWarn:
		return component.GetStyle("toastWarning")
	default:
		return component.GetStyle("dim")
	}
}
