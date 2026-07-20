package detail

import (
	"testing"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
)

func TestBuildImageDetailDataSectionsUsesHistorySemantics(t *testing.T) {
	normal := buildImageDetailDataSections(&dockerclient.ImageDetailData{
		ID:      "sha256:normal",
		History: []dockerclient.ImageHistoryLayer{{CreatedBy: "RUN echo test", Size: 10}},
	})
	if got := normal[len(normal)-1]; got.Title != "History" || got.Subtitle != "Layers" {
		t.Fatalf("normal history = %#v", got)
	}

	manifest := buildImageDetailDataSections(&dockerclient.ImageDetailData{
		ID: "sha256:index", IsManifest: true,
		ManifestVariants: []dockerclient.ImageManifestEntry{{
			Digest: "sha256:variant", Platform: dockerclient.ManifestPlatform{OS: "linux", Architecture: "amd64"},
		}},
	})
	if got := manifest[len(manifest)-1]; got.Title != "History" || got.Subtitle != "Manifest variants" {
		t.Fatalf("manifest history = %#v", got)
	}
}

func TestBuildImageDetailDataSectionsOmitsEmptyHistory(t *testing.T) {
	sections := buildImageDetailDataSections(&dockerclient.ImageDetailData{ID: "sha256:empty"})
	for _, section := range sections {
		if section.Title == "History" {
			t.Fatal("empty history should not create a placeholder section")
		}
	}
}
