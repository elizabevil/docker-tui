package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestFilterInputAppliesImmediately(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Resources.Containers.Items = []dockerclient.ContainerSummary{{Name: "api"}, {Name: "worker"}}
	ToFilter(app)

	updated, _ := handleFilterInput("a", app)
	if updated.Resources.Containers.Filter != "a" || len(updated.Resources.Containers.FilteredItems()) != 1 {
		t.Fatalf("filter=%q items=%d, want immediate match", updated.Resources.Containers.Filter, len(updated.Resources.Containers.FilteredItems()))
	}
}

func TestFilterRequiresDoubleEscToClearAndExit(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Navigation.FilterInput = state.QueryInputState{Text: "api", Cursor: 3}
	app.Navigation.Mode = state.ModeFilter
	ApplyFilter(app)

	updated, cmd := handleFilterInput(keys.KeyEsc, app)
	if cmd == nil || !updated.Navigation.FilterExitPending || updated.Navigation.Mode != state.ModeFilter || updated.Resources.Containers.Filter != "api" {
		t.Fatalf("first Esc changed filter state: mode=%v pending=%v filter=%q", updated.Navigation.Mode, updated.Navigation.FilterExitPending, updated.Resources.Containers.Filter)
	}

	updated, _ = handleFilterInput(keys.KeyEsc, updated)
	if updated.Navigation.Mode != state.ModeNormal || updated.Navigation.FilterExitPending || updated.Resources.Containers.Filter != "" {
		t.Fatalf("second Esc did not clear filter: mode=%v pending=%v filter=%q", updated.Navigation.Mode, updated.Navigation.FilterExitPending, updated.Resources.Containers.Filter)
	}
}

func TestLogSearchAppliesOnEnterAndCancelPreservesCurrentQuery(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Navigation.Mode = state.ModeLogView
	app.Log.LogContainerID = "container"
	app.Log.LogContent = []string{"ready", "request failed", "failed again"}
	ToSearch(app)
	app.Navigation.SearchInput = state.QueryInputState{Text: "failed", Cursor: 6}

	updated := handleSearchInput(keys.KeyEnter, app)
	if updated.Navigation.Mode != state.ModeLogView || updated.Log.LogSearchText != "failed" || updated.Log.LogViewOffset != 1 {
		t.Fatalf("search apply mode=%v query=%q offset=%d", updated.Navigation.Mode, updated.Log.LogSearchText, updated.Log.LogViewOffset)
	}

	ToSearch(updated)
	updated.Navigation.SearchInput.Text = "other"
	updated = handleSearchInput(keys.KeyEsc, updated)
	if updated.Navigation.Mode != state.ModeLogView || updated.Log.LogSearchText != "failed" {
		t.Fatalf("search cancel mode=%v query=%q", updated.Navigation.Mode, updated.Log.LogSearchText)
	}
}

func TestLogFilterBindingOpensSearchMode(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Navigation.Mode = state.ModeLogView
	app.Log.LogContainerID = "container"

	updated, _ := HandleKeyPress(keyMessage(keys.KeySlash), app)
	if updated.Navigation.Mode != state.ModeSearch {
		t.Fatalf("log filter binding mode=%v, want search", updated.Navigation.Mode)
	}
}
