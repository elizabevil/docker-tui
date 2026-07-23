package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	dockermodel "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestComposeProjectAndServiceHelpers(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), nil, "")
	m.Resources.Containers.Items = []dockermodel.ContainerSummary{
		{ID: "c1", ComposeProject: "proj-b", ComposeService: "api"},
		{ID: "c2", ComposeProject: "proj-a", ComposeService: "web"},
		{ID: "c3", ComposeProject: "proj-a", ComposeService: "db"},
		{ID: "c4", ComposeProject: "", ComposeService: ""},
	}

	projects := composeProjectNames(m)
	if len(projects) != 2 || projects[0] != "proj-a" || projects[1] != "proj-b" {
		t.Fatalf("unexpected projects: %#v", projects)
	}

	services := composeServiceNames(m, "proj-a")
	if len(services) != 2 || services[0] != "db" || services[1] != "web" {
		t.Fatalf("unexpected services for proj-a: %#v", services)
	}

	m.Compose.ComposeCursor = 1
	if got := currentComposeProject(m); got != "proj-b" {
		t.Fatalf("expected proj-b, got %q", got)
	}
}
