package podman

import runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"

// MapVolumes converts Podman list results into runtime.Volume items.
// Field differences: CreatedAt (time.Time) is formatted as RFC3339;
// Scope defaults to "local".
func MapVolumes(raw []VolumeItem) []runtimeapi.Volume {
	result := make([]runtimeapi.Volume, 0, len(raw))
	for _, v := range raw {
		scope := v.Scope
		if scope == "" {
			scope = "local"
		}
		result = append(result, runtimeapi.Volume{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			Labels:     v.Labels,
			Scope:      scope,
			CreatedAt:  FormatPodmanTime(v.CreatedAt),
		})
	}
	return result
}

// MapVolumeInspect converts a single Podman volume inspect result
// into the structured VolumeDetail domain type for the detail view.
func MapVolumeInspect(raw VolumeItem) *runtimeapi.VolumeDetail {
	scope := raw.Scope
	if scope == "" {
		scope = "local"
	}
	return &runtimeapi.VolumeDetail{
		Name:       raw.Name,
		Driver:     raw.Driver,
		Mountpoint: raw.Mountpoint,
		CreatedAt:  FormatPodmanTime(raw.CreatedAt),
		Labels:     raw.Labels,
		Scope:      scope,
		Options:    raw.Options,
		Status:     raw.Status,
	}
}
