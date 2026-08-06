package compose

import (
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
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
