package docker

import (
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// podmanVolumeConfigResponse is the adapter's stable representation of the
// Libpod volume inspect/list response. Both CGO bindings and non-CGO REST
// normalize into this type before mapping to runtime.Volume or
// runtime.VolumeDetail. The DTO exists because Podman's native type
// (define.InspectVolumeData) carries fields (UID, GID, LockNumber,
// StorageID, Anonymous, NeedsCopyUp, NeedsChown) that have no Docker
// equivalent and would leak runtime-specific details into the domain model.
type podmanVolumeConfigResponse struct {
	Name       string            `json:"Name"`
	Driver     string            `json:"Driver"`
	Mountpoint string            `json:"Mountpoint"`
	CreatedAt  time.Time         `json:"CreatedAt"`
	Labels     map[string]string `json:"Labels"`
	Scope      string            `json:"Scope"`
	Options    map[string]string `json:"Options"`
	Status     map[string]any    `json:"Status,omitempty"`
}

// mapPodmanVolumes converts Podman list results into runtime.Volume items.
// Field differences: CreatedAt (time.Time) is formatted as RFC3339;
// Scope is absent from Podman responses so it defaults to "local".
func mapPodmanVolumes(raw []podmanVolumeConfigResponse) []runtimeapi.Volume {
	result := make([]runtimeapi.Volume, 0, len(raw))
	for _, v := range raw {
		result = append(result, runtimeapi.Volume{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			Labels:     v.Labels,
			Scope:      v.Scope,
			CreatedAt:  v.CreatedAt.Format(time.RFC3339),
		})
	}
	return result
}

// mapPodmanVolumeInspect converts a single Podman volume inspect result
// into the structured VolumeDetail domain type for the detail view.
func mapPodmanVolumeInspect(raw podmanVolumeConfigResponse) *runtimeapi.VolumeDetail {
	return &runtimeapi.VolumeDetail{
		Name:       raw.Name,
		Driver:     raw.Driver,
		Mountpoint: raw.Mountpoint,
		CreatedAt:  raw.CreatedAt.Format(time.RFC3339),
		Labels:     raw.Labels,
		Scope:      raw.Scope,
		Options:    raw.Options,
		Status:     raw.Status,
	}
}
