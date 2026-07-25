package state

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
)

func TestConfirmStateLifecycle(t *testing.T) {
	trace := audit.Trace{ID: "trace", Action: "stop"}
	var s ConfirmState
	s.Open("stop", "container", "Stop container?", trace)
	if s.ConfirmAction != "stop" || s.ConfirmTarget != "container" || s.ConfirmAudit.ID != trace.ID {
		t.Fatalf("open confirm = %#v", s)
	}
	s.Close()
	if s.ConfirmAction != "" || s.ConfirmTarget != "" || s.ConfirmMessage != "" || s.ConfirmAudit.Valid() {
		t.Fatalf("close confirm = %#v", s)
	}
}
