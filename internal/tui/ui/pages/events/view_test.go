package events

import (
	"strings"
	"testing"

	dockerruntime "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func sample(ev *state.EventsPanelState, n int) {
	items := make([]dockerruntime.EventItem, 0, n)
	for i := range n {
		items = append(items, dockerruntime.EventItem{Event: dockerruntime.Event{
			ResourceType: "container",
			Action:       "start",
			ActorID:      "abc",
			Attributes:   map[string]string{"name": "web"},
			Time:         int64(1700000000 + i),
		}})
	}
	ev.AppendItems(items)
}

func tail(s string) string {
	if len(s) > 120 {
		return s[len(s)-120:]
	}
	return s
}

func TestRenderViewEmpty(t *testing.T) {
	got := RenderView(&state.EventsPanelState{}, 20, 120)
	if got == "" {
		t.Fatal("empty state rendered nothing")
	}
	if !strings.Contains(got, "Events") {
		t.Errorf("missing title in header: %q", got)
	}
}

func TestRenderViewShowsRowContent(t *testing.T) {
	ev := &state.EventsPanelState{}
	sample(ev, 3)
	got := RenderView(ev, 20, 120)
	if !strings.Contains(got, "container") || !strings.Contains(got, "start") {
		t.Errorf("row content missing: %q", got)
	}
}

func TestRenderViewOnlyVisibleRows(t *testing.T) {
	ev := &state.EventsPanelState{}
	sample(ev, 50)
	// 20-4=16 visible rows; footer page info should reflect 1-16/50.
	got := RenderView(ev, 20, 120)
	if !strings.Contains(got, "1-16/50") {
		t.Errorf("expected window 1-16/50, got tail: %q", tail(got))
	}
}

func TestRenderViewPausedStatus(t *testing.T) {
	ev := &state.EventsPanelState{}
	sample(ev, 2)
	ev.Pause()
	got := RenderView(ev, 20, 120)
	if !strings.Contains(got, "(paused)") {
		t.Errorf("expected paused title, got: %q", tail(got))
	}
}

func TestRenderViewClampOffset(t *testing.T) {
	ev := &state.EventsPanelState{}
	sample(ev, 3)
	ev.ViewOffset = 1000
	got := RenderView(ev, 20, 120) // must not panic
	if got == "" {
		t.Fatal("rendered nothing after clamping offset")
	}
}

func TestRenderViewFilteredOut(t *testing.T) {
	ev := &state.EventsPanelState{}
	sample(ev, 3)
	ev.SetFilter("network")
	if len(ev.FilteredEvents()) != 0 {
		t.Fatalf("network filter matched events, want 0")
	}
	got := RenderView(ev, 20, 120)
	if !strings.Contains(got, "filter: network") {
		t.Errorf("filter label missing from title: %q", tail(got))
	}
}

func TestEventRowFormat(t *testing.T) {
	row := eventRow(dockerruntime.Event{
		ResourceType: "image",
		Action:       "pull",
		ActorID:      "sha256:deadbeef",
		Attributes:   map[string]string{},
		Time:         int64(1700000000),
	})
	if len(row) != 4 {
		t.Fatalf("row length = %d, want 4", len(row))
	}
	if row[1] != "image" || row[2] != "pull" {
		t.Errorf("type/action mismatch: %#v", row)
	}
	if row[3] != "sha256:deadbeef" {
		t.Errorf("resource = %q, want actor id", row[3])
	}
	if !strings.HasPrefix(row[0], "2023-") {
		t.Errorf("time = %q", row[0])
	}
}
