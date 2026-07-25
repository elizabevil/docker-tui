package podman

import (
	"time"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// ContainerItem aliases the canonical dto type, resolved in both CGO and
// non-CGO builds via dto/ build-tag aliases.
type ContainerItem = dto.ContainerItem

// ContainerPort aliases the canonical dto port mapping type.
type ContainerPort = dto.ContainerPort

// ImageItem aliases the canonical dto image list type.
type ImageItem = dto.ImageItem

// VolumeItem aliases the canonical dto volume type.
type VolumeItem = dto.VolumeItem

// VolumePruneReport aliases the canonical dto prune report type.
type VolumePruneReport = dto.VolumePruneReport

// NetworkPruneReport aliases the canonical dto prune report type.
type NetworkPruneReport = dto.NetworkPruneReport

// EventItem aliases the canonical dto event type.
type EventItem = dto.EventItem

// Network aliases the canonical dto network type.
type Network = dto.Network

// NetworkSubnet aliases the canonical dto subnet type.
type NetworkSubnet = dto.Subnet

// NetInterface aliases the canonical dto network interface type.
type NetInterface = dto.NetInterface

// NetAddress aliases the canonical dto network address type.
type NetAddress = dto.NetAddress

// NetworkInspectItem aliases the canonical dto network inspect type.
type NetworkInspectItem = dto.NetworkInspect

// NetworkContainerInfo aliases the canonical dto container info type.
type NetworkContainerInfo = dto.NetworkContainerInfo

// FormatPodmanTime formats a Podman time value as RFC3339. Delegates to
// dto.FormatPodmanTime so the helper is available anywhere in the package.
func FormatPodmanTime(value time.Time) string {
	return dto.FormatPodmanTime(value)
}
