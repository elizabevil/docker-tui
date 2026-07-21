package footer

import (
	"fmt"

	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
	"github.com/elizabevil/docker-tui/internal/tui/utils"
)

// OperationLogLine returns a fixed-height operation log line for footer.
// Empty content is rendered as an empty line to keep footer height stable.
func OperationLogLine(app *state.AppModel) string {
	if app == nil {
		return ""
	}
	if app.ErrorMessage != "" {
		return "LOG: " + utils.Truncate(app.ErrorMessage, 96)
	}
	if app.AuditOperationMessage != "" {
		return "LOG: " + utils.Truncate(app.AuditOperationMessage, 96)
	}
	if app.InfoMessage != "" {
		return "LOG: " + utils.Truncate(app.InfoMessage, 96)
	}
	return ""
}

func operationLogStatus(app *state.AppModel) string {
	if app == nil {
		return ""
	}
	if app.ErrorMessage != "" {
		short := app.ErrorMessage
		if len(short) > 40 {
			short = short[:37] + "..."
		}
		if app.ErrorCount > 1 {
			short += fmt.Sprintf(" [%d]", app.ErrorCount)
		}
		return component.GetStyle("toastError").Render(short)
	}
	if app.AuditOperationMessage != "" {
		return component.GetStyle("toastInfo").Render(utils.Truncate(app.AuditOperationMessage, 40))
	}
	if app.InfoMessage != "" {
		short := app.InfoMessage
		if len(short) > 40 {
			short = short[:37] + "..."
		}
		return component.GetStyle("toastInfo").Render(short)
	}
	return ""
}
