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
	app.Containers.Items = []dockerclient.ContainerSummary{{Name: "api"}, {Name: "worker"}}
	ToFilter(app)

	updated, _ := handleFilterInput("a", app)
	if updated.Containers.Filter != "a" || len(updated.Containers.FilteredItems()) != 1 {
		t.Fatalf("filter=%q items=%d, want immediate match", updated.Containers.Filter, len(updated.Containers.FilteredItems()))
	}
}

func TestFilterRequiresDoubleEscToClearAndExit(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.FilterInput = state.QueryInputState{Text: "api", Cursor: 3}
	app.Mode = state.ModeFilter
	ApplyFilter(app)

	updated, cmd := handleFilterInput(keys.KeyEsc, app)
	if cmd == nil || !updated.FilterExitPending || updated.Mode != state.ModeFilter || updated.Containers.Filter != "api" {
		t.Fatalf("first Esc changed filter state: mode=%v pending=%v filter=%q", updated.Mode, updated.FilterExitPending, updated.Containers.Filter)
	}

	updated, _ = handleFilterInput(keys.KeyEsc, updated)
	if updated.Mode != state.ModeNormal || updated.FilterExitPending || updated.Containers.Filter != "" {
		t.Fatalf("second Esc did not clear filter: mode=%v pending=%v filter=%q", updated.Mode, updated.FilterExitPending, updated.Containers.Filter)
	}
}

func TestLogSearchAppliesOnEnterAndCancelPreservesCurrentQuery(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Mode = state.ModeLogView
	app.LogContainerID = "container"
	app.LogContent = []string{"ready", "request failed", "failed again"}
	ToSearch(app)
	app.SearchInput = state.QueryInputState{Text: "failed", Cursor: 6}

	updated := handleSearchInput(keys.KeyEnter, app)
	if updated.Mode != state.ModeLogView || updated.LogSearchText != "failed" || updated.LogViewOffset != 1 {
		t.Fatalf("search apply mode=%v query=%q offset=%d", updated.Mode, updated.LogSearchText, updated.LogViewOffset)
	}

	ToSearch(updated)
	updated.SearchInput.Text = "other"
	updated = handleSearchInput(keys.KeyEsc, updated)
	if updated.Mode != state.ModeLogView || updated.LogSearchText != "failed" {
		t.Fatalf("search cancel mode=%v query=%q", updated.Mode, updated.LogSearchText)
	}
}

func TestLogFilterBindingOpensSearchMode(t *testing.T) {
	app := state.NewAppModel(config.DefaultConfig(), nil, "test")
	app.Mode = state.ModeLogView
	app.LogContainerID = "container"

	updated, _ := HandleKeyPress(keyMessage(keys.KeySlash), app)
	if updated.Mode != state.ModeSearch {
		t.Fatalf("log filter binding mode=%v, want search", updated.Mode)
	}
}
