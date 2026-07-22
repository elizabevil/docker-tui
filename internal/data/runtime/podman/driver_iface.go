package podman

import (
	"context"

	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

// DriverBackend is the optional CGO-backed transport that a *Client may
// carry alongside its REST client. When DriverBackend is non-nil, the
// methods on *Client prefer it over the REST transport; otherwise they
// fall through to REST.
//
// The interface uses Podman-specific types (dto/local) so that the driver
// layer has zero dependency on the parent runtime package.
type DriverBackend interface {
	ListContainers(ctx context.Context, all bool, limit int, filters map[string][]string) ([]ContainerItem, error)
	ListImages(ctx context.Context, all bool, filters map[string][]string) ([]ImageItem, error)

	ListNetworks(ctx context.Context, filters map[string][]string) ([]Network, error)
	InspectNetwork(ctx context.Context, id string) (*NetworkInspectItem, error)
	CreateNetwork(ctx context.Context, name, driver string, internal, ipv6 bool, labels, options map[string]string) (*Network, error)
	RemoveNetwork(ctx context.Context, id string) error
	PruneNetworks(ctx context.Context, filters map[string][]string) ([]dto.NetworkPruneReportItem, error)

	ListVolumes(ctx context.Context, filters map[string][]string) ([]VolumeItem, error)
	InspectVolume(ctx context.Context, name string) (*VolumeItem, error)
	CreateVolume(ctx context.Context, name, driver string, labels, options map[string]string) (*VolumeItem, error)
	RemoveVolume(ctx context.Context, name string, force bool) error
	PruneVolumes(ctx context.Context, filters map[string][]string) ([]dto.VolumePruneReportItem, error)

	Close() error
}
