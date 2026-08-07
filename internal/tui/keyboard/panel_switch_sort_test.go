package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestBracketKeysSwitchPanels(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Navigation.ActivePanel = state.PanelContainers

	updated, _ := HandleKeyPress(keyMessage("]"), app)
	if updated.Navigation.ActivePanel != state.PanelImages {
		t.Fatalf("] from containers: panel=%v, want images", updated.Navigation.ActivePanel)
	}

	app.Navigation.ActivePanel = state.PanelContainers
	updated, _ = HandleKeyPress(keyMessage("["), app)
	if updated.Navigation.ActivePanel != state.PanelAudit {
		t.Fatalf("[ from containers: panel=%v, want audit (wrap-around)", updated.Navigation.ActivePanel)
	}
}

func TestBracketKeysWrapAround(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Navigation.ActivePanel = state.PanelAudit

	updated, _ := HandleKeyPress(keyMessage("]"), app)
	if updated.Navigation.ActivePanel != state.PanelContainers {
		t.Fatalf("] from audit: panel=%v, want containers (wrap-around)", updated.Navigation.ActivePanel)
	}
}

func TestNSetsAscendingSort(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Navigation.ActivePanel = state.PanelContainers
	app.Resources.Containers.SortAsc = false
	app.Resources.Containers.Items = []dockerclient.ContainerSummary{{ID: "c1"}, {ID: "c2"}}

	updated, _ := HandleKeyPress(keyMessage("n"), app)
	if !updated.Resources.Containers.SortAsc {
		t.Fatal("n did not set ascending sort")
	}
}

func TestCtrlNSetsDescendingSort(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Navigation.ActivePanel = state.PanelContainers
	app.Resources.Containers.SortAsc = true
	app.Resources.Containers.Items = []dockerclient.ContainerSummary{{ID: "c1"}, {ID: "c2"}}

	updated, _ := HandleKeyPress(keyMessage("ctrl+n"), app)
	if updated.Resources.Containers.SortAsc {
		t.Fatal("ctrl+n did not set descending sort")
	}
}

func TestSortShortcutsApplyToImages(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Navigation.ActivePanel = state.PanelImages
	app.Resources.Images.SortAsc = false

	updated, _ := HandleKeyPress(keyMessage("n"), app)
	if !updated.Resources.Images.SortAsc {
		t.Fatal("n did not set ascending sort on images panel")
	}

	updated, _ = HandleKeyPress(keyMessage("ctrl+n"), updated)
	if updated.Resources.Images.SortAsc {
		t.Fatal("ctrl+n did not set descending sort on images panel")
	}
}

func TestSortShortcutsIgnoredOnNonSortablePanel(t *testing.T) {
	app := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	app.Navigation.ActivePanel = state.PanelVolumes

	updated, cmd := HandleKeyPress(keyMessage("n"), app)
	if cmd != nil {
		t.Fatalf("n on volumes: cmd=%v, want nil", cmd)
	}
	if updated.Navigation.ActivePanel != state.PanelVolumes {
		t.Fatalf("n on volumes changed panel to %v", updated.Navigation.ActivePanel)
	}
}
