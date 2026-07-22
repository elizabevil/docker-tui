package podman

import "testing"

func TestMapImageSummaries(t *testing.T) {
	result := MapImageSummaries([]ImageItem{{
		ID:             "sha256:123",
		RepoTags:       []string{"quay.io/example/app:latest"},
		Arch:           "arm64",
		IsManifestList: new(true),
	}})

	if len(result) != 1 {
		t.Fatalf("expected one image, got %d", len(result))
	}
	if result[0].Arch != "arm64" || !result[0].IsManifest {
		t.Fatalf("unexpected mapped image: %+v", result[0])
	}
	if result[0].Registry != "quay.io" {
		t.Fatalf("expected quay.io registry, got %q", result[0].Registry)
	}
}

func TestMapImageSummariesUsesUnknownArchitecture(t *testing.T) {
	result := MapImageSummaries([]ImageItem{{ID: "sha256:123"}})
	if result[0].Arch != "\u2014" {
		t.Fatalf("expected unknown architecture marker, got %q", result[0].Arch)
	}
}
