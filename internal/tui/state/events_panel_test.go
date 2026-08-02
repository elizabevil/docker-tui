package state

import (
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func sampleEvents(n int) []runtimeapi.EventItem {
	items := make([]runtimeapi.EventItem, n)
	for i := range items {
		items[i] = runtimeapi.EventItem{Event: runtimeapi.Event{
			ResourceType: "container",
			Action:       "start",
			ActorID:      "abc123",
			Attributes:   map[string]string{"name": "web"},
			Time:         int64(1700000000 + i),
		}}
	}
	return items
}

func TestAppendItemsRingCap(t *testing.T) {
	var s EventsPanelState
	s.AppendItems(sampleEvents(MaxEvents + 25))
	if len(s.Events) != MaxEvents {
		t.Fatalf("len = %d, want %d", len(s.Events), MaxEvents)
	}
	if first := s.Events[0].Time; first != int64(1700000000+25) {
		t.Errorf("oldest event not dropped: Time=%d, want %d", first, int64(1700000025))
	}
	if last := s.Events[len(s.Events)-1].Time; last != int64(1700000000+MaxEvents+24) {
		t.Errorf("newest event missing: Time=%d", last)
	}
}

func TestAppendItemsSkipsErrorItems(t *testing.T) {
	var s EventsPanelState
	s.AppendItems([]runtimeapi.EventItem{{Error: errBrokenStream(1)}})
	if len(s.Events) != 0 {
		t.Fatalf("error item leaked into buffer: %d", len(s.Events))
	}
}

func TestPauseBuffersAndResumeFlushes(t *testing.T) {
	var s EventsPanelState
	s.AppendItems(sampleEvents(3))
	s.Pause()
	s.AppendItems(sampleEvents(2)) // 2 buffered while paused
	if len(s.Events) != 3 {
		t.Fatalf("live list grew while paused: %d", len(s.Events))
	}
	if len(s.Pending) != 2 {
		t.Fatalf("pending = %d, want 2", len(s.Pending))
	}
	s.Resume()
	if s.Paused {
		t.Fatal("Resume did not clear paused")
	}
	if len(s.Events) != 5 {
		t.Fatalf("after resume events = %d, want 5", len(s.Events))
	}
	if len(s.Pending) != 0 {
		t.Fatalf("pending not drained: %d", len(s.Pending))
	}
}

func TestTogglePause(t *testing.T) {
	var s EventsPanelState
	s.TogglePause()
	if !s.Paused {
		t.Fatal("first toggle should pause")
	}
	s.TogglePause()
	if s.Paused {
		t.Fatal("second toggle should resume")
	}
}

func TestPendingBufferCapped(t *testing.T) {
	var s EventsPanelState
	s.Pause()
	s.AppendItems(sampleEvents(MaxEvents + 10))
	if len(s.Pending) != MaxEvents {
		t.Fatalf("pending cap not enforced: %d", len(s.Pending))
	}
}

func TestCombinedLiveAndPendingCap(t *testing.T) {
	var s EventsPanelState
	s.AppendItems(sampleEvents(600))
	s.Pause()
	s.AppendItems(sampleEvents(600)) // 600 live + 600 pending > MaxEvents
	if total := len(s.Events) + len(s.Pending); total != MaxEvents {
		t.Fatalf("combined live+pending = %d, want %d", total, MaxEvents)
	}
	// Oldest live events must be dropped before any pending events.
	if len(s.Pending) != 600 {
		t.Fatalf("pending = %d, want 600 (pending must survive live eviction)", len(s.Pending))
	}
	if len(s.Events) != 400 {
		t.Fatalf("live = %d, want 400 (oldest 200 dropped)", len(s.Events))
	}
	if first := s.Events[0].Time; first != int64(1700000000+200) {
		t.Errorf("oldest retained live event Time=%d, want %d", first, int64(1700000200))
	}
	s.Resume()
	if len(s.Events) != MaxEvents || len(s.Pending) != 0 {
		t.Fatalf("resume retained events=%d pending=%d, want %d/0", len(s.Events), len(s.Pending), MaxEvents)
	}
}

func TestClampOffset(t *testing.T) {
	var s EventsPanelState
	s.ViewOffset = 900
	s.Cursor = 950
	s.ClampOffset(100, 10)
	if s.ViewOffset != 90 {
		t.Fatalf("ViewOffset = %d, want 90 (total-visible)", s.ViewOffset)
	}
	if s.Cursor != 99 {
		t.Fatalf("Cursor = %d, want 99 (total-1)", s.Cursor)
	}
	s.ViewOffset = -5
	s.Cursor = -1
	s.ClampOffset(10, 10)
	if s.ViewOffset != 0 {
		t.Fatalf("negative ViewOffset not clamped: %d", s.ViewOffset)
	}
	s.ClampOffset(0, 10)
	if s.Cursor != 0 || s.ViewOffset != 0 {
		t.Fatalf("zero-total not reset: cursor=%d offset=%d", s.Cursor, s.ViewOffset)
	}
}

func TestClearEmptiesEventsAndPending(t *testing.T) {
	var s EventsPanelState
	s.AppendItems(sampleEvents(4))
	s.Pause()
	s.AppendItems(sampleEvents(2))
	s.SetFilter("container")
	s.Cursor = 3
	s.Clear()
	if len(s.Events) != 0 || len(s.Pending) != 0 {
		t.Fatalf("clear left events=%d pending=%d", len(s.Events), len(s.Pending))
	}
	if s.Cursor != 0 || s.Filter != "" {
		t.Fatalf("clear left cursor=%d filter=%q", s.Cursor, s.Filter)
	}
}

func TestSetFilterResetCursor(t *testing.T) {
	var s EventsPanelState
	s.AppendItems(sampleEvents(5))
	s.Cursor = 4
	s.ViewOffset = 3
	s.SetFilter("IMAGE")
	filtered := s.FilteredEvents()
	if len(filtered) != 0 {
		t.Fatalf("image filter matched %d, want 0", len(filtered))
	}
	s.SetFilter("CONTAINER")
	if len(s.FilteredEvents()) != 5 {
		t.Fatalf("container filter matched != all")
	}
	if s.Cursor != 0 || s.ViewOffset != 0 {
		t.Fatalf("filter did not reset cursor/offset")
	}
}

func TestEventMatches(t *testing.T) {
	ev := runtimeapi.Event{ResourceType: "network", Action: "disconnect"}
	if !EventMatches(ev, "network") {
		t.Error("type filter missed")
	}
	if !EventMatches(ev, "DISCONNECT") {
		t.Error("case-insensitive action filter missed")
	}
	if EventMatches(ev, "volume") {
		t.Error("non-matching filter matched")
	}
}

func TestMoveCursorAndEnsureVisible(t *testing.T) {
	var s EventsPanelState
	s.MoveCursor(+1, 10, 5)
	if s.Cursor != 1 || s.ViewOffset != 0 {
		t.Fatalf("cursor=%d offset=%d", s.Cursor, s.ViewOffset)
	}
	s.MoveCursor(+10, 10, 5)
	if s.Cursor != 9 || s.ViewOffset != 5 {
		t.Fatalf("cursor=%d offset=%d", s.Cursor, s.ViewOffset)
	}
	s.MoveCursor(-10, 10, 5)
	if s.Cursor != 0 || s.ViewOffset != 0 {
		t.Fatalf("cursor=%d offset=%d", s.Cursor, s.ViewOffset)
	}
	s.MoveCursor(0, 0, 5)
	if s.Cursor != 0 {
		t.Fatal("zero-total did not reset cursor")
	}
}

type brokenErr int

func (b brokenErr) Error() string { return "stream error" }

func errBrokenStream(i int) error { return brokenErr(i) }
