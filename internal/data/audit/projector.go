package audit

import "sync"

type NotificationProjector struct {
	mu     sync.RWMutex
	active *UIMessage
}

func (p *NotificationProjector) OnAuditRecord(record Record) {
	if record.Result != ResultSucceeded && record.Result != ResultFailed && record.Result != ResultCancelled {
		return
	}
	p.OnUIMessage(UIMessage{Level: record.Level, Message: record.Message})
}

func (p *NotificationProjector) OnUIMessage(message UIMessage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	copy := message
	p.active = &copy
}

func (p *NotificationProjector) Current() *UIMessage {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.active == nil {
		return nil
	}
	copy := *p.active
	return &copy
}

type OperationLogProjector struct {
	mu      sync.RWMutex
	history []Record
	limit   int
}

func NewOperationLogProjector(limit int) *OperationLogProjector {
	if limit <= 0 {
		limit = 100
	}
	return &OperationLogProjector{limit: limit}
}

func (p *OperationLogProjector) OnAuditRecord(record Record) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.history = append(p.history, record)
	if overflow := len(p.history) - p.limit; overflow > 0 {
		copy(p.history, p.history[overflow:])
		p.history = p.history[:p.limit]
	}
}

func (p *OperationLogProjector) Recent() []Record {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]Record(nil), p.history...)
}

func (p *OperationLogProjector) Current() *Record {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.history) == 0 {
		return nil
	}
	copy := p.history[len(p.history)-1]
	return &copy
}
