//go:build cgo

package podman

import (
	"context"
	"encoding/json"
	"fmt"

	networktypes "go.podman.io/common/libnetwork/types"
	"go.podman.io/podman/v6/pkg/bindings"
	"go.podman.io/podman/v6/pkg/bindings/containers"
	"go.podman.io/podman/v6/pkg/bindings/images"
	"go.podman.io/podman/v6/pkg/bindings/network"
	"go.podman.io/podman/v6/pkg/bindings/volumes"
	entitytypes "go.podman.io/podman/v6/pkg/domain/entities/types"

	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

// cgoDriver wraps the official Podman Go bindings. It implements
// DriverBackend and is the preferred transport when the CGO build is
// enabled. Connections are established lazily on first use.
type cgoDriver struct {
	host string
}

// newCGODriver constructs a cgoDriver for the given host.
func newCGODriver(host string) *cgoDriver {
	return &cgoDriver{host: host}
}

func (d *cgoDriver) connect(ctx context.Context) (context.Context, error) {
	conn, err := bindings.NewConnection(ctx, d.host)
	if err != nil {
		return nil, fmt.Errorf("podman connect: %w", err)
	}
	return conn, nil
}

// Close releases no resources; bindings.Connection is stateless and reused.
func (d *cgoDriver) Close() error { return nil }

// ListContainers returns containers via the Podman CGO bindings.
func (d *cgoDriver) ListContainers(ctx context.Context, all bool, limit int, filters map[string][]string) ([]ContainerItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	listOpts := new(containers.ListOptions).WithAll(all).WithFilters(filters)
	if limit > 0 {
		listOpts.WithLast(limit)
	}
	list, err := containers.List(conn, listOpts)
	if err != nil {
		return nil, fmt.Errorf("podman list containers: %w", err)
	}
	raw := make([]ContainerItem, 0, len(list))
	for _, container := range list {
		ports := make([]ContainerPort, 0, len(container.Ports))
		for _, port := range container.Ports {
			ports = append(ports, ContainerPort{
				ContainerPort: port.ContainerPort, HostPort: port.HostPort, Range: port.Range,
				Protocol: port.Protocol, HostIP: port.HostIP,
			})
		}
		raw = append(raw, ContainerItem{
			ID: container.ID, Names: container.Names, Image: container.Image,
			Status: container.Status, State: container.State, Created: container.Created,
			Ports: ports, Mounts: container.Mounts, Networks: container.Networks, Labels: container.Labels,
		})
	}
	return raw, nil
}

// ListImages returns images via the Podman CGO bindings.
func (d *cgoDriver) ListImages(ctx context.Context, all bool, filters map[string][]string) ([]ImageItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	listOpts := new(images.ListOptions).WithAll(all).WithFilters(filters)
	list, err := images.List(conn, listOpts)
	if err != nil {
		return nil, fmt.Errorf("podman list images: %w", err)
	}
	raw := make([]ImageItem, 0, len(list))
	for _, p := range list {
		raw = append(raw, ImageItem{
			ID:             p.ID,
			RepoTags:       p.RepoTags,
			Created:        p.Created,
			Size:           p.Size,
			Labels:         p.Labels,
			Arch:           p.Arch,
			IsManifestList: p.IsManifestList,
		})
	}
	return raw, nil
}

// ListNetworks returns networks via the Podman CGO bindings.
func (d *cgoDriver) ListNetworks(ctx context.Context, filters map[string][]string) ([]Network, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	listOpts := new(network.ListOptions).WithFilters(filters)
	list, err := network.List(conn, listOpts)
	if err != nil {
		return nil, fmt.Errorf("podman list networks: %w", err)
	}
	return list, nil
}

// InspectNetwork returns structured network detail via the Podman CGO bindings.
func (d *cgoDriver) InspectNetwork(ctx context.Context, id string) (*NetworkInspectItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	report, err := network.Inspect(conn, id, nil)
	if err != nil {
		return nil, fmt.Errorf("podman inspect network %s: %w", id, err)
	}
	return report, nil
}

// CreateNetwork creates a network via the Podman CGO bindings.
func (d *cgoDriver) CreateNetwork(ctx context.Context, name, driver string, internal, ipv6 bool, labels, options map[string]string) (*Network, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	created, err := network.Create(conn, &networktypes.Network{
		Name:        name,
		Driver:      driver,
		Internal:    internal,
		IPv6Enabled: ipv6,
		Labels:      labels,
		Options:     options,
	})
	if err != nil {
		return nil, fmt.Errorf("podman create network: %w", err)
	}
	return created, nil
}

// RemoveNetwork deletes a network via the Podman CGO bindings.
func (d *cgoDriver) RemoveNetwork(ctx context.Context, id string) error {
	conn, err := d.connect(ctx)
	if err != nil {
		return err
	}
	if _, err = network.Remove(conn, id, nil); err != nil {
		return fmt.Errorf("podman remove network: %w", err)
	}
	return nil
}

// PruneNetworks removes unused networks via the Podman CGO bindings.
func (d *cgoDriver) PruneNetworks(ctx context.Context, filters map[string][]string) ([]dto.NetworkPruneReportItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	reports, err := network.Prune(conn, new(network.PruneOptions).WithFilters(filters))
	if err != nil {
		return nil, fmt.Errorf("podman prune networks: %w", err)
	}
	result := make([]dto.NetworkPruneReportItem, 0, len(reports))
	for _, report := range reports {
		errBytes, _ := json.Marshal(report.Error)
		result = append(result, dto.NetworkPruneReportItem{Name: report.Name, Error: errBytes})
	}
	return result, nil
}

// ListVolumes returns volumes via the Podman CGO bindings.
func (d *cgoDriver) ListVolumes(ctx context.Context, filters map[string][]string) ([]VolumeItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	listOpts := new(volumes.ListOptions).WithFilters(filters)
	list, err := volumes.List(conn, listOpts)
	if err != nil {
		return nil, fmt.Errorf("podman list volumes: %w", err)
	}
	raw := make([]VolumeItem, 0, len(list))
	for _, v := range list {
		raw = append(raw, VolumeItem{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			CreatedAt:  v.CreatedAt,
			Labels:     v.Labels,
			Scope:      v.Scope,
			Options:    v.Options,
			Status:     v.Status,
		})
	}
	return raw, nil
}

// InspectVolume returns structured volume detail via the Podman CGO bindings.
func (d *cgoDriver) InspectVolume(ctx context.Context, name string) (*VolumeItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	vol, err := volumes.Inspect(conn, name, nil)
	if err != nil {
		return nil, fmt.Errorf("podman inspect volume %s: %w", name, err)
	}
	return &VolumeItem{
		Name:       vol.Name,
		Driver:     vol.Driver,
		Mountpoint: vol.Mountpoint,
		CreatedAt:  vol.CreatedAt,
		Labels:     vol.Labels,
		Scope:      vol.Scope,
		Options:    vol.Options,
		Status:     vol.Status,
	}, nil
}

// CreateVolume creates a volume via the Podman CGO bindings.
func (d *cgoDriver) CreateVolume(ctx context.Context, name, driver string, labels, options map[string]string) (*VolumeItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	created, err := volumes.Create(conn, entitytypes.VolumeCreateOptions{
		Name:    name,
		Driver:  driver,
		Labels:  labels,
		Options: options,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("podman create volume: %w", err)
	}
	return &VolumeItem{
		Name:       created.Name,
		Driver:     created.Driver,
		Mountpoint: created.Mountpoint,
		CreatedAt:  created.CreatedAt,
		Labels:     created.Labels,
		Scope:      created.Scope,
	}, nil
}

// RemoveVolume deletes a volume via the Podman CGO bindings.
func (d *cgoDriver) RemoveVolume(ctx context.Context, name string, force bool) error {
	conn, err := d.connect(ctx)
	if err != nil {
		return err
	}
	opts := new(volumes.RemoveOptions).WithForce(force)
	if err := volumes.Remove(conn, name, opts); err != nil {
		return fmt.Errorf("podman remove volume: %w", err)
	}
	return nil
}

// PruneVolumes removes unused volumes via the Podman CGO bindings.
func (d *cgoDriver) PruneVolumes(ctx context.Context, filters map[string][]string) ([]dto.VolumePruneReportItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	reports, err := volumes.Prune(conn, new(volumes.PruneOptions).WithFilters(filters))
	if err != nil {
		return nil, fmt.Errorf("podman prune volumes: %w", err)
	}
	result := make([]dto.VolumePruneReportItem, 0, len(reports))
	for _, report := range reports {
		result = append(result, dto.VolumePruneReportItem{ID: report.Id, Size: report.Size, Err: json.RawMessage(report.Err)})
	}
	return result, nil
}
