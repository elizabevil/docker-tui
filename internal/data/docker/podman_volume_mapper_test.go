package docker

import (
	"testing"
	"time"
)

func TestMapPodmanVolumesFormatsCreatedAt(t *testing.T) {
	created := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	result := mapPodmanVolumes([]podmanVolumeConfigResponse{{
		Name:       "test-vol",
		Driver:     "local",
		Mountpoint: "/var/lib/containers/storage/volumes/test-vol/_data",
		CreatedAt:  created,
		Labels:     map[string]string{"app": "api"},
		Scope:      "local",
		Options:    map[string]string{"type": "none"},
	}})
	if len(result) != 1 {
		t.Fatalf("expected one volume, got %d", len(result))
	}
	vol := result[0]
	if vol.Name != "test-vol" || vol.Driver != "local" {
		t.Errorf("name/driver = %q/%q", vol.Name, vol.Driver)
	}
	if vol.CreatedAt != "2025-06-15T10:30:00Z" {
		t.Errorf("CreatedAt = %q", vol.CreatedAt)
	}
	if vol.Labels["app"] != "api" {
		t.Error("labels not preserved")
	}
}

func TestMapPodmanVolumeInspect(t *testing.T) {
	created := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	result := mapPodmanVolumeInspect(podmanVolumeConfigResponse{
		Name:       "inspect-vol",
		Driver:     "overlay",
		Mountpoint: "/data",
		CreatedAt:  created,
		Scope:      "local",
	})
	if result.Name != "inspect-vol" || result.Driver != "overlay" {
		t.Errorf("name/driver = %q/%q", result.Name, result.Driver)
	}
}

func TestMapPodmanVolumesEmpty(t *testing.T) {
	result := mapPodmanVolumes([]podmanVolumeConfigResponse{})
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}
