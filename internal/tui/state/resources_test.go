package state

import "testing"

func TestNewResourceStateInitializesLists(t *testing.T) {
	s := NewResourceState()
	if s.Containers == nil || s.Images == nil || s.Volumes == nil || s.Networks == nil {
		t.Fatalf("resource lists not initialized: %#v", s)
	}
}
