package detail

import (
	"strings"
	"testing"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func TestBuildImageDetailDataSectionsUsesHistorySemantics(t *testing.T) {
	normal := buildImageDetailDataSections(&dockerclient.ImageDetail{
		ID:            "sha256:normal",
		HistorySource: dockerclient.ImageHistoryLayerAPI,
		History:       []dockerclient.ImageHistoryLayer{{CreatedBy: "RUN echo test", Size: 10, Comment: "build step"}},
	})
	if got := normal[len(normal)-1]; got.Title != "History" || got.Subtitle != "Layers" || !strings.Contains(got.Lines[0], "build step") {
		t.Fatalf("normal history = %#v", got)
	}

	manifest := buildImageDetailDataSections(&dockerclient.ImageDetail{
		ID: "sha256:index", IsManifest: true,
		ManifestVariants: []dockerclient.ImageManifestEntry{{
			Digest: "sha256:variant", Platform: dockerclient.ManifestPlatform{OS: "linux", Architecture: "amd64"},
		}},
	})
	if got := manifest[len(manifest)-1]; got.Title != "History" || got.Subtitle != "Manifest variants" {
		t.Fatalf("manifest history = %#v", got)
	}
}

func TestBuildImageDetailDataSectionsExplainsEmptyHistory(t *testing.T) {
	sections := buildImageDetailDataSections(&dockerclient.ImageDetail{
		ID: "sha256:empty", HistorySource: dockerclient.ImageHistoryLayerAPI,
	})
	history := sections[len(sections)-1]
	if history.Title != "History" || len(history.Lines) != 1 || !strings.Contains(history.Lines[0], "no layer history") {
		t.Fatalf("empty history state = %#v", history)
	}
}

func TestBuildImageDetailDataSectionsDistinguishesPendingAndFailure(t *testing.T) {
	pending := buildImageDetailDataSections(&dockerclient.ImageDetail{ID: "sha256:pending", HistorySource: dockerclient.ImageHistoryPending})
	if got := pending[len(pending)-1].Lines[0]; !strings.Contains(got, "Loading") {
		t.Fatalf("pending history = %q", got)
	}

	failed := buildImageDetailDataSections(&dockerclient.ImageDetail{ID: "sha256:failed", HistorySource: dockerclient.ImageHistoryLayerAPI, HistoryError: "unsupported"})
	if got := failed[len(failed)-1].Lines[0]; !strings.Contains(got, "unavailable") || !strings.Contains(got, "unsupported") {
		t.Fatalf("failed history = %q", got)
	}
}
