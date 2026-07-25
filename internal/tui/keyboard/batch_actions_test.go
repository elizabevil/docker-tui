package keyboard

import (
	"errors"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/mockengine"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// runBatch executes cmd and asserts the returned message is a BatchActioned.
func runBatch(t *testing.T, cmd tea.Cmd) state.BatchActioned {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected batch command, got nil")
	}
	msg := cmd()
	batch, ok := msg.(state.BatchActioned)
	if !ok {
		t.Fatalf("cmd returned %T, want state.BatchActioned", msg)
	}
	return batch
}

// newMockEngine returns a fresh mockengine.Engine. Wrapped so tests stay
// self-documenting.
func newMockEngine() *mockengine.Engine { return mockengine.New() }

// newAppModelWithEngine builds a state.AppModel wired to a runtime.Engine
// stub.
func newAppModelWithEngine(eng runtimeapi.Engine) *state.AppModel {
	m := state.NewAppModel(config.DefaultConfig(), nil, "test")
	m.Connection.Engine = eng
	return m
}

// markIDs populates the selection marks with the provided IDs.
func markIDs(m *state.AppModel, ids ...string) {
	m.Selection.MarkedIDs = make(map[string]bool, len(ids))
	for _, id := range ids {
		m.Selection.MarkedIDs[id] = true
	}
}

// TestExecuteBatchActionAggregatesResults verifies that a container batch
// start returns one BatchActioned message with the per-target breakdown
// instead of N independent messages.
func TestExecuteBatchActionAggregatesResults(t *testing.T) {
	eng := newMockEngine()
	eng.Fail("b", errors.New("simulated engine error"))

	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{
		{ID: "a", Name: "a", State: state.ContainerStateExited},
		{ID: "b", Name: "b", State: state.ContainerStateExited},
	}
	markIDs(m, "a", "b")

	trace := beginAudit(m, "resource.container.start", audit.ContainerTarget{Name: "2 containers"}, "Batch start 2 containers")
	_, cmd := executeBatchAction(m, "start", trace)
	batch := runBatch(t, cmd)

	if batch.Total != 2 || batch.Success != 1 || batch.Failed != 1 {
		t.Errorf("unexpected summary: total=%d success=%d failed=%d", batch.Total, batch.Success, batch.Failed)
	}
	if batch.Resource != state.ResourceContainer {
		t.Errorf("resource = %q, want container", batch.Resource)
	}
	if batch.Scope != "container.batch.start" {
		t.Errorf("scope = %q, want container.batch.start", batch.Scope)
	}
	if len(batch.FailedIDs) != 1 || batch.FailedIDs[0] != "b" {
		t.Errorf("failedIDs = %v, want [b]", batch.FailedIDs)
	}
	if batch.Error == nil || batch.Error.Error() == "" {
		t.Error("expected joined error to surface the engine failure")
	}
}

// TestExecuteBatchActionRejectsUnknownAction verifies the unknown-action
// guard returns nil instead of a command and surfaces a toast.
func TestExecuteBatchActionRejectsUnknownAction(t *testing.T) {
	eng := newMockEngine()
	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelContainers
	markIDs(m, "a")

	_, cmd := executeBatchAction(m, "frobnicate", audit.Trace{})
	if cmd != nil {
		t.Fatalf("unknown action should return nil cmd, got %T", cmd)
	}
	if m.Feedback.ToastMessage == "" {
		t.Error("expected toast for unknown batch action")
	}
}

// TestExecuteBatchActionEmptyMarksIsNoop guards against races where the
// user cancels mid-confirmation and the marks are already cleared.
func TestExecuteBatchActionEmptyMarksIsNoop(t *testing.T) {
	eng := newMockEngine()
	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelContainers

	_, cmd := executeBatchAction(m, "start", audit.Trace{})
	if cmd != nil {
		t.Fatalf("empty marks should return nil cmd, got %T", cmd)
	}
	if got := eng.CallCount(); got != 0 {
		t.Errorf("engine should not be invoked, got %d calls", got)
	}
}

// TestExecuteBulkDeleteAggregatesImages verifies the bulk-delete path on
// the images panel collects per-target outcomes into a single summary.
func TestExecuteBulkDeleteAggregatesImages(t *testing.T) {
	eng := newMockEngine()
	eng.Fail("img-2", errors.New("image busy"))

	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelImages
	m.Resources.Images.Items = []runtimeapi.ImageSummary{
		{ID: "img-1"}, {ID: "img-2"},
	}
	markIDs(m, "img-1", "img-2")

	trace := beginAudit(m, "resource.image.delete", audit.ImageTarget{ID: "bulk", Name: "2 items"}, "Delete 2 items?")
	_, cmd := executeBulkDelete(m, trace)
	batch := runBatch(t, cmd)

	if batch.Total != 2 || batch.Success != 1 || batch.Failed != 1 {
		t.Errorf("unexpected summary: %+v", batch)
	}
	if batch.Scope != "bulk-delete" {
		t.Errorf("scope = %q", batch.Scope)
	}
	if batch.Resource != state.ResourceImage {
		t.Errorf("resource = %q, want image", batch.Resource)
	}
	if len(batch.FailedIDs) != 1 || batch.FailedIDs[0] != "img-2" {
		t.Errorf("failedIDs = %v", batch.FailedIDs)
	}
}

// TestExecuteBulkDeleteContainerPanelSanity verifies the container panel
// takes the same code path.
func TestExecuteBulkDeleteContainerPanelSanity(t *testing.T) {
	eng := newMockEngine()
	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{ID: "c-1"}}
	markIDs(m, "c-1")

	_, cmd := executeBulkDelete(m, audit.Trace{})
	batch := runBatch(t, cmd)

	if batch.Total != 1 || batch.Success != 1 {
		t.Errorf("summary = %+v", batch)
	}
	if batch.Resource != state.ResourceContainer {
		t.Errorf("resource = %q", batch.Resource)
	}
}

// TestComposeStartAggregatesContainers verifies compose start collects
// per-container outcomes into a single BatchActioned message.
func TestComposeStartAggregatesContainers(t *testing.T) {
	eng := newMockEngine()
	eng.Fail("c2", errors.New("container start failed"))

	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelCompose
	m.Compose.ComposeDetailProject = "demo"
	m.Compose.ComposeServiceCursor = 0
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{
		{ID: "c1", Name: "web", ComposeProject: "demo", ComposeService: "web"},
		{ID: "c2", Name: "db", ComposeProject: "demo", ComposeService: "db"},
	}

	_, cmd := doComposeStart(m)
	batch := runBatch(t, cmd)

	if batch.Total != 2 || batch.Success != 1 || batch.Failed != 1 {
		t.Errorf("summary = %+v", batch)
	}
	if batch.Scope != "compose.start" {
		t.Errorf("scope = %q", batch.Scope)
	}
	if batch.Resource != state.ResourceComposeProject {
		t.Errorf("resource = %q", batch.Resource)
	}
	if len(batch.FailedIDs) != 1 || batch.FailedIDs[0] != "c2" {
		t.Errorf("failedIDs = %v", batch.FailedIDs)
	}
}

// TestComposeStopAggregatesContainers mirrors TestComposeStartAggregatesContainers
// for the stop path.
func TestComposeStopAggregatesContainers(t *testing.T) {
	eng := newMockEngine()
	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelCompose
	m.Compose.ComposeDetailProject = "demo"
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{
		{ID: "c1", Name: "web", ComposeProject: "demo", ComposeService: "web"},
		{ID: "c2", Name: "db", ComposeProject: "demo", ComposeService: "db"},
	}
	_, cmd := doComposeStop(m)
	batch := runBatch(t, cmd)
	if batch.Total != 2 || batch.Success != 2 {
		t.Errorf("summary = %+v", batch)
	}
	if batch.Scope != "compose.stop" {
		t.Errorf("scope = %q", batch.Scope)
	}
}

// TestComposeDownAggregatesResources verifies compose down collects
// container + volume + network outcomes into a single summary.
func TestComposeDownAggregatesResources(t *testing.T) {
	eng := newMockEngine()
	eng.Fail("vol-data", errors.New("volume remove failed"))

	m := newAppModelWithEngine(eng)
	m.Navigation.ActivePanel = state.PanelCompose
	m.Compose.ComposeDetailProject = "demo"
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{
		{ID: "c1", Name: "web", ComposeProject: "demo", ComposeService: "web"},
	}
	m.Resources.Volumes.Items = []runtimeapi.Volume{
		{Name: "vol-data", Labels: map[string]string{"com.docker.compose.project": "demo"}},
	}
	m.Resources.Networks.Items = []runtimeapi.Network{
		{ID: "net-demo", Labels: map[string]string{"com.docker.compose.project": "demo"}},
	}

	_, cmd := doComposeDown(m)
	batch := runBatch(t, cmd)
	if batch.Total != 3 || batch.Success != 2 || batch.Failed != 1 {
		t.Errorf("summary = %+v", batch)
	}
	if batch.Scope != "compose.down" {
		t.Errorf("scope = %q", batch.Scope)
	}
	if len(batch.FailedIDs) != 1 || batch.FailedIDs[0] != "vol-data" {
		t.Errorf("failedIDs = %v", batch.FailedIDs)
	}
}
