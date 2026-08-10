package podman

import (
	"testing"
	"time"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

func TestMapVolumesFormatsCreatedAt(t *testing.T) {
	created := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	result := MapVolumes([]dto.VolumeItem{{
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
	if !vol.CreatedAt.Equal(time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)) {
		t.Errorf("CreatedAt = %v", vol.CreatedAt)
	}
	if vol.Labels["app"] != "api" {
		t.Error("labels not preserved")
	}
}

func TestMapVolumeInspect(t *testing.T) {
	created := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	result := MapVolumeInspect(dto.VolumeItem{
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

func TestMapVolumesEmpty(t *testing.T) {
	result := MapVolumes([]dto.VolumeItem{})
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}

func TestMapVolumeDefaultsScopeAndOmitsZeroTime(t *testing.T) {
	result := MapVolumes([]dto.VolumeItem{{Name: "cache"}})
	if result[0].Scope != "local" {
		t.Fatalf("scope = %q", result[0].Scope)
	}
	if !result[0].CreatedAt.IsZero() {
		t.Fatalf("CreatedAt = %v", result[0].CreatedAt)
	}
}
