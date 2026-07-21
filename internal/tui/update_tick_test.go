package tui

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestHandleFilterExitTimeoutKeepsFilterActive(t *testing.T) {
	app := &state.AppModel{
		Mode:              state.ModeFilter,
		FilterExitPending: true,
		FilterExitToken:   2,
		FilterInput:       state.QueryInputState{Text: "api", Cursor: 3},
	}

	updated, _ := handleFilterExitTimeout(app, state.FilterExitTimeout{Token: 2})
	if updated.FilterExitPending {
		t.Fatal("filter exit window remained pending after timeout")
	}
	if updated.Mode != state.ModeFilter || updated.FilterInput.Text != "api" {
		t.Fatalf("timeout changed active filter: mode=%v text=%q", updated.Mode, updated.FilterInput.Text)
	}
}

func TestHandleFilterExitTimeoutIgnoresStaleWindow(t *testing.T) {
	app := &state.AppModel{Mode: state.ModeFilter, FilterExitPending: true, FilterExitToken: 3}
	updated, _ := handleFilterExitTimeout(app, state.FilterExitTimeout{Token: 2})
	if !updated.FilterExitPending {
		t.Fatal("stale timeout cleared the current filter exit window")
	}
}
