package keyboard

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/mockengine"
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
	cmd := FetchComposeLogsBatch(eng, composeLogVirtualID("demo", ""), containers, "10s", "200", false)
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
	cmd := FetchComposeLogsBatch(eng, composeLogVirtualID("demo", ""), containers, "10s", "200", false)
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
	cmd := FetchComposeLogsBatch(eng, composeLogVirtualID("demo", ""), containers, "10s", "200", false)
	msg := cmd()
	batch, ok := msg.(state.ComposeLogBatchReceived)
	if !ok {
		t.Fatalf("expected ComposeLogBatchReceived (partial-failure carrier), got %T", msg)
	}
	if len(batch.Lines) != 1 || batch.Lines[0] != "[web] ok" {
		t.Errorf("lines = %#v", batch.Lines)
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

// TestComposePull verifies the R08-05 compose pull loop: image
// references are deduplicated across services, each distinct image is
// pulled once, and failures aggregate without blocking the rest.
func TestComposePull(t *testing.T) {
	newPull := func() (*state.AppModel, *mockengine.Engine) {
		eng := newMockEngine()
		m := newAppModelWithEngine(eng)
		m.Navigation.ActivePanel = state.PanelCompose
		m.Compose.ComposeDetailProject = "demo"
		m.Resources.Containers.Items = []runtimeapi.ContainerSummary{
			{ID: "c1", Name: "web", Image: "demo_web:latest", ComposeProject: "demo", ComposeService: "web"},
			{ID: "c2", Name: "db", Image: "postgres:16", ComposeProject: "demo", ComposeService: "db",
				Labels: map[string]string{"com.docker.compose.image": "postgres:16"}},
		}
		return m, eng
	}
	pulledRefs := func(t *testing.T, eng *mockengine.Engine) []string {
		t.Helper()
		var refs []string
		for _, call := range eng.Calls() {
			if call.Type == runtimeapi.ResourceImage {
				refs = append(refs, call.ID)
			}
		}
		return refs
	}

	t.Run("dedupe", func(t *testing.T) {
		m, eng := newPull()
		// Three containers resolve to two distinct refs: c2 carries the
		// label, c3 resolves via the Image fallback — pulled once.
		m.Resources.Containers.Items = append(m.Resources.Containers.Items,
			runtimeapi.ContainerSummary{ID: "c3", Name: "worker", Image: "postgres:16", ComposeProject: "demo", ComposeService: "worker"})
		_, cmd := doComposePull(m)
		msg, ok := cmd().(state.ComposeServicePullCompleted)
		if !ok {
			t.Fatalf("cmd returned %T, want ComposeServicePullCompleted", msg)
		}
		if msg.Error != nil || msg.Success != 2 {
			t.Errorf("summary = success=%d err=%v, want 2 nil", msg.Success, msg.Error)
		}
		got := pulledRefs(t, eng)
		if len(got) != 2 || got[0] != "demo_web:latest" || got[1] != "postgres:16" {
			t.Errorf("pulled refs = %v, want [demo_web:latest postgres:16]", got)
		}
	})

	t.Run("success", func(t *testing.T) {
		m, eng := newPull()
		_, cmd := doComposePull(m)
		msg, ok := cmd().(state.ComposeServicePullCompleted)
		if !ok {
			t.Fatalf("cmd returned %T, want ComposeServicePullCompleted", msg)
		}
		if msg.Error != nil || msg.Success != 2 {
			t.Errorf("summary = success=%d err=%v, want 2 nil", msg.Success, msg.Error)
		}
		got := pulledRefs(t, eng)
		if len(got) != 2 || got[0] != "demo_web:latest" || got[1] != "postgres:16" {
			t.Errorf("pulled refs = %v, want [demo_web:latest postgres:16]", got)
		}
	})

	t.Run("partial failure", func(t *testing.T) {
		m, eng := newPull()
		eng.Fail("postgres:16", errors.New("registry unreachable"))
		_, cmd := doComposePull(m)
		msg, ok := cmd().(state.ComposeServicePullCompleted)
		if !ok {
			t.Fatalf("cmd returned %T, want ComposeServicePullCompleted", msg)
		}
		if msg.Success != 1 || msg.Error == nil {
			t.Errorf("summary = success=%d err=%v, want 1 non-nil error", msg.Success, msg.Error)
		}
		if !strings.Contains(msg.Error.Error(), "postgres:16") {
			t.Errorf("error = %v, want postgres:16 ref", msg.Error)
		}
	})

	t.Run("no project", func(t *testing.T) {
		m, _ := newPull()
		m.Resources.Containers.Items = nil
		updated, cmd := doComposePull(m)
		if cmd != nil {
			t.Fatalf("cmd = %v, want nil", cmd)
		}
		if updated.Feedback.ToastMessage != "" {
			t.Errorf("toast = %q, want no toast without a project", updated.Feedback.ToastMessage)
		}
	})
}

// TestComposeBuildCapabilityToast verifies the R08-05 build capability
// gate: the mock engine does not advertise CapabilityComposeBuild, so
// doComposeBuild degrades to the spec-mandated CLI hint toast.
func TestComposeBuildCapabilityToast(t *testing.T) {
	m := newAppModelWithEngine(newMockEngine())
	m.Navigation.ActivePanel = state.PanelCompose
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{
		{ID: "c1", Name: "web", Image: "demo_web:latest", ComposeProject: "demo", ComposeService: "web"},
	}

	updated, cmd := doComposeBuild(m)
	if cmd != nil {
		t.Fatalf("cmd = %v, want nil (capability-gated)", cmd)
	}
	if !strings.Contains(updated.Feedback.ToastMessage, "requires compose CLI") {
		t.Errorf("toast = %q, want requires-compose-CLI hint", updated.Feedback.ToastMessage)
	}
	if updated.Feedback.ToastLevel != state.NotificationWarning {
		t.Errorf("toast level = %v, want warning", updated.Feedback.ToastLevel)
	}
}
