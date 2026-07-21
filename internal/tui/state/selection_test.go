package state

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
)

func TestSelectionStateMarksAndPendingPull(t *testing.T) {
	var s SelectionState
	s.Toggle("one")
	if !s.MarkedIDs["one"] {
		t.Fatal("toggle did not mark item")
	}
	s.Toggle("one")
	if s.MarkedIDs["one"] {
		t.Fatal("second toggle did not clear item")
	}
	s.Toggle("two")
	s.ClearMarks()
	if len(s.MarkedIDs) != 0 {
		t.Fatalf("marks not cleared: %#v", s.MarkedIDs)
	}

	trace := audit.Trace{ID: "trace", Action: "pull"}
	s.QueueImagePull("alpine:latest", trace)
	ref, gotTrace := s.TakeImagePull()
	if ref != "alpine:latest" || gotTrace.ID != trace.ID {
		t.Fatalf("take pull = %q, %#v", ref, gotTrace)
	}
	if s.PendingImagePull != "" || s.PendingImagePullAudit.Valid() {
		t.Fatalf("pending pull not cleared: %#v", s)
	}
}
