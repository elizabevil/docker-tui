package audit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileSinkWritesJSONL(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs-test")
	sink, err := NewFileSink(dir)
	if err != nil {
		t.Fatal(err)
	}
	rec := Record{
		Time:    time.Date(2025, 7, 25, 10, 0, 0, 0, time.UTC),
		TraceID: "trace-1",
		EventID: "event-1",
		Action:  "test.action",
		Result:  ResultSucceeded,
		Level:   LevelInfo,
		Message: "hello",
	}
	if err := sink.WriteAudit(context.Background(), rec); err != nil {
		t.Fatalf("WriteAudit: %v", err)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	got := files[0].Name()
	want := "audit-2025-07-25.jsonl"
	if got != want {
		t.Errorf("file name = %q, want %q", got, want)
	}
	// Verify contents
	body, err := os.ReadFile(filepath.Join(dir, got))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(body) == 0 {
		t.Error("file is empty")
	}
}

func TestNewFileSinkCreatesDirectory(t *testing.T) {
	// Nested directory that does not yet exist.
	dir := filepath.Join(t.TempDir(), "deeply", "nested", "logs")
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("precondition: dir must not exist; got err=%v", err)
	}
	sink, err := NewFileSink(dir)
	if err != nil {
		t.Fatalf("NewFileSink: %v", err)
	}
	if sink == nil {
		t.Fatal("expected non-nil sink")
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected directory, got file mode %v", info.Mode())
	}
}

func TestNewFileSinkAcceptsEmptyDir(t *testing.T) {
	// An empty dir disables file logging without erroring.
	sink, err := NewFileSink("")
	if err != nil {
		t.Fatalf("NewFileSink(\"\"): %v", err)
	}
	if sink == nil {
		t.Fatal("expected non-nil sink")
	}
	if err := sink.WriteAudit(context.Background(), Record{Action: "x"}); err == nil {
		t.Error("expected error when writing to empty-dir sink, got nil")
	}
}
