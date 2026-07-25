package update

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestContainerActionResultCompletesAuditProjection(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Dependencies.Audit = audit.NewService(nil)
	trace := app.Dependencies.Audit.Begin("resource.container.start", audit.ContainerTarget{ID: "short", Name: "api"}, audit.RuntimeContext{}, audit.UIContext{}, "Starting api")

	updated, _ := handleContainerActioned(app, state.ContainerActioned{Action: state.ActionStarted, ID: "short", Success: true, Audit: trace})
	records := updated.Dependencies.Audit.RecentOperations()
	if len(records) != 3 || records[2].Result != audit.ResultSucceeded || records[2].TraceID != trace.ID {
		t.Fatalf("action records=%#v", records)
	}
	if updated.Feedback.ToastMessage == "" || updated.Feedback.AuditOperationMessage == "" {
		t.Fatalf("audit projection missing: toast=%q operation=%q", updated.Feedback.ToastMessage, updated.Feedback.AuditOperationMessage)
	}
}
