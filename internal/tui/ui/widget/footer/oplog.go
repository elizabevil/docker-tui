package footer

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// OperationLogLine returns a fixed-height operation log line for footer.
// Empty content is rendered as an empty line to keep footer height stable.
func OperationLogLine(app *state.AppModel) string {
	if app == nil {
		return ""
	}
	if app.Feedback.ErrorMessage != "" {
		return "LOG: " + utils.Truncate(app.Feedback.ErrorMessage, 96)
	}
	if app.Feedback.AuditOperationMessage != "" {
		return "LOG: " + utils.Truncate(app.Feedback.AuditOperationMessage, 96)
	}
	if app.Feedback.InfoMessage != "" {
		return "LOG: " + utils.Truncate(app.Feedback.InfoMessage, 96)
	}
	return ""
}

func operationLogStatus(app *state.AppModel) string {
	if app == nil {
		return ""
	}
	if app.Feedback.ErrorMessage != "" {
		short := app.Feedback.ErrorMessage
		if len(short) > 40 {
			short = short[:37] + "..."
		}
		if app.Feedback.ErrorCount > 1 {
			short += fmt.Sprintf(" [%d]", app.Feedback.ErrorCount)
		}
		return component.GetStyle(component.StyleToastError).Render(short)
	}
	if app.Feedback.AuditOperationMessage != "" {
		return component.GetStyle(component.StyleToastInfo).Render(utils.Truncate(app.Feedback.AuditOperationMessage, 40))
	}
	if app.Feedback.InfoMessage != "" {
		short := app.Feedback.InfoMessage
		if len(short) > 40 {
			short = short[:37] + "..."
		}
		return component.GetStyle(component.StyleToastInfo).Render(short)
	}
	return ""
}
