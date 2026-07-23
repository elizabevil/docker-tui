package podman

import (
	"context"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// DriverBackend is the optional CGO-backed transport that a *Client may
// carry alongside its REST client. When DriverBackend is non-nil, the
// methods on *Client prefer it over the REST transport; otherwise they
// fall through to REST.
//
// All types are dto.* so that CGO and non-CGO builds share the same wire
// types. The CGO driver converts SDK types at the boundary. Parameter
// structs follow Podman binding conventions (not flattened args).
type DriverBackend interface {
	ListContainers(ctx context.Context, opts dto.ContainerListOptions) ([]dto.ContainerItem, error)
	ListImages(ctx context.Context, opts dto.ImageListOptions) ([]dto.ImageItem, error)

	ListNetworks(ctx context.Context, opts dto.NetworkListOptions) ([]dto.Network, error)
	InspectNetwork(ctx context.Context, id string) (*dto.NetworkInspect, error)
	CreateNetwork(ctx context.Context, opts dto.Network) (*dto.Network, error)
	RemoveNetwork(ctx context.Context, id string) error
	PruneNetworks(ctx context.Context, filters map[string][]string) ([]dto.NetworkPruneReportItem, error)

	ListVolumes(ctx context.Context, opts dto.VolumeListOptions) ([]dto.VolumeItem, error)
	InspectVolume(ctx context.Context, name string) (*dto.VolumeItem, error)
	CreateVolume(ctx context.Context, opts dto.VolumeItem) (*dto.VolumeItem, error)
	RemoveVolume(ctx context.Context, name string, force bool) error
	PruneVolumes(ctx context.Context, filters map[string][]string) ([]dto.VolumePruneReportItem, error)

	Close() error
}
