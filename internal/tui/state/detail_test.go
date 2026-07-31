package state

import (
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

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

func TestDetailStateWritesImageRawJSON(t *testing.T) {
	var detail DetailState
	detail.OpenImage("img", "Image", &runtimeapi.ImageDetail{ID: "img", Name: "demo"})
	if detail.DetailResourceType != ResourceImage {
		t.Fatalf("open image resource type = %q", detail.DetailResourceType)
	}
	if len(detail.DetailRawJSON) == 0 {
		t.Fatal("open image raw json missing")
	}

	if !detail.ApplyImage("img", &runtimeapi.ImageDetail{ID: "img", Name: "demo2"}) {
		t.Fatal("apply image rejected current result")
	}
	if len(detail.DetailRawJSON) == 0 || detail.ImageDetailData == nil || detail.ImageDetailData.Name != "demo2" {
		t.Fatalf("apply image raw json not updated: %#v", detail)
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
