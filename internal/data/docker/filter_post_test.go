package docker

import (
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func TestPostFilterContainersUsesSameFieldAND(t *testing.T) {
	items := []runtimeapi.ContainerSummary{
		{ID: "one", Labels: map[string]string{"app": "api", "tier": "backend"}},
		{ID: "two", Labels: map[string]string{"app": "api", "tier": "frontend"}},
	}
	filtered, err := postFilterContainers(items, runtimeapi.FilterSet{
		runtimeapi.ContainerFilterLabel: {"app=api", "tier=backend"},
	})
	if err != nil || len(filtered) != 1 || filtered[0].ID != "one" {
		t.Fatalf("filtered containers = %#v, err=%v", filtered, err)
	}
}

func TestPostFilterNetworksUsesSameFieldAND(t *testing.T) {
	items := []runtimeapi.Network{{ID: "one", Name: "team-backend"}, {ID: "two", Name: "team-frontend"}}
	filtered, err := postFilterNetworks(items, runtimeapi.FilterSet{
		runtimeapi.NetworkFilterName: {"team", "backend"},
	})
	if err != nil || len(filtered) != 1 || filtered[0].ID != "one" {
		t.Fatalf("filtered networks = %#v, err=%v", filtered, err)
	}
}

func TestPostFilterRejectsUnprovableImageAND(t *testing.T) {
	_, err := postFilterImages([]runtimeapi.ImageSummary{{ID: "image"}}, runtimeapi.FilterSet{
		runtimeapi.ImageFilterBefore: {"first", "second"},
	})
	if !runtimeapi.IsErrorKind(err, runtimeapi.ErrorUnsupported) {
		t.Fatalf("error = %v, want unsupported", err)
	}
}
