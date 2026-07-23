//go:build integration

package docker

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/runtime"
)

// TestListImagesManifest verifies that ListImages properly populates Arch, IsManifest,
// and Manifests fields from the Docker API response.
// This is an integration test that requires a running Podman/Docker socket.
func TestListImagesManifest(t *testing.T) {
	client, err := connectForTest(t)
	if err != nil {
		t.Skip("No container runtime available:", err)
	}

	images, err := client.Images().List(context.Background(), runtime.ImageListOptions{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(images) == 0 {
		t.Fatal("No images found — cannot test")
	}

	t.Logf("Found %d images", len(images))
	for _, img := range images {
		if img.Arch == "" {
			t.Errorf("Image %s has empty Arch", img.ID[:12])
		}

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

// TestImageInspectArch verifies that ImageService.Inspect returns full detail.
func TestImageInspectArch(t *testing.T) {
	client, err := connectForTest(t)
	if err != nil {
		t.Skip("No container runtime available:", err)
	}

	images, err := client.Images().List(context.Background(), runtime.ImageListOptions{})
	if err != nil || len(images) == 0 {
		t.Skip("No images to inspect")
	}

	detail, err := client.Images().Inspect(context.Background(), images[0])
	if err != nil {
		t.Fatalf("Inspect(%s) failed: %v", images[0].ID[:12], err)
	}
	if detail.Architecture == "" {
		t.Error("ImageDetail.Architecture is empty")
	}
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
			client, err := NewClient(runtime.ClientConfig{Host: path})
			if err == nil {
				_, err := client.Images().List(context.Background(), runtime.ImageListOptions{})
				if err == nil {
					return client, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no usable container runtime found")
}
