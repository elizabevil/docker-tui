//go:build !cgo

package podman

import (
	"time"

	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

// ImageItem is an alias for dto.ImageItem. The non-CGO path uses the
// independent REST DTO; the CGO path aliases the Podman native type.
// JSON tag sets are identical so mapper_*.go works in both modes.
type ImageItem = dto.ImageItem

// VolumeItem is an alias for dto.VolumeItem.
type VolumeItem = dto.VolumeItem

// VolumePruneReport is an alias for dto.VolumePruneReport.
type VolumePruneReport = dto.VolumePruneReport

// NetworkPruneReport is an alias for dto.NetworkPruneReport.
type NetworkPruneReport = dto.NetworkPruneReport

// EventItem is an alias for dto.EventItem.
type EventItem = dto.EventItem

// FormatPodmanTime formats a Podman time value as RFC3339. Delegates to
// the dto helper so the non-CGO path stays self-contained.
func FormatPodmanTime(value time.Time) string {
	return dto.FormatPodmanTime(value)
}
