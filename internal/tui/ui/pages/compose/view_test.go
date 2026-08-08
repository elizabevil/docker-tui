package compose

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/data/runtime/mockengine"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func TestRenderProjectDetailUsesTableLayout(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "")
	app.Resources.Containers.Items = []runtimeapi.ContainerSummary{
		{ID: "one", ComposeProject: "proj-a", ComposeService: "api", Image: "nginx:latest", State: state.ContainerStateRunning},
		{ID: "two", ComposeProject: "proj-a", ComposeService: "worker", Image: "worker:latest", State: state.ContainerStateStopped},
	}
	got := component.StripANSI(RenderProjectDetailTable(app, 100, 12))
	for _, expected := range []string{"SERVICE", "IMAGE", "STATUS", "api", "nginx:latest", "1/2 running"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("compose detail table missing %q: %q", expected, got)
		}
	}
	if strings.Contains(got, "Project:") || strings.Contains(got, "快捷操作") {
		t.Fatalf("compose detail leaked legacy text sections: %q", got)
	}
}

func TestServicePanelUsesBreadcrumbOrder(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "")
	got := component.StripANSI(renderServicePanel(app, composeProj{
		name: "integration",
		svcs: map[string]composeSvc{
			"api": {count: 1, image: "nginx:latest", running: 1, total: 1},
		},
		total:   1,
		running: 1,
	}, 80, 12))
	if !strings.Contains(got, "Compose > integration > Services") {
		t.Fatalf("service breadcrumb = %q", got)
	}
	if strings.Contains(got, "Services: integration") {
		t.Fatalf("legacy service title remains: %q", got)
	}
	// Service count summary now lives in the header row alongside the title;
	// the previous "s:start S:stop l:logs" FooterHint has been removed from
	// this view because the same shortcuts show elsewhere in the chrome.
	if !strings.Contains(got, "Services: 1") {
		t.Fatalf("service count must surface in the header, got %q", got)
	}
	if strings.Contains(got, "s:start S:stop l:logs") {
		t.Fatalf("duplicate shortcut hint must not render in the services panel: %q", got)
	}
}

// TestRenderComposeContainersHidesPodColumnOnDocker pins §4.3
// decision C "docker 下完全不渲染 Pod 列": when the engine lacks
// CapabilityComposePodScope, the container sub-view's column set
// must not include the pod column — header not rendered, column
// space reclaimed for the remaining columns.
func TestRenderComposeContainersHidesPodColumnOnDocker(t *testing.T) {
	eng := mockengine.New()
	eng.SetCapability(runtimeapi.CapabilityComposePodScope, runtimeapi.CapabilityInfo{Support: runtimeapi.Unsupported})

	m := state.NewAppModel(config.DefaultAppConfig(), eng, "test")
	m.Compose.ComposeContainerViewID = "web"
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{
		ID: "abc123", Name: "web-1", State: state.ContainerStateRunning,
		ComposeProject: "demo", ComposeService: "web",
	}}

	out := component.StripANSI(renderComposeContainers(m, 120, 24))
	if strings.Contains(out, "POD") {
		t.Fatalf("pod column rendered without capability:\n%s", out)
	}
}

// TestRenderComposeContainersShowsPodColumnOnPodman verifies §4.3
// decision C: when the engine advertises CapabilityComposePodScope,
// the pod column IS rendered. Cell-value rendering is a separate
// concern pinned by TestRenderComposeContainersPodValue in T7.
func TestRenderComposeContainersShowsPodColumnOnPodman(t *testing.T) {
	eng := mockengine.New()
	eng.SetCapability(runtimeapi.CapabilityComposePodScope, runtimeapi.CapabilityInfo{Support: runtimeapi.Available})

	m := state.NewAppModel(config.DefaultAppConfig(), eng, "test")
	m.Compose.ComposeContainerViewID = "web"
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{
		ID: "abc123", Name: "web-1", State: state.ContainerStateRunning,
		ComposeProject: "demo", ComposeService: "web",
		CoLocatedGroupID: "pod_demo",
	}}

	out := component.StripANSI(renderComposeContainers(m, 120, 24))
	if !strings.Contains(out, "POD") {
		t.Fatalf("pod column missing with capability Available:\n%s", out)
	}
}
