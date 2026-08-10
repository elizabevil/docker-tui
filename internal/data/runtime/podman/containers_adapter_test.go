package podman

import (
	"testing"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

func TestMapContainerNetworksFieldInitializedEmpty(t *testing.T) {
	items := []dto.ContainerItem{{
		ID:       "10b2c97269cc",
		Names:    []string{"redis"},
		Image:    "redis:7-alpine",
		Networks: []string{"bridge", "myNet"},
	}}
	mapped := MapContainerSummaries(items)
	if len(mapped) != 1 {
		t.Fatalf("expected one summary, got %d", len(mapped))
	}
	if mapped[0].Networks == nil {
		t.Fatalf("podman Networks map not initialized: %+v", mapped[0].Networks)
	}
	if len(mapped[0].IPs) != 0 {
		t.Fatalf("podman IPs slice should be empty without inspect data: %+v", mapped[0].IPs)
	}
	_ = dockerclient.ContainerSummary{}
}