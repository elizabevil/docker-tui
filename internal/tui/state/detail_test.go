package state

import "testing"

func TestDetailStateLifecycleAndSources(t *testing.T) {
	var detail DetailState
	detail.Open("Container", "content")
	detail.Scroll(4)
	detail.CycleSource()
	if detail.DetailSourceType != DetailSourceYAML || detail.DetailOffset != 0 {
		t.Fatalf("yaml detail = %#v", detail)
	}
	detail.CycleSource()
	detail.CycleSource()
	if detail.DetailSourceType != DetailSourceSection {
		t.Fatalf("cycled source = %q", detail.DetailSourceType)
	}
	detail.Close()
	if detail.DetailTitle != "" || detail.ImageDetailContent != "" {
		t.Fatalf("closed detail = %#v", detail)
	}
}

func TestDetailStateRejectsStaleImageResult(t *testing.T) {
	var detail DetailState
	detail.OpenImage("current", "Image", nil)
	if detail.ApplyImage("stale", nil) {
		t.Fatal("stale image result was applied")
	}
	detail.DetailOffset = 5
	if !detail.ApplyImage("current", nil) || detail.DetailOffset != 0 {
		t.Fatalf("current image result = %#v", detail)
	}
}

func TestClampVisibleOffsetRemovesOverscroll(t *testing.T) {
	detail := DetailState{DetailOffset: 100}
	if got := detail.ClampVisibleOffset(30, 10); got != 20 {
		t.Fatalf("ClampVisibleOffset() = %d, want 20", got)
	}
	if detail.DetailOffset != 20 {
		t.Fatalf("DetailOffset = %d, want 20", detail.DetailOffset)
	}

	detail.Scroll(-1)
	if detail.DetailOffset != 19 {
		t.Fatalf("up scroll after clamp = %d, want 19", detail.DetailOffset)
	}
}
