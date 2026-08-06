package state

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
)

func TestSelectionStateMarksAndPendingPull(t *testing.T) {
	var s SelectionState
	s.Toggle(PanelContainers, "one")
	if !s.PanelMarks[PanelContainers]["one"] {
		t.Fatal("toggle did not mark item")
	}
	s.Toggle(PanelContainers, "one")
	if s.PanelMarks[PanelContainers]["one"] {
		t.Fatal("second toggle did not clear item")
	}
	s.Toggle(PanelContainers, "two")
	s.Toggle(PanelImages, "image")
	s.ClearPanelMarks(PanelContainers)
	if s.MarkedCount(PanelContainers) != 0 || s.MarkedCount(PanelImages) != 1 {
		t.Fatalf("panel marks not isolated or cleared: %#v", s.PanelMarks)
	}
	s.ClearMarks()
	if len(s.PanelMarks) != 0 {
		t.Fatalf("marks not cleared: %#v", s.PanelMarks)
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
