package actionbar

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime/docker"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestContainersExposeOnlyComplexActions(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), &dockerclient.Client{}, "test")
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{ID: "one", Name: "api", State: state.ContainerStateRunning}}
	items := VisibleItems(m)
	want := []keys.KeyAction{keys.ActionContainerRename, keys.ActionContainerTop, keys.ActionContainerPort, keys.ActionContainerDiff, keys.ActionContainerWait, keys.ActionContainerCopy, keys.ActionContainerUpdate, keys.ActionContainerExport, keys.ActionContainerCommit}
	if len(items) != len(want) {
		t.Fatalf("items = %#v", items)
	}
	for index, action := range want {
		if items[index].Action != action || items[index].Key != "" || items[index].Disabled {
			t.Fatalf("item %d = %#v", index, items[index])
		}
	}
}

func TestDirectAndGlobalActionsAreExcluded(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), nil, "test")
	// Image History is a no-shortcut complex action; other panels have
	// no Action Bar items until a per-panel set is added.
	for _, panel := range []state.PanelType{state.PanelVolumes, state.PanelNetworks, state.PanelCompose, state.PanelAudit} {
		m.Navigation.ActivePanel = panel
		if items := VisibleItems(m); len(items) != 0 {
			t.Fatalf("panel %v contains direct/global actions: %#v", panel, items)
		}
	}
}

func TestContainerActionAvailabilityAndFiltering(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), nil, "test")
	items := VisibleItems(m)
	for _, item := range items {
		if !item.Disabled {
			t.Fatalf("missing container action should be disabled: %#v", item)
		}
	}
	m.Navigation.ActionBar.Filter = "rename"
	items = VisibleItems(m)
	if len(items) != 1 || items[0].Action != keys.ActionContainerRename {
		t.Fatalf("filtered items = %#v", items)
	}
}

func TestImagesExposeHistoryWhenEligible(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), &dockerclient.Client{}, "test")
	m.Navigation.ActivePanel = state.PanelImages
	m.Resources.Images.Items = []runtimeapi.ImageSummary{{ID: "sha256:one", RepoTags: []string{"nginx:latest"}}}
	items := VisibleItems(m)
	if len(items) != 1 || items[0].Action != keys.ActionImageHistory || items[0].Disabled {
		t.Fatalf("image actions = %#v", items)
	}
	m.Resources.Images.Items[0].IsManifest = true
	if items = VisibleItems(m); len(items) != 1 || !items[0].Disabled {
		t.Fatalf("manifest history should be disabled: %#v", items)
	}
}
