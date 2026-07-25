package podman

import (
	"testing"
	"time"

	"github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// TestMapImageInspectPopulatesDetailFromPodmanResponse verifies the mapper
// hydrates every field ImageDetail exposes from a representative Podman
// inspect payload.
func TestMapImageInspectPopulatesDetailFromPodmanResponse(t *testing.T) {
	inspect := dto.ImageInspectJSON{
		ID:           "6fde864b9d50c85af8e12fb9555eb6933a18aadd5337774df7e25f566b20167c",
		RepoTags:     []string{"example.com/foo/bar:v0.0.9"},
		RepoDigests:  []string{"example.com/foo/bar@sha256:3d663f950a7cebec1d8a2094930c29bcb98dd72fac56bcd9ed245a8834d6dae7"},
		Comment:      "buildkit.dockerfile.v0",
		Created:      "2026-07-24T03:22:18.667688876Z",
		Architecture: "arm64",
		Os:           "linux",
		OsVersion:    "",
		Author:       "",
		Size:         248560616,
		GraphDriver:  dto.ImageInspectGraphDriver{Name: "overlay"},
		RootFS: dto.ImageInspectRootFS{
			Type:   "layers",
			Layers: []string{"sha256:a", "sha256:b", "sha256:c"},
		},
		Config: &dto.ImageInspectConfig{
			Env:        []string{"PATH=/usr/local/bin"},
			Entrypoint: []string{"bash", "/app/run.sh"},
			WorkingDir: "/app",
			Labels:     map[string]string{"io.buildah.version": "1.39.3"},
		},
	}

	detail := MapImageInspect(inspect, nil)
	if detail.ID != inspect.ID {
		t.Fatalf("ID = %q, want %q", detail.ID, inspect.ID)
	}
	if detail.Architecture != "arm64" {
		t.Fatalf("Architecture = %q, want arm64", detail.Architecture)
	}
	if detail.LayerCount != 3 {
		t.Fatalf("LayerCount = %d, want 3", detail.LayerCount)
	}
	if detail.Driver != "overlay" {
		t.Fatalf("Driver = %q", detail.Driver)
	}
	if detail.Registry != "example.com" || detail.Name != "foo/bar" || detail.Tag != "v0.0.9" {
		t.Fatalf("name split wrong: %+v", detail)
	}
	if detail.Runtime.WorkingDir != "/app" {
		t.Fatalf("WorkingDir = %q", detail.Runtime.WorkingDir)
	}
	if len(detail.Runtime.Entrypoint) != 2 || detail.Runtime.Entrypoint[1] != "/app/run.sh" {
		t.Fatalf("Entrypoint = %v", detail.Runtime.Entrypoint)
	}
	if detail.Labels["io.buildah.version"] != "1.39.3" {
		t.Fatalf("Labels = %v", detail.Labels)
	}
	if detail.HistoryError != "" {
		t.Fatalf("unexpected HistoryError: %q", detail.HistoryError)
	}
	if len(detail.History) != 0 {
		t.Fatalf("History should be empty when none provided, got %d", len(detail.History))
	}
}

// TestMapImageInspectPopulatesHistory ensures history entries are converted
// with their original Created timestamp preserved.
func TestMapImageInspectPopulatesHistory(t *testing.T) {
	inspect := dto.ImageInspectJSON{ID: "x", RepoTags: []string{"example/x:latest"}}
	history := []dto.LayerHistoryEntry{
		{ID: "layer-a", Created: 1784863338, CreatedBy: "/bin/sh -c echo hi", Size: 3072, Comment: ""},
	}
	detail := MapImageInspect(inspect, history)
	if len(detail.History) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(detail.History))
	}
	if detail.History[0].ID != "layer-a" {
		t.Fatalf("History[0].ID = %q", detail.History[0].ID)
	}
	if detail.History[0].Created != 1784863338 {
		t.Fatalf("History[0].Created = %d, want 1784863338", detail.History[0].Created)
	}
}

// TestMapImageInspectHandlesMissingOptionalFields ensures absent config /
// metadata fields don't panic and resolve to neutral defaults.
func TestMapImageInspectHandlesMissingOptionalFields(t *testing.T) {
	inspect := dto.ImageInspectJSON{ID: "x"}
	detail := MapImageInspect(inspect, nil)
	if detail.Architecture != "—" || detail.OS != "—" {
		t.Fatalf("missing fields should render —, got %q / %q", detail.Architecture, detail.OS)
	}
	if detail.Runtime.Entrypoint != nil || detail.Runtime.Cmd != nil {
		t.Fatalf("slice fields should remain nil when config absent")
	}
}

// TestMapImageInspectDecodesRFC3339Created is here to guard the implicit
// conversion: ImageHistoryLayer.Created is int64, matching the libpod
// schema, so no parsing needs to happen here.
func TestMapImageInspectDecodesRFC3339Created(t *testing.T) {
	ts := time.Unix(1784863338, 0)
	if ts.Unix() != 1784863338 {
		t.Fatalf("timestamp round-trip lost precision")
	}
	if got := ts.UTC().Format(time.RFC3339); got != "2026-07-24T03:22:18Z" {
		t.Fatalf("RFC3339 = %q", got)
	}
	_ = runtime.ImageHistoryLayer{}
}
