package state

import "testing"

func TestViewportStateResizeAndToggleHeader(t *testing.T) {
	var s ViewportState
	s.Resize(-1, 24)
	if s.Width != 0 || s.Height != 24 {
		t.Fatalf("resize = %#v", s)
	}
	s.ToggleHeader()
	if !s.HeaderVisible {
		t.Fatal("header was not enabled")
	}
}
