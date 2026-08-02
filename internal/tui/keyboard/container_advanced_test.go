package keyboard

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/mockengine"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// trackerReadCloser records Close invocations so tests can assert reader
// teardown even when a copy/export fails.
type trackerReadCloser struct {
	mu     sync.Mutex
	closed bool
	data   []byte
}

func (t *trackerReadCloser) Read(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p, t.data)
	t.data = t.data[n:]
	return n, nil
}

func (t *trackerReadCloser) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return nil
}

func (t *trackerReadCloser) isClosed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closed
}

// stubContainerService implements runtimeapi.ContainerService with pluggable
// diff / wait / export / copy behaviors. Unused methods are no-op stubs.
type stubContainerService struct {
	diffFn   func(context.Context, string) ([]runtimeapi.ContainerDiffChange, error)
	waitFn   func(context.Context, string, string) (runtimeapi.ContainerWaitResult, error)
	exportFn func(context.Context, string) (io.ReadCloser, error)
	copyFn   func(context.Context, string, string) (io.ReadCloser, error)
}

func (s *stubContainerService) List(context.Context, runtimeapi.ContainerListOptions) ([]runtimeapi.ContainerSummary, error) {
	return nil, nil
}
func (s *stubContainerService) Inspect(context.Context, string) (*runtimeapi.ContainerDetail, error) {
	return nil, nil
}
func (s *stubContainerService) Top(context.Context, string) (runtimeapi.ContainerProcesses, error) {
	return runtimeapi.ContainerProcesses{}, nil
}
func (s *stubContainerService) Stats(context.Context, string) (runtimeapi.ContainerStats, error) {
	return runtimeapi.ContainerStats{}, nil
}
func (s *stubContainerService) Logs(context.Context, string, runtimeapi.ContainerLogOptions) (io.ReadCloser, error) {
	return nil, nil
}
func (s *stubContainerService) Update(context.Context, string, runtimeapi.ContainerUpdateOptions) (runtimeapi.ContainerUpdateResult, error) {
	return runtimeapi.ContainerUpdateResult{}, nil
}
func (s *stubContainerService) Diff(ctx context.Context, id string) ([]runtimeapi.ContainerDiffChange, error) {
	if s.diffFn != nil {
		return s.diffFn(ctx, id)
	}
	return nil, nil
}
func (s *stubContainerService) Export(ctx context.Context, id string) (io.ReadCloser, error) {
	if s.exportFn != nil {
		return s.exportFn(ctx, id)
	}
	return nil, nil
}
func (s *stubContainerService) Commit(context.Context, string, runtimeapi.ContainerCommitOptions) (runtimeapi.ContainerCommitResult, error) {
	return runtimeapi.ContainerCommitResult{}, nil
}
func (s *stubContainerService) Wait(ctx context.Context, id, cond string) (runtimeapi.ContainerWaitResult, error) {
	if s.waitFn != nil {
		return s.waitFn(ctx, id, cond)
	}
	return runtimeapi.ContainerWaitResult{}, nil
}
func (s *stubContainerService) CopyFromContainer(ctx context.Context, id, src string) (io.ReadCloser, error) {
	if s.copyFn != nil {
		return s.copyFn(ctx, id, src)
	}
	return nil, nil
}

var _ runtimeapi.ContainerService = (*stubContainerService)(nil)

// advancedEngine wraps a mock Engine so Containers() returns the stub service.
type advancedEngine struct {
	*mockengine.Engine
	containers runtimeapi.ContainerService
}

func (e *advancedEngine) Containers() runtimeapi.ContainerService { return e.containers }

func newAdvancedEngine(svc runtimeapi.ContainerService) *advancedEngine {
	return &advancedEngine{Engine: mockengine.New(), containers: svc}
}

// instantMsg runs a tea.Cmd immediately and returns its message, or nil when
// the command is nil (used by no-selection tests).
func instantMsg(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

func TestContainerDiffCmdSuccess(t *testing.T) {
	svc := &stubContainerService{
		diffFn: func(_ context.Context, id string) ([]runtimeapi.ContainerDiffChange, error) {
			if id != "c1" {
				t.Fatalf("diff got id %q, want c1", id)
			}
			return []runtimeapi.ContainerDiffChange{{Kind: runtimeapi.ChangeModified, Path: "/etc/app.conf"}}, nil
		},
	}
	msg := containerDiffCmd(newAdvancedEngine(svc), "c1")()
	done, ok := msg.(ContainerDiffDone)
	if !ok {
		t.Fatalf("unexpected message %T, want ContainerDiffDone", msg)
	}
	if done.ContainerID != "c1" {
		t.Errorf("container id = %q, want c1", done.ContainerID)
	}
	if done.Error != nil {
		t.Fatalf("unexpected error: %v", done.Error)
	}
	if len(done.Changes) != 1 || done.Changes[0].Path != "/etc/app.conf" {
		t.Errorf("unexpected changes: %+v", done.Changes)
	}
}

func TestContainerDiffCmdError(t *testing.T) {
	want := errors.New("diff engine error")
	svc := &stubContainerService{
		diffFn: func(_ context.Context, _ string) ([]runtimeapi.ContainerDiffChange, error) {
			return nil, want
		},
	}
	msg := containerDiffCmd(newAdvancedEngine(svc), "c1")()
	done := msg.(ContainerDiffDone)
	if done.Error != want {
		t.Fatalf("error = %v, want %v", done.Error, want)
	}
}

func TestDoContainerDiffNoSelection(t *testing.T) {
	eng := newAdvancedEngine(&stubContainerService{})
	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{} // empty

	_, cmd := doContainerDiff(m)
	if cmd != nil {
		t.Fatalf("no-selection should return nil cmd, got %T", cmd)
	}

	// Wrong panel should also be a no-op.
	m2 := newAppModelWithEngine(eng)
	m2.Navigation.ActivePanel = state.PanelImages
	m2.Resources.Containers.Items = []runtimeapi.ContainerSummary{{ID: "c1"}}
	_, cmd2 := doContainerDiff(m2)
	if cmd2 != nil {
		t.Fatalf("off-panel should return nil cmd, got %T", cmd2)
	}
}

func TestDoContainerDiffSuccess(t *testing.T) {
	svc := &stubContainerService{
		diffFn: func(_ context.Context, _ string) ([]runtimeapi.ContainerDiffChange, error) {
			return []runtimeapi.ContainerDiffChange{{Kind: runtimeapi.ChangeAdded, Path: "/new"}}, nil
		},
	}
	eng := newAdvancedEngine(svc)
	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{ID: "c1", Name: "web"}}

	_, cmd := doContainerDiff(m)
	if cmd == nil {
		t.Fatal("expected a diff command")
	}
	msg := instantMsg(cmd)
	done, ok := msg.(ContainerDiffDone)
	if !ok {
		t.Fatalf("unexpected message %T, want ContainerDiffDone", msg)
	}
	if done.Error != nil {
		t.Fatalf("unexpected error: %v", done.Error)
	}
	if len(done.Changes) != 1 {
		t.Errorf("changes = %d, want 1", len(done.Changes))
	}
}

func TestDoContainerWaitSuccess(t *testing.T) {
	svc := &stubContainerService{
		waitFn: func(_ context.Context, _ string, _ string) (runtimeapi.ContainerWaitResult, error) {
			return runtimeapi.ContainerWaitResult{StatusCode: 42}, nil
		},
	}
	eng := newAdvancedEngine(svc)
	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{ID: "c1", Name: "web"}}

	_, cmd := doContainerWait(m)
	if cmd == nil {
		t.Fatal("expected a wait command")
	}
	msg := instantMsg(cmd)
	done, ok := msg.(ContainerWaitDone)
	if !ok {
		t.Fatalf("unexpected message %T, want ContainerWaitDone", msg)
	}
	if done.Error != nil {
		t.Fatalf("unexpected error: %v", done.Error)
	}
	if done.Result.StatusCode != 42 {
		t.Errorf("status code = %d, want 42", done.Result.StatusCode)
	}
	if done.ContainerID != "c1" {
		t.Errorf("container id = %q, want c1", done.ContainerID)
	}
}

func TestDoContainerWaitError(t *testing.T) {
	want := errors.New("engine wait failure")
	svc := &stubContainerService{
		waitFn: func(_ context.Context, _, _ string) (runtimeapi.ContainerWaitResult, error) {
			return runtimeapi.ContainerWaitResult{}, want
		},
	}
	eng := newAdvancedEngine(svc)
	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{ID: "c1", Name: "c1"}}

	_, cmd := doContainerWait(m)
	if cmd == nil {
		t.Fatal("expected a wait command")
	}
	msg := instantMsg(cmd)
	done, ok := msg.(ContainerWaitDone)
	if !ok {
		t.Fatalf("unexpected message %T, want ContainerWaitDone", msg)
	}
	if done.Error != want {
		t.Fatalf("error = %v, want %v", done.Error, want)
	}
}

func TestDoContainerWaitNoSelection(t *testing.T) {
	eng := newAdvancedEngine(&stubContainerService{})
	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.Items = nil

	_, cmd := doContainerWait(m)
	if cmd != nil {
		t.Fatalf("no-selection should return nil cmd, got %T", cmd)
	}
}

func TestContainerExportClosesReaderOnSuccess(t *testing.T) {
	rc := &trackerReadCloser{data: []byte("tar-bytes")}
	svc := &stubContainerService{
		exportFn: func(_ context.Context, _ string) (io.ReadCloser, error) { return rc, nil },
	}
	eng := newAdvancedEngine(svc)
	dst := filepath.Join(t.TempDir(), "out.tar")

	msg := containerExportCmd(eng, "c1", dst)()
	done, ok := msg.(ContainerExportDone)
	if !ok {
		t.Fatalf("unexpected message %T, want ContainerExportDone", msg)
	}
	if done.Error != nil {
		t.Fatalf("unexpected error: %v", done.Error)
	}
	if done.Bytes != int64(len("tar-bytes")) {
		t.Errorf("bytes = %d, want %d", done.Bytes, int64(len("tar-bytes")))
	}
	if !rc.isClosed() {
		t.Error("reader was not closed after export success")
	}
	b, err := os.ReadFile(dst)
	if err != nil || string(b) != "tar-bytes" {
		t.Errorf("destination content = %q (err %v), want tar-bytes", b, err)
	}
}

func TestContainerExportErrorPropagates(t *testing.T) {
	want := errors.New("export refused")
	svc := &stubContainerService{
		exportFn: func(_ context.Context, _ string) (io.ReadCloser, error) { return nil, want },
	}
	msg := containerExportCmd(newAdvancedEngine(svc), "c1", "/tmp/out.tar")()
	done := msg.(ContainerExportDone)
	if done.Error != want {
		t.Fatalf("error = %v, want %v", done.Error, want)
	}
}

func TestContainerExportClosesReaderOnDestinationError(t *testing.T) {
	rc := &trackerReadCloser{data: []byte("tar")}
	svc := &stubContainerService{
		exportFn: func(_ context.Context, _ string) (io.ReadCloser, error) { return rc, nil },
	}
	eng := newAdvancedEngine(svc)
	// Destination in a non-existent directory forces os.Create to fail.
	badDst := filepath.Join(t.TempDir(), "missing", "out.tar")

	msg := containerExportCmd(eng, "c1", badDst)()
	done := msg.(ContainerExportDone)
	if done.Error == nil {
		t.Fatal("expected an error from an unwritable destination")
	}
	if !rc.isClosed() {
		t.Error("reader must be closed even when export write fails (FD leak guard)")
	}
}

func TestContainerCopyClosesReaderOnSuccess(t *testing.T) {
	rc := &trackerReadCloser{data: []byte("copy-tar")}
	svc := &stubContainerService{
		copyFn: func(_ context.Context, _ string, src string) (io.ReadCloser, error) {
			if src != "/etc/app.conf" {
				t.Errorf("copy src = %q, want /etc/app.conf", src)
			}
			return rc, nil
		},
	}
	eng := newAdvancedEngine(svc)
	dst := filepath.Join(t.TempDir(), "app.tar")

	msg := containerCopyCmd(eng, "c1", "/etc/app.conf", dst)()
	done, ok := msg.(ContainerCopyDone)
	if !ok {
		t.Fatalf("unexpected message %T, want ContainerCopyDone", msg)
	}
	if done.Error != nil {
		t.Fatalf("unexpected error: %v", done.Error)
	}
	if !rc.isClosed() {
		t.Error("reader was not closed after a successful copy")
	}
	b, err := os.ReadFile(dst)
	if err != nil || string(b) != "copy-tar" {
		t.Errorf("destination content = %q (err %v), want copy-tar", b, err)
	}
}

func TestContainerCopyErrorPropagates(t *testing.T) {
	want := errors.New("copy refused")
	svc := &stubContainerService{
		copyFn: func(_ context.Context, _, _ string) (io.ReadCloser, error) { return nil, want },
	}
	eng := newAdvancedEngine(svc)

	msg := containerCopyCmd(eng, "c1", "/etc/app.conf", "out.tar")()
	done := msg.(ContainerCopyDone)
	if done.Error != want {
		t.Fatalf("error = %v, want %v", done.Error, want)
	}
}

// failingReadCloser returns err after yielding data, simulating a stream that
// dies mid-copy.
type failingReadCloser struct {
	data  []byte
	err   error
	mu    sync.Mutex
	read  bool
	close bool
}

func (f *failingReadCloser) Read(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.read = true
	if len(f.data) > 0 {
		n := copy(p, f.data)
		f.data = f.data[n:]
		return n, nil
	}
	return 0, f.err
}

func (f *failingReadCloser) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.close = true
	return nil
}

func (f *failingReadCloser) wasClosed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.close
}

func TestExportFailedCopyPreservesExistingDestination(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "out.tar")
	if err := os.WriteFile(dst, []byte("precious-existing-data"), 0o644); err != nil {
		t.Fatalf("seed dst: %v", err)
	}
	rc := &failingReadCloser{data: []byte("partial"), err: errors.New("stream cut mid-copy")}
	svc := &stubContainerService{
		exportFn: func(_ context.Context, _ string) (io.ReadCloser, error) { return rc, nil },
	}
	eng := newAdvancedEngine(svc)

	msg := containerExportCmd(eng, "c1", dst)()
	done := msg.(ContainerExportDone)
	if done.Error == nil {
		t.Fatal("expected a mid-copy failure")
	}
	if !rc.wasClosed() {
		t.Error("reader must be closed after a failed copy")
	}
	b, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read existing dst: %v", err)
	}
	if string(b) != "precious-existing-data" {
		t.Errorf("existing destination was clobbered: %q", b)
	}
	// No stray temp file may remain in the destination directory.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("temp file leaked: %v", names)
	}
}

func TestContainerWaitUsesCancellableContext(t *testing.T) {
	ctxSeen := make(chan context.Context, 1)
	svc := &stubContainerService{
		waitFn: func(ctx context.Context, _ string, _ string) (runtimeapi.ContainerWaitResult, error) {
			ctxSeen <- ctx
			<-ctx.Done()
			return runtimeapi.ContainerWaitResult{}, ctx.Err()
		},
	}
	eng := newAdvancedEngine(svc)
	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{ID: "c1", Name: "web"}}

	_, cmd := doContainerWait(m)
	if cmd == nil {
		t.Fatal("expected a wait command")
	}
	// Run the wait command concurrently: the stub blocks inside Wait until the
	// context is cancelled, so the command must be executing while we cancel.
	resultCh := make(chan tea.Msg, 1)
	go func() { resultCh <- cmd() }()

	if waitCtx := <-ctxSeen; waitCtx == nil {
		t.Fatal("wait received nil context")
	}
	m.ContainerWait.Stop()

	msg := <-resultCh
	done, ok := msg.(ContainerWaitDone)
	if !ok {
		t.Fatalf("unexpected message %T, want ContainerWaitDone", msg)
	}
	if !errors.Is(done.Error, context.Canceled) {
		t.Fatalf("wait error = %v, want context.Canceled after Stop()", done.Error)
	}
}
