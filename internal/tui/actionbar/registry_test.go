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
	m := state.NewAppModel(config.DefaultAppConfig(), &dockerclient.Client{}, "test")
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
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
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
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
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

// TestBuildItemsForScopeSkipsInActionBarFalse pins the C04 contract
// directly on the filter: Operations declaring inActionBar=false must
// never surface as action-bar items, even though they live in the same
// scope slice. ScopeForPanel already keeps Volume/Network panels out of
// the action bar; this unit test locks the filter itself so the
// two-layer guard cannot silently regress.
func TestBuildItemsForScopeSkipsInActionBarFalse(t *testing.T) {
	ops := &config.Operations{
		ByScope: map[config.OperationScope][]config.OperationSpec{
			config.OperationScopeContainer: {
				{Action: "containerRename", Label: "Rename", InActionBar: true},
				{Action: "containerTop", Label: "Top", InActionBar: false},
			},
		},
	}
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	items := buildItemsForScope(m, ops, config.OperationScopeContainer)
	if len(items) != 1 || items[0].Action != keys.ActionContainerRename {
		t.Fatalf("items = %#v, want exactly [containerRename]", items)
	}
}

func TestImagesExposeHistoryWhenEligible(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), &dockerclient.Client{}, "test")
	m.Navigation.ActivePanel = state.PanelImages
	m.Resources.Images.Items = []runtimeapi.ImageSummary{{ID: "sha256:one", RepoTags: []string{"nginx:latest"}}}
	items := VisibleItems(m)
	wantActions := []keys.KeyAction{keys.ActionImageHistory, keys.ActionImagePrune}
	if len(items) != len(wantActions) {
		t.Fatalf("image actions = %#v, want %d items", items, len(wantActions))
	}
	for i, action := range wantActions {
		if items[i].Action != action || items[i].Disabled {
			t.Fatalf("item %d = %#v, want action=%q disabled=false", i, items[i], action)
		}
	}
	m.Resources.Images.Items[0].IsManifest = true
	if items = VisibleItems(m); len(items) != len(wantActions) {
		t.Fatalf("image actions after manifest: %#v", items)
	}
	if !items[0].Disabled {
		t.Fatalf("manifest history should be disabled: %#v", items[0])
	}
	if items[1].Disabled {
		t.Fatalf("imagePrune should remain enabled with manifest image: %#v", items[1])
	}
}
