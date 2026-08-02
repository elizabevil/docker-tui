package state

import (
	"strings"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// MaxEvents caps the retained runtime event payload. The oldest events are
// dropped first (ring-buffer semantics).
const MaxEvents = 1000

// EventsPanelState is the UI-facing state backing the Events panel (BR-035).
//
// It is independent of EventState (which owns the background subscription
// lifecycle in events.go). The two live on distinct AppModel fields:
//
//	Events     EventState        // subscription lifecycle (update_events.go)
//	EventPanel EventsPanelState  // BR-035 panel buffer (this file)
//
// A subscriber drains runtime events into this panel state via AppendItems;
// the panel renders Time/Type/Action/Resource.
//
// Pause semantics: while Paused, newly delivered events move into the Pending
// buffer instead of the visible list. Resume() flushes Pending into Events so a
// paused panel can catch up without losing entries. The retention cap applies
// to the combined live+pending buffer, so a paused panel never holds more than
// MaxEvents events in total.
type EventsPanelState struct {
	Events       []runtimeapi.Event
	Pending      []runtimeapi.Event
	Paused       bool
	Filter       string
	Filtering    bool
	Cursor       int
	ViewOffset   int
	PreviousMode AppMode
}

// Open resets panel navigation and transient state (BR-035: F3 entry resets
// the cursor and paused state). The retained event buffer is left intact.
func (s *EventsPanelState) Open(previous AppMode) {
	if s.Paused {
		s.Resume()
	}
	s.Paused = false
	s.Filter = ""
	s.Filtering = false
	s.Cursor = 0
	s.ViewOffset = 0
	s.PreviousMode = previous
}

// Close resets the whole panel state.
func (s *EventsPanelState) Close() {
	*s = EventsPanelState{}
}

// Pause stops live accumulation; new events buffer into Pending.
func (s *EventsPanelState) Pause() { s.Paused = true }

// Resume flushes the pending buffer into Events and resumes live delivery.
func (s *EventsPanelState) Resume() {
	if !s.Paused {
		return
	}
	s.Paused = false
	s.Events = append(s.Events, s.Pending...)
	s.Pending = s.Pending[:0]
	if excess := len(s.Events) - MaxEvents; excess > 0 {
		s.Events = append([]runtimeapi.Event(nil), s.Events[excess:]...)
	}
}

// TogglePause flips the paused state.
func (s *EventsPanelState) TogglePause() {
	if s.Paused {
		s.Resume()
	} else {
		s.Pause()
	}
}

// AppendItems ingests runtime event stream items. Error / terminal items are
// skipped; delivered events respect the live-or-paused routing rule.
func (s *EventsPanelState) AppendItems(items []runtimeapi.EventItem) {
	live := !s.Paused
	for _, it := range items {
		if it.Error != nil {
			continue
		}
		if live {
			s.append(&s.Events, it.Event)
		} else {
			s.append(&s.Pending, it.Event)
		}
	}
}

// append inserts events into a capped slice. The retention cap applies to the
// combined live+pending buffer: when total retained events exceed MaxEvents,
// the oldest are dropped — from the live list first, then from the paused
// buffer — so a paused panel never holds more than MaxEvents events in total.
func (s *EventsPanelState) append(dst *[]runtimeapi.Event, events ...runtimeapi.Event) {
	*dst = append(*dst, events...)
	excess := len(s.Events) + len(s.Pending) - MaxEvents
	if excess <= 0 {
		return
	}
	if drop := min(excess, len(s.Events)); drop > 0 {
		s.Events = append([]runtimeapi.Event(nil), s.Events[drop:]...)
		excess -= drop
	}
	if excess > 0 {
		s.Pending = append([]runtimeapi.Event(nil), s.Pending[excess:]...)
	}
}

// Clear empties both the live list and the paused buffer and resets the
// filter back to the default (open-view) state.
func (s *EventsPanelState) Clear() {
	s.Events = s.Events[:0]
	s.Pending = s.Pending[:0]
	s.Filter = ""
	s.Cursor = 0
	s.ViewOffset = 0
}

// SetFilter applies a case-insensitive substring filter across type/action and
// resets the cursor.
func (s *EventsPanelState) SetFilter(f string) {
	s.Filter = strings.TrimSpace(f)
	s.Cursor = 0
	s.ViewOffset = 0
}

// FilteredEvents returns the live events matching Filter (all when empty).
func (s *EventsPanelState) FilteredEvents() []runtimeapi.Event {
	if s.Filter == "" {
		return s.Events
	}
	q := strings.ToLower(s.Filter)
	out := make([]runtimeapi.Event, 0, len(s.Events))
	for _, ev := range s.Events {
		if EventMatches(ev, q) {
			out = append(out, ev)
		}
	}
	return out
}

// EventMatches reports whether an event matches a query across its resource
// type and action. Matching is case-insensitive.
func EventMatches(ev runtimeapi.Event, query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	return strings.Contains(strings.ToLower(ev.ResourceType), q) ||
		strings.Contains(strings.ToLower(ev.Action), q)
}

// MoveCursor moves the selection within total (already-filtered) events and
// keeps the cursor visible.
func (s *EventsPanelState) MoveCursor(delta, total, visible int) {
	if total <= 0 {
		s.Cursor = 0
		return
	}
	s.Cursor = min(max(s.Cursor+delta, 0), total-1)
	s.EnsureVisible(visible)
}

// EnsureVisible scrolls ViewOffset so the cursor lies inside the visible window.
func (s *EventsPanelState) EnsureVisible(visible int) {
	if visible <= 0 {
		s.ViewOffset = 0
		return
	}
	if s.Cursor < s.ViewOffset {
		s.ViewOffset = s.Cursor
	} else if s.Cursor >= s.ViewOffset+visible {
		s.ViewOffset = s.Cursor - visible + 1
	}
	if s.ViewOffset < 0 {
		s.ViewOffset = 0
	}
}

// ClampOffset clamps ViewOffset and Cursor against the filtered result total
// so the state never retains a stale offset after the result set shrinks
// (filter change, list truncation, or buffer eviction). The renderer calls
// this before drawing so subsequent keyboard navigation stays consistent with
// the actual viewport.
func (s *EventsPanelState) ClampOffset(total, visible int) {
	if total <= 0 {
		s.Cursor = 0
		s.ViewOffset = 0
		return
	}
	maxOffset := max(0, total-visible)
	if s.ViewOffset > maxOffset {
		s.ViewOffset = maxOffset
	}
	if s.Cursor > total-1 {
		s.Cursor = total - 1
	}
	if s.ViewOffset < 0 {
		s.ViewOffset = 0
	}
}
