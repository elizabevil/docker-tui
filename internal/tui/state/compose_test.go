package state

import "testing"

func TestComposeStateBoundsCursorsWithoutMutation(t *testing.T) {
	s := ComposeState{ComposeCursor: 9, ComposeServiceCursor: -2, ComposeContainerCursor: 5}
	if got := s.ProjectCursor(3); got != 2 {
		t.Fatalf("project cursor = %d, want 2", got)
	}
	if got := s.ServiceCursor(3); got != 0 {
		t.Fatalf("service cursor = %d, want 0", got)
	}
	if got := s.ContainerCursor(0); got != 0 {
		t.Fatalf("container cursor = %d, want 0", got)
	}
	if s.ComposeCursor != 9 || s.ComposeServiceCursor != -2 || s.ComposeContainerCursor != 5 {
		t.Fatalf("read-only cursor methods mutated state: %#v", s)
	}
}

func TestComposeStateContainerLifecycle(t *testing.T) {
	s := ComposeState{ComposeContainerCursor: 4}
	s.OpenContainers("api")
	if s.ComposeContainerViewID != "api" || s.ComposeContainerCursor != 0 {
		t.Fatalf("open containers = %#v", s)
	}
	s.ComposeContainerCursor = 2
	s.CloseContainers()
	if s.ComposeContainerViewID != "" || s.ComposeContainerCursor != 0 {
		t.Fatalf("close containers = %#v", s)
	}
}
