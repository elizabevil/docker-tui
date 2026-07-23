package tui

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestContainerRefreshRestoresSelectionAnchor(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), nil, "test")
	m.Resources.Containers.SelectionAnchorID = "two"
	updated, _ := handleContainersLoaded(m, state.ContainersLoaded{Containers: []dockerclient.ContainerSummary{
		{ID: "one", Name: "a"},
		{ID: "two", Name: "renamed"},
	}})
	if updated.Resources.Containers.Cursor != 1 || updated.Resources.Containers.SelectionAnchorID != "" {
		t.Fatalf("selection was not restored: cursor=%d anchor=%q", updated.Resources.Containers.Cursor, updated.Resources.Containers.SelectionAnchorID)
	}
}
