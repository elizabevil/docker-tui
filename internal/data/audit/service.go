package audit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Service struct {
	sink          Sink
	notifications *NotificationProjector
	operations    *OperationLogProjector
	now           func() time.Time
	newID         func() string
	lastErrorMu   sync.RWMutex
	lastError     error
}

func NewService(sink Sink) *Service {
	return &Service{
		sink:          sink,
		notifications: &NotificationProjector{},
		operations:    NewOperationLogProjector(100),
		now:           time.Now,
		newID:         randomID,
	}
}

func (s *Service) Begin(action string, target Target, runtime RuntimeContext, ui UIContext, message string) Trace {
	if s == nil || target == nil || action == "" {
		return Trace{}
	}
	now := s.now().UTC()
	trace := Trace{ID: s.newID(), Action: action, StartedAt: now, Runtime: runtime, UI: ui, Target: target.ToDTO()}
	s.publish(Record{Time: now, TraceID: trace.ID, EventID: s.newID(), Action: action, Result: ResultRequested, Level: LevelInfo, Message: message, Runtime: runtime, UI: ui, Target: trace.Target})
	s.publish(Record{Time: s.now().UTC(), TraceID: trace.ID, EventID: s.newID(), Action: action, Result: ResultStarted, Level: LevelInfo, Message: message, Runtime: runtime, UI: ui, Target: trace.Target})
	return trace
}

func (s *Service) Finish(trace Trace, result Result, message string, details Details) {
	if s == nil || !trace.Valid() {
		return
	}
	if details.DurationMs == 0 {
		details.DurationMs = s.now().Sub(trace.StartedAt).Milliseconds()
	}
	level := LevelInfo
	if result == ResultFailed {
		level = LevelError
	} else if result == ResultCancelled {
		level = LevelWarn
	}
	s.publish(Record{Time: s.now().UTC(), TraceID: trace.ID, EventID: s.newID(), Action: trace.Action, Result: result, Level: level, Message: message, Runtime: trace.Runtime, UI: trace.UI, Target: trace.Target, Details: details})
}

func (s *Service) PublishUI(level Level, message string) {
	if s == nil || message == "" {
		return
	}
	s.notifications.OnUIMessage(UIMessage{Level: level, Message: message})
}

func (s *Service) CurrentNotification() *UIMessage { return s.notifications.Current() }
func (s *Service) CurrentOperation() *Record       { return s.operations.Current() }
func (s *Service) RecentOperations() []Record      { return s.operations.Recent() }

func (s *Service) LastError() error {
	if s == nil {
		return nil
	}
	s.lastErrorMu.RLock()
	defer s.lastErrorMu.RUnlock()
	return s.lastError
}

func (s *Service) publish(record Record) {
	s.operations.OnAuditRecord(record)
	s.notifications.OnAuditRecord(record)
	if s.sink != nil {
		if err := s.sink.WriteAudit(context.Background(), record); err != nil {
			s.lastErrorMu.Lock()
			s.lastError = err
			s.lastErrorMu.Unlock()
		}
	}
}

var fallbackID atomic.Uint64

func randomID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err == nil {
		return hex.EncodeToString(value[:])
	}
	return fmt.Sprintf("%x-%x", time.Now().UnixNano(), fallbackID.Add(1))
}
