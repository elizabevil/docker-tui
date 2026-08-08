package keyboard

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// fakeEngine is a minimal runtimeapi.Engine stub for log tests. Only
// the ContainerService.Logs path is implemented because the rest of
// FetchComposeLogsBatch never touches the engine.
type fakeEngine struct {
	logs map[string]string // container id → log body
	errs map[string]error  // container id → fetch error
}

func (f *fakeEngine) Identity() runtimeapi.Identity {
	return runtimeapi.Identity{Type: "test", Name: "fake"}
}
func (f *fakeEngine) Capabilities() runtimeapi.CapabilitySet          { return runtimeapi.CapabilitySet{} }
func (f *fakeEngine) PingContext(context.Context) error               { return nil }
func (f *fakeEngine) Volumes() runtimeapi.VolumeService               { return nil }
func (f *fakeEngine) Networks() runtimeapi.NetworkService             { return nil }
func (f *fakeEngine) Images() runtimeapi.ImageService                 { return nil }
func (f *fakeEngine) ImageTransfers() runtimeapi.ImageTransferService { return nil }
func (f *fakeEngine) Actions() runtimeapi.ResourceActionService       { return nil }
func (f *fakeEngine) Exec() runtimeapi.ExecService                    { return nil }
func (f *fakeEngine) Events() runtimeapi.EventService                 { return nil }
func (f *fakeEngine) Compose() runtimeapi.ComposeService              { return nil }
func (f *fakeEngine) Pods() runtimeapi.PodService                     { return nil }
func (f *fakeEngine) Close() error                                    { return nil }
func (f *fakeEngine) Containers() runtimeapi.ContainerService {
	return &fakeContainerService{engine: f}
}

type fakeContainerService struct{ engine *fakeEngine }

func (s *fakeContainerService) List(context.Context, runtimeapi.ContainerListOptions) ([]runtimeapi.ContainerSummary, error) {
	return nil, nil
}
func (s *fakeContainerService) Inspect(context.Context, string) (*runtimeapi.ContainerDetail, error) {
	return nil, nil
}
func (s *fakeContainerService) Top(context.Context, string) (runtimeapi.ContainerProcesses, error) {
	return runtimeapi.ContainerProcesses{}, nil
}
func (s *fakeContainerService) Stats(context.Context, string) (runtimeapi.ContainerStats, error) {
	return runtimeapi.ContainerStats{}, nil
}
func (s *fakeContainerService) Logs(_ context.Context, id string, _ runtimeapi.ContainerLogOptions) (io.ReadCloser, error) {
	if err, ok := s.engine.errs[id]; ok {
		return nil, err
	}
	return io.NopCloser(strings.NewReader(s.engine.logs[id])), nil
}

func (s *fakeContainerService) Update(context.Context, string, runtimeapi.ContainerUpdateOptions) (runtimeapi.ContainerUpdateResult, error) {
	return runtimeapi.ContainerUpdateResult{}, nil
}
func (s *fakeContainerService) Diff(context.Context, string) ([]runtimeapi.ContainerDiffChange, error) {
	return nil, nil
}
func (s *fakeContainerService) Export(context.Context, string) (io.ReadCloser, error) {
	return nil, nil
}
func (s *fakeContainerService) Commit(context.Context, string, runtimeapi.ContainerCommitOptions) (runtimeapi.ContainerCommitResult, error) {
	return runtimeapi.ContainerCommitResult{}, nil
}
func (s *fakeContainerService) Wait(context.Context, string, string) (runtimeapi.ContainerWaitResult, error) {
	return runtimeapi.ContainerWaitResult{}, nil
}
func (s *fakeContainerService) CopyFromContainer(context.Context, string, string) (io.ReadCloser, error) {
	return nil, nil
}

// errBoom is a sentinel error used by the partial-failure test.
var errBoom = errors.New("boom")

func TestFetchComposeLogsBatchAggregatesMultipleSources(t *testing.T) {
	eng := &fakeEngine{
		logs: map[string]string{
			"web1": "GET /api 200\nGET /healthz 200",
			"db1":  "INSERT 12 rows\nSELECT 1",
		},
	}
	containers := []runtimeapi.ContainerSummary{
		{ID: "web1", Name: "demo_web_1", ComposeProject: "demo", ComposeService: "web"},
		{ID: "db1", Name: "demo_db_1", ComposeProject: "demo", ComposeService: "db"},
	}
	cmd := FetchComposeLogsBatch(eng, composeLogVirtualID("demo", ""), containers, runtimeapi.ComposeLogOptions{
		Since: "10s",
		Tail:  "200",
	})
	if cmd == nil {
		t.Fatal("FetchComposeLogsBatch returned nil cmd")
	}
	msg := cmd()
	batch, ok := msg.(state.LogBatchReceived)
	if !ok {
		t.Fatalf("expected LogBatchReceived, got %T", msg)
	}
	if batch.ContainerID != composeLogVirtualID("demo", "") {
		t.Errorf("containerID = %q", batch.ContainerID)
	}
	if len(batch.Lines) != 4 {
		t.Fatalf("expected 4 lines (2 per service), got %d: %#v", len(batch.Lines), batch.Lines)
	}
	wantCounts := map[string]int{"[db]": 2, "[web]": 2}
	counts := map[string]int{}
	for _, line := range batch.Lines {
		for prefix := range wantCounts {
			if strings.HasPrefix(line, prefix+" ") {
				counts[prefix]++
				break
			}
		}
	}
	for prefix, want := range wantCounts {
		if counts[prefix] != want {
			t.Errorf("prefix %s count = %d, want %d (lines=%v)", prefix, counts[prefix], want, batch.Lines)
		}
	}
}

func TestFetchComposeLogsBatchAssignsReplicaIndex(t *testing.T) {
	eng := &fakeEngine{
		logs: map[string]string{
			"web1": "from web1",
			"web2": "from web2",
		},
	}
	containers := []runtimeapi.ContainerSummary{
		{ID: "web1", Name: "demo_web_1", ComposeProject: "demo", ComposeService: "web"},
		{ID: "web2", Name: "demo_web_2", ComposeProject: "demo", ComposeService: "web"},
	}
	cmd := FetchComposeLogsBatch(eng, composeLogVirtualID("demo", ""), containers, runtimeapi.ComposeLogOptions{
		Since: "10s",
		Tail:  "200",
	})
	msg := cmd()
	batch := msg.(state.LogBatchReceived)
	if len(batch.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(batch.Lines))
	}
	want := []string{"[web.1]", "[web.2]"}
	for i, line := range batch.Lines {
		if !strings.HasPrefix(line, want[i]+" ") {
			t.Errorf("line %d prefix = %q, want %q", i, line, want[i])
		}
	}
}

func TestFetchComposeLogsBatchPartialFailure(t *testing.T) {
	eng := &fakeEngine{
		logs: map[string]string{"web1": "ok"},
		errs: map[string]error{"db1": errBoom},
	}
	containers := []runtimeapi.ContainerSummary{
		{ID: "web1", ComposeProject: "demo", ComposeService: "web"},
		{ID: "db1", ComposeProject: "demo", ComposeService: "db"},
	}
	cmd := FetchComposeLogsBatch(eng, composeLogVirtualID("demo", ""), containers, runtimeapi.ComposeLogOptions{
		Since: "10s",
		Tail:  "200",
	})
	msg := cmd()
	batch, ok := msg.(state.ComposeLogBatchReceived)
	if !ok {
		t.Fatalf("expected ComposeLogBatchReceived (partial-failure carrier), got %T", msg)
	}
	// R08-06 §F6: failed containers get a placeholder line "(service.replica:
	// connection lost)" inserted into the merged stream so the user
	// still sees which container is unavailable. The db1 placeholder
	// comes before web1 because both have zero timestamps and the sort
	// is stable over the original iteration order (db1 first).
	if len(batch.Lines) != 2 {
		t.Fatalf("lines = %#v, want 2 lines (web1 success + db1 placeholder)", batch.Lines)
	}
	if batch.Lines[0] != "[db] (db.1: connection lost)" {
		t.Errorf("lines[0] = %q, want db1 placeholder", batch.Lines[0])
	}
	if batch.Lines[1] != "[web] ok" {
		t.Errorf("lines[1] = %q, want [web] ok", batch.Lines[1])
	}
	if len(batch.StreamErrs) != 1 || batch.StreamErrs[0].ContainerID != "db1" {
		t.Errorf("streamErrs = %#v", batch.StreamErrs)
	}
}

func TestComposeLogVirtualIDScopesByService(t *testing.T) {
	if got := composeLogVirtualID("demo", ""); got != "compose:demo" {
		t.Errorf("project virtualID = %q", got)
	}
	if got := composeLogVirtualID("demo", "web"); got != "compose:demo:web" {
		t.Errorf("service virtualID = %q", got)
	}
}


