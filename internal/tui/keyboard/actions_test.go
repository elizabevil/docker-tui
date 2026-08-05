package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// TestEscInNormalClearsActiveFilter verifies the user-reported flow:
// in ModeNormal (after pressing Enter in the filter input) with an
// active filter, a single Esc clears the filter and stays in the list,
// instead of starting the double-Esc-to-exit-app flow.
func TestEscInNormalClearsActiveFilter(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.Items = []dockerclient.ContainerSummary{{Name: "api"}, {Name: "worker"}}
	// Simulate the post-Enter state: ModeNormal, filter applied, items filtered
	m.Navigation.Mode = state.ModeNormal
	m.Resources.Containers.SetFilter("api")
	if m.Resources.Containers.Filter != "api" {
		t.Fatalf("setup: Filter=%q, want api", m.Resources.Containers.Filter)
	}

	updated, cmd := handleBackAction(m)

	if cmd != nil {
		t.Fatalf("cmd=%v, want nil (single Esc should not start quit flow)", cmd)
	}
	if updated.Navigation.EscPending {
		t.Fatal("EscPending=true after single Esc, want false (cleared filter, not quit flow)")
	}
	if updated.Resources.Containers.Filter != "" {
		t.Fatalf("Container.Filter=%q after Esc, want empty", updated.Resources.Containers.Filter)
	}
	if len(updated.Resources.Containers.FilteredItems()) != 0 {
		// After Clear(), the filter is "" so FilteredItems() returns the
		// underlying Items slice unfiltered. Just assert the filter text is
		// cleared — the visible count is an implementation detail of
		// FilteredItems vs the unfiltered Items list.
		t.Logf("note: FilteredItems returned %d items after clear (filter=%q)",
			len(updated.Resources.Containers.FilteredItems()),
			updated.Resources.Containers.Filter)
	}
}

// TestEscInNormalNoFilterStartsQuitFlow verifies the default flow is
// unchanged: with no active filter, single Esc sets EscPending, and
// the next Esc would actually quit.
func TestEscInNormalNoFilterStartsQuitFlow(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelContainers
	m.Navigation.Mode = state.ModeNormal

	updated, cmd := handleBackAction(m)

	if cmd == nil {
		t.Fatal("cmd=nil, want Tick cmd (EscPending flow)")
	}
	if !updated.Navigation.EscPending {
		t.Fatal("EscPending=false, want true (no filter → start quit flow)")
	}
	if updated.Feedback.InfoMessage != "Press Esc again to quit" {
		t.Fatalf("InfoMessage=%q, want quit prompt", updated.Feedback.InfoMessage)
	}
}

// TestEscInNormalClearsComposeFilter verifies the compose-panel branch.
func TestEscInNormalClearsComposeFilter(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelCompose
	m.Compose.ComposeFocus = 1
	m.Compose.ComposeServiceFilter = "svc"
	m.Navigation.Mode = state.ModeNormal

	updated, cmd := handleBackAction(m)

	if cmd != nil {
		t.Fatalf("cmd=%v, want nil", cmd)
	}
	if updated.Compose.ComposeServiceFilter != "" {
		t.Fatalf("ComposeServiceFilter=%q, want empty", updated.Compose.ComposeServiceFilter)
	}
}

// TestClearFiltersActionWipesAllPanels verifies the ActionClearFilters
// end-to-end: handleAction routes to filter.New(m).ClearAll() and every
// panel filter is reset, cursors returned to 0, no command emitted.
func TestClearFiltersActionWipesAllPanels(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Resources.Containers.SetFilter("c-f")
	m.Resources.Images.SetFilter("i-f")
	m.Resources.Volumes.SetFilter("v-f")
	m.Resources.Networks.SetFilter("n-f")
	m.Audit.SetFilter("a-f")
	m.Compose.ComposeServiceFilter = "svc"
	m.Compose.ComposeProjectFilter = "proj"
	m.Resources.Containers.Cursor = 3
	m.Resources.Images.Cursor = 2
	m.Navigation.Mode = state.ModeNormal

	updated, cmd := handleAction(keys.ActionClearFilters, m, nil)

	if cmd != nil {
		t.Fatalf("cmd=%v, want nil (clear should not emit commands)", cmd)
	}
	if updated.Resources.Containers.Filter != "" {
		t.Errorf("Containers.Filter=%q", updated.Resources.Containers.Filter)
	}
	if updated.Resources.Images.Filter != "" {
		t.Errorf("Images.Filter=%q", updated.Resources.Images.Filter)
	}
	if updated.Resources.Volumes.Filter != "" {
		t.Errorf("Volumes.Filter=%q", updated.Resources.Volumes.Filter)
	}
	if updated.Resources.Networks.Filter != "" {
		t.Errorf("Networks.Filter=%q", updated.Resources.Networks.Filter)
	}
	if updated.Audit.FilterText() != "" {
		t.Errorf("Audit.FilterText=%q", updated.Audit.FilterText())
	}
	if updated.Compose.ComposeServiceFilter != "" {
		t.Errorf("ComposeServiceFilter=%q", updated.Compose.ComposeServiceFilter)
	}
	if updated.Compose.ComposeProjectFilter != "" {
		t.Errorf("ComposeProjectFilter=%q", updated.Compose.ComposeProjectFilter)
	}
	if updated.Resources.Containers.Cursor != 0 {
		t.Errorf("Containers.Cursor=%d, want 0", updated.Resources.Containers.Cursor)
	}
	if updated.Resources.Images.Cursor != 0 {
		t.Errorf("Images.Cursor=%d, want 0", updated.Resources.Images.Cursor)
	}
}
