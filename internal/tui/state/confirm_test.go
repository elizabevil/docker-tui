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
	if len(s.Options) != 2 || s.Focus != 0 || s.Options[0].ID != "cancel" || s.Options[1].ID != "confirm" {
		t.Fatalf("open options = %#v", s.Options)
	}
	s.MoveFocus(1)
	if s.Focus != 1 {
		t.Fatalf("focus after tab = %d", s.Focus)
	}
	s.MoveFocus(1)
	if s.Focus != 0 {
		t.Fatalf("focus should wrap = %d", s.Focus)
	}
	s.Close()
	if s.ConfirmAction != "" || s.ConfirmTarget != "" || s.ConfirmMessage != "" || s.ConfirmAudit.Valid() || len(s.Options) != 0 || s.Focus != 0 {
		t.Fatalf("close confirm = %#v", s)
	}
}

func TestConfirmStateMoveFocusSkipsDisabledOptions(t *testing.T) {
	s := ConfirmState{Options: []ChoiceOption{
		{ID: "one", Label: "One"},
		{ID: "two", Label: "Two", Disabled: true},
		{ID: "three", Label: "Three"},
	}}
	s.MoveFocus(1)
	if s.Focus != 2 {
		t.Fatalf("focus skipped disabled option = %d", s.Focus)
	}
	s.MoveFocus(-1)
	if s.Focus != 0 {
		t.Fatalf("reverse focus skipped disabled option = %d", s.Focus)
	}
}
