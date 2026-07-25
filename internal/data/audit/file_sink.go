package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileSink struct {
	dir string
	now func() time.Time
	mu  sync.Mutex
}

func NewFileSink(dir string) (*FileSink, error) {
	if dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("create audit directory: %w", err)
		}
	}
	return &FileSink{dir: dir, now: time.Now}, nil
}

func (s *FileSink) WriteAudit(_ context.Context, record Record) error {
	if s == nil || s.dir == "" {
		return fmt.Errorf("audit log directory is empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return fmt.Errorf("create audit directory: %w", err)
	}
	date := record.Time
	if date.IsZero() {
		date = s.now()
	}
	path := filepath.Join(s.dir, "audit-"+date.Format("2006-01-02")+".jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer file.Close() //nolint:errcheck // append-only writer; close error is non-actionable.
	encoded, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode audit record: %w", err)
	}
	encoded = append(encoded, '\n')
	if _, err := file.Write(encoded); err != nil {
		return fmt.Errorf("write audit record: %w", err)
	}
	return nil
}
