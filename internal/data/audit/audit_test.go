package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type memorySink struct{ records []Record }

func (s *memorySink) WriteAudit(_ context.Context, record Record) error {
	s.records = append(s.records, record)
	return nil
}

func TestServiceWritesTraceAndProjectsTerminalResult(t *testing.T) {
	sink := &memorySink{}
	service := NewService(sink)
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	sequence := 0
	service.newID = func() string { sequence++; return string(rune('a' + sequence - 1)) }

	trace := service.Begin("resource.container.stop", ContainerTarget{ID: "id", Name: "api"}, RuntimeContext{Type: "docker"}, UIContext{View: "containers"}, "Stopping api")
	service.Finish(trace, ResultSucceeded, "Stopped api", Details{})

	if len(sink.records) != 3 {
		t.Fatalf("records=%d, want requested/started/succeeded", len(sink.records))
	}
	for _, record := range sink.records {
		if record.TraceID != trace.ID || record.Target.Type != "container" {
			t.Fatalf("record lost trace or target: %#v", record)
		}
	}
	if got := service.CurrentNotification(); got == nil || got.Message != "Stopped api" {
		t.Fatalf("notification=%#v", got)
	}
	if got := service.RecentOperations(); len(got) != 3 {
		t.Fatalf("operation history=%d", len(got))
	}
}

func TestFileSinkWritesDatedJSONL(t *testing.T) {
	dir := t.TempDir()
	sink := NewFileSink(dir)
	record := Record{Time: time.Date(2026, 7, 20, 1, 2, 3, 0, time.UTC), TraceID: "trace", EventID: "event", Action: "resource.image.pull", Result: ResultStarted, Level: LevelInfo, Target: ImageTarget{ID: "nginx", Name: "nginx"}.ToDTO()}
	if err := sink.WriteAudit(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "audit-2026-07-20.jsonl")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		t.Fatal("audit log is empty")
	}
	var got Record
	if err := json.Unmarshal(scanner.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.TraceID != "trace" || got.Target.Name != "nginx" {
		t.Fatalf("record=%#v", got)
	}
}

func TestOperationProjectorIsBounded(t *testing.T) {
	projector := NewOperationLogProjector(2)
	for _, id := range []string{"one", "two", "three"} {
		projector.OnAuditRecord(Record{EventID: id})
	}
	recent := projector.Recent()
	if len(recent) != 2 || recent[0].EventID != "two" || recent[1].EventID != "three" {
		t.Fatalf("recent=%#v", recent)
	}
}
