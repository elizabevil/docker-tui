package state

import (
	"context"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

const eventReconnectMaxDelay = 30 * time.Second

type EventState struct {
	Generation     uint64
	Context        context.Context
	Cancel         context.CancelFunc
	Failures       int
	Degraded       bool
	FlushScheduled bool
	FlushToken     uint64
	Dirty          map[string]bool
}

func (s *EventState) Begin() (context.Context, uint64) {
	s.Stop()
	s.Generation++
	ctx := s.renewContext()
	s.Failures = 0
	s.Degraded = false
	s.FlushScheduled = false
	s.Dirty = make(map[string]bool)
	return ctx, s.Generation
}

// RenewContext cancels the active subscription while preserving its connection generation.
func (s *EventState) RenewContext() context.Context {
	if s.Cancel != nil {
		s.Cancel()
	}
	return s.renewContext()
}

func (s *EventState) renewContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	s.Context = ctx
	s.Cancel = cancel
	return ctx
}

func (s *EventState) Stop() {
	if s.Cancel != nil {
		s.Cancel()
	}
	s.Cancel = nil
	s.Context = nil
	s.FlushScheduled = false
	s.Dirty = nil
}

func (s *EventState) Current(generation uint64) bool {
	return generation != 0 && generation == s.Generation && s.Cancel != nil
}

func (s *EventState) MarkDirty(resourceType string) (uint64, bool) {
	if s.Dirty == nil {
		s.Dirty = make(map[string]bool)
	}
	s.Dirty[resourceType] = true
	if s.FlushScheduled {
		return s.FlushToken, false
	}
	s.FlushScheduled = true
	s.FlushToken++
	return s.FlushToken, true
}

func (s *EventState) TakeDirty(token uint64) map[string]bool {
	if !s.FlushScheduled || token != s.FlushToken {
		return nil
	}
	dirty := s.Dirty
	s.Dirty = make(map[string]bool)
	s.FlushScheduled = false
	return dirty
}

func (s *EventState) RecordFailure() time.Duration {
	s.Failures++
	s.Degraded = true
	delay := time.Second << min(s.Failures-1, 5)
	return min(delay, eventReconnectMaxDelay)
}

func (s *EventState) RecordReady() {
	s.Failures = 0
	s.Degraded = false
}

type EventStreamReady struct {
	Generation uint64
	Items      <-chan runtimeapi.EventItem
}

type EventStreamFailed struct {
	Generation uint64
	Error      error
}

type RuntimeEventReceived struct {
	Generation uint64
	Item       runtimeapi.EventItem
	Items      <-chan runtimeapi.EventItem
}

type EventStreamClosed struct{ Generation uint64 }
type EventFlush struct {
	Generation uint64
	Token      uint64
}
type EventReconnect struct{ Generation uint64 }
type EventFallbackTick struct{ Generation uint64 }
