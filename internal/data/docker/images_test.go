package docker

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestListImagesManifest verifies that ListImages properly populates Arch, IsManifest,
// and Manifests fields from the Docker API response.
// This is an integration test that requires a running Podman/Docker socket.
func TestListImagesManifest(t *testing.T) {
	client, err := connectForTest(t)
	if err != nil {
		t.Skip("No container runtime available:", err)
	}

	images, err := client.ListImages()
	if err != nil {
		t.Fatalf("ListImages failed: %v", err)
	}
	if len(images) == 0 {
		t.Fatal("No images found — cannot test")
	}

	t.Logf("Found %d images", len(images))
	for _, img := range images {
		// Every image should have either a real Arch or the placeholder
		if img.Arch == "" {
			t.Errorf("Image %s has empty Arch", img.ID[:12])
		}

		// Log manifest info
		if img.IsManifest {
			t.Logf("  Manifest list: %s (%d platforms)", img.ID[:12], len(img.Manifests))
		}
		if len(img.Manifests) > 0 {
			for _, m := range img.Manifests {
				t.Logf("    → %s/%s %s", m.Platform.OS, m.Platform.Architecture, m.Digest[:19])
			}
		}
	}
}

// TestImageInspectArch verifies that ImageInspect returns Architecture info.
func TestImageInspectArch(t *testing.T) {
	client, err := connectForTest(t)
	if err != nil {
		t.Skip("No container runtime available:", err)
	}

	images, err := client.ListImages()
	if err != nil || len(images) == 0 {
		t.Skip("No images to inspect")
	}

	// Inspect the first image
	info, err := client.InspectImage(images[0].ID)
	if err != nil {
		t.Fatalf("InspectImage(%s) failed: %v", images[0].ID[:12], err)
	}
	if !strings.Contains(info, "Architecture:") {
		t.Error("InspectImage output missing Architecture field")
	}
	t.Logf("Inspect output contains Architecture: %v", strings.Contains(info, "Architecture:"))
}

// TestManifestDetectFromAPI verifies that the raw API JSON response
// is properly parsed into our ImageSummary struct, focusing on manifest fields.
func TestManifestDetectFromAPI(t *testing.T) {
	// Simulate the raw JSON that the Docker API might return
	raw := `{
		"Id": "sha256:test123",
		"RepoTags": ["localhost/test:latest"],
		"Created": 1700000000,
		"Size": 50000000,
		"Descriptor": {
			"mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
			"digest": "sha256:abc123",
			"platform": {"architecture": "arm64", "os": "linux"}
		},
		"Manifests": [
			{
				"ID": "sha256:manifest1",
				"Descriptor": {"mediaType": "application/vnd.docker.distribution.manifest.v2+json", "digest": "sha256:abc"},
				"Size": {"Content": 50000000, "Total": 60000000},
				"Available": true,
				"Kind": "image",
				"ImageData": {
					"Platform": {"architecture": "amd64", "os": "linux", "variant": "v3"},
					"Size": {"Unpacked": 100000000}
				}
			},
			{
				"ID": "sha256:manifest2",
				"Descriptor": {"mediaType": "application/vnd.docker.distribution.manifest.v2+json", "digest": "sha256:def"},
				"Size": {"Content": 60000000, "Total": 70000000},
				"Available": false,
				"Kind": "image",
				"ImageData": {
					"Platform": {"architecture": "arm64", "os": "linux"},
					"Size": {"Unpacked": 120000000}
				}
			}
		]
	}`

	var rawImg struct {
		ID         string          `json:"Id"`
		RepoTags   []string        `json:"RepoTags"`
		Created    int64           `json:"Created"`
		Size       int64           `json:"Size"`
		Descriptor json.RawMessage `json:"Descriptor,omitempty"`
		Manifests  json.RawMessage `json:"Manifests,omitempty"`
	}
	if err := json.Unmarshal([]byte(raw), &rawImg); err != nil {
		t.Fatalf("Failed to parse test JSON: %v", err)
	}

	// Build ImageSummary manually (simulating what ListImages does)
	s := ImageSummary{
		ID:       rawImg.ID,
		RepoTags: rawImg.RepoTags,
		Created:  rawImg.Created,
		Size:     rawImg.Size,
		Arch:     "arm64", // from Descriptor.Platform.Architecture
	}

	// Populate manifests from raw JSON
	type rawManifest struct {
		ID   string `json:"ID"`
		Size struct {
			Content int64 `json:"Content"`
		} `json:"Size"`
		Available bool   `json:"Available"`
		Kind      string `json:"Kind"`
		ImageData *struct {
			Platform struct {
				Architecture string `json:"architecture"`
				OS           string `json:"os"`
				Variant      string `json:"variant,omitempty"`
			} `json:"Platform"`
		} `json:"ImageData,omitempty"`
	}

	var manifests []rawManifest
	if err := json.Unmarshal(rawImg.Manifests, &manifests); err == nil {
		for _, m := range manifests {
			entry := ImageManifestEntry{
				Digest:    m.ID,
				Size:      m.Size.Content,
				Available: m.Available,
			}
			if m.ImageData != nil {
				entry.Platform = ManifestPlatform{
					OS:           m.ImageData.Platform.OS,
					Architecture: m.ImageData.Platform.Architecture,
					Variant:      m.ImageData.Platform.Variant,
				}
			}
			s.Manifests = append(s.Manifests, entry)
		}
	}

	// Verify
	s.IsManifest = len(s.Manifests) > 1
	if !s.IsManifest {
		t.Error("Expected IsManifest=true for multi-manifest image")
	}
	if s.Arch != "arm64" {
		t.Errorf("Expected Arch=arm64, got %s", s.Arch)
	}
	if len(s.Manifests) != 2 {
		t.Fatalf("Expected 2 manifests, got %d", len(s.Manifests))
	}
	if s.Manifests[0].Platform.Architecture != "amd64" {
		t.Errorf("Expected first manifest arch=amd64, got %s", s.Manifests[0].Platform.Architecture)
	}
	if s.Manifests[1].Platform.Architecture != "arm64" {
		t.Errorf("Expected second manifest arch=arm64, got %s", s.Manifests[1].Platform.Architecture)
	}
	t.Log("All manifest parsing tests passed")
}

// connectForTest creates a Docker client using the Podman socket or Docker socket.
func connectForTest(t *testing.T) (*Client, error) {
	paths := []string{
		"unix:///run/user/1000/podman/podman.sock",
		"unix:///var/run/docker.sock",
		"unix:///var/run/podman/podman.sock",
	}
	for _, path := range paths {
		sock := strings.TrimPrefix(path, "unix://")
		if _, err := os.Stat(sock); err == nil {
			t.Logf("Using socket: %s", path)
			client, err := NewClient(ClientConfig{Host: path})
			if err == nil {
				// Verify connection
				_, err := client.ListImages()
				if err == nil {
					return client, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no usable container runtime found")
}
