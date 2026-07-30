package detail

import (
	"testing"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func TestBuildImageDetailDataSectionsOmitsHistorySection(t *testing.T) {
	normal := buildImageDetailDataSections(&dockerclient.ImageDetail{
		ID:            "sha256:normal",
		HistorySource: dockerclient.ImageHistoryLayerAPI,
		History:       []dockerclient.ImageHistoryLayer{{CreatedBy: "RUN echo test", Size: 10, Comment: "build step"}},
	})
	for _, sec := range normal {
		if sec.Title == "History" {
			t.Fatalf("history section still rendered: %#v", sec)
		}
	}
}

func TestBuildImageDetailDataSectionsKeepsCoreSections(t *testing.T) {
	sections := buildImageDetailDataSections(&dockerclient.ImageDetail{
		ID:          "sha256:index",
		IsManifest:  true,
		RepoTags:    []string{"example/app:latest"},
		ManifestVariants: []dockerclient.ImageManifestEntry{{
			Digest: "sha256:variant", Platform: dockerclient.ManifestPlatform{OS: "linux", Architecture: "amd64"},
		}},
	})
	if len(sections) == 0 {
		t.Fatalf("sections empty: %#v", sections)
	}
	for _, sec := range sections {
		if sec.Title == "History" {
			t.Fatalf("history section still rendered: %#v", sec)
		}
	}
}
