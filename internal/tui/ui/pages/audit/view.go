package audit

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/tables"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"
)

var tc = tables.MustLoad("audit")

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

	availW := max(50, width)
	colsDef := tc.Columns.Get("default")
	rowHeight := component.CalcTableRowHeight(panelHeight, false)
	component.EnsureVisible(&auditState.ViewOffset, auditState.Cursor, rowHeight, total)
	viewOffset := auditState.ViewOffset

	rows := component.BuildRows(records, colsDef, viewOffset, rowHeight,
		func(record audit.Record, cd tables.ColumnDef, _ int) string {
			return auditCellValue(cd.Key, &record)
		})

	levelHint := ""
	switch auditState.FilterLevel {
	case state.AuditFilterErrors:
		levelHint = " [Errors]"
	case state.AuditFilterWarnings:
		levelHint = " [Warnings]"
	}
	hint := fmt.Sprintf("%d records%s", total, levelHint)
	if filter := auditState.FilterText(); filter != "" {
		hint += " " + component.BorderLineVertical + " Filter: " + filter
	}

	return component.RenderTable(component.TableData{
		Cols:       colsDef,
		Rows:       rows,
		Selected:   auditState.Cursor - viewOffset,
		Total:      total,
		Offset:     viewOffset,
		Limit:      rowHeight,
		BannerW:    availW,
		BodyHeight: panelHeight,
		FooterHint: hint,
		ColStyles:  component.GetColumnStyles(colsDef),
	})
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
			return utils.ShortID(r.Target.ID)
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
