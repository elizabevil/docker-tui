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

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
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
func (d *cgoDriver) ListContainers(ctx context.Context, opts dto.ContainerListOptions) ([]dto.ContainerItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	listOpts := new(containers.ListOptions).WithAll(opts.All).WithFilters(opts.Filters)
	if opts.Limit > 0 {
		listOpts.WithLast(opts.Limit)
	}
	list, err := containers.List(conn, listOpts)
	if err != nil {
		return nil, fmt.Errorf("podman list containers: %w", err)
	}
	raw := make([]dto.ContainerItem, 0, len(list))
	for _, container := range list {
		ports := make([]dto.ContainerPort, 0, len(container.Ports))
		for _, port := range container.Ports {
			ports = append(ports, dto.ContainerPort{
				ContainerPort: port.ContainerPort, HostPort: port.HostPort, Range: port.Range,
				Protocol: port.Protocol, HostIP: port.HostIP,
			})
		}
		raw = append(raw, dto.ContainerItem{
			ID: container.ID, Names: container.Names, Image: container.Image,
			Status: container.Status, State: container.State, Created: container.Created,
			Ports: ports, Mounts: container.Mounts, Networks: container.Networks, Labels: container.Labels,
		})
	}
	return raw, nil
}

// ListImages returns images via the Podman CGO bindings.
func (d *cgoDriver) ListImages(ctx context.Context, opts dto.ImageListOptions) ([]dto.ImageItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	listOpts := new(images.ListOptions).WithAll(opts.All).WithFilters(opts.Filters)
	list, err := images.List(conn, listOpts)
	if err != nil {
		return nil, fmt.Errorf("podman list images: %w", err)
	}
	raw := make([]dto.ImageItem, 0, len(list))
	for _, p := range list {
		raw = append(raw, dto.ImageItem{
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
func (d *cgoDriver) ListNetworks(ctx context.Context, opts dto.NetworkListOptions) ([]dto.Network, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	listOpts := new(network.ListOptions).WithFilters(opts.Filters)
	list, err := network.List(conn, listOpts)
	if err != nil {
		return nil, fmt.Errorf("podman list networks: %w", err)
	}
	raw := make([]dto.Network, 0, len(list))
	for _, n := range list {
		var item dto.Network
		b, _ := json.Marshal(n)
		json.Unmarshal(b, &item)
		raw = append(raw, item)
	}
	return raw, nil
}

// InspectNetwork returns structured network detail via the Podman CGO bindings.
func (d *cgoDriver) InspectNetwork(ctx context.Context, id string) (*dto.NetworkInspect, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	report, err := network.Inspect(conn, id, nil)
	if err != nil {
		return nil, fmt.Errorf("podman inspect network %s: %w", id, err)
	}
	var net dto.NetworkInspect
	b, _ := json.Marshal(report)
	json.Unmarshal(b, &net)
	return &net, nil
}

// CreateNetwork creates a network via the Podman CGO bindings.
func (d *cgoDriver) CreateNetwork(ctx context.Context, opts dto.Network) (*dto.Network, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	created, err := network.Create(conn, &networktypes.Network{
		Name:        opts.Name,
		Driver:      opts.Driver,
		Internal:    opts.Internal,
		IPv6Enabled: opts.IPv6Enabled,
		Labels:      opts.Labels,
		Options:     opts.Options,
	})
	if err != nil {
		return nil, fmt.Errorf("podman create network: %w", err)
	}
	net := dto.Network{
		Name:              created.Name,
		ID:                created.ID,
		Driver:            created.Driver,
		NetworkInterface:  created.NetworkInterface,
		Created:           created.Created,
		IPv6Enabled:       created.IPv6Enabled,
		Internal:          created.Internal,
		DNSEnabled:        created.DNSEnabled,
		NetworkDNSServers: created.NetworkDNSServers,
		Labels:            created.Labels,
		Options:           created.Options,
		IPAMOptions:       created.IPAMOptions,
	}
	net.Subnets = make([]dto.Subnet, len(created.Subnets))
	for i, s := range created.Subnets {
		sub := dto.Subnet{Subnet: dto.IPNet{IPNet: s.Subnet.IPNet}, Gateway: s.Gateway}
		if s.LeaseRange != nil {
			sub.LeaseRange = &dto.LeaseRange{StartIP: s.LeaseRange.StartIP, EndIP: s.LeaseRange.EndIP}
		}
		net.Subnets[i] = sub
	}
	net.Routes = make([]dto.Route, len(created.Routes))
	for i, r := range created.Routes {
		net.Routes[i] = dto.Route{
			Destination: dto.IPNet{IPNet: r.Destination.IPNet},
			Gateway:     r.Gateway,
			Metric:      r.Metric,
		}
	}
	return &net, nil
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
		result = append(result, dto.NetworkPruneReportItem{Name: report.Name, Error: string(errBytes)})
	}
	return result, nil
}

// ListVolumes returns volumes via the Podman CGO bindings.
func (d *cgoDriver) ListVolumes(ctx context.Context, opts dto.VolumeListOptions) ([]dto.VolumeItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	listOpts := new(volumes.ListOptions).WithFilters(opts.Filters)
	list, err := volumes.List(conn, listOpts)
	if err != nil {
		return nil, fmt.Errorf("podman list volumes: %w", err)
	}
	raw := make([]dto.VolumeItem, 0, len(list))
	for _, v := range list {
		raw = append(raw, dto.VolumeItem{
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
func (d *cgoDriver) InspectVolume(ctx context.Context, name string) (*dto.VolumeItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	vol, err := volumes.Inspect(conn, name, nil)
	if err != nil {
		return nil, fmt.Errorf("podman inspect volume %s: %w", name, err)
	}
	return &dto.VolumeItem{
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
func (d *cgoDriver) CreateVolume(ctx context.Context, opts dto.VolumeItem) (*dto.VolumeItem, error) {
	conn, err := d.connect(ctx)
	if err != nil {
		return nil, err
	}
	created, err := volumes.Create(conn, entitytypes.VolumeCreateOptions{
		Name:    opts.Name,
		Driver:  opts.Driver,
		Labels:  opts.Labels,
		Options: opts.Options,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("podman create volume: %w", err)
	}
	return &dto.VolumeItem{
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
		errStr := ""
		if report.Err != nil {
			errStr = report.Err.Error()
		}
		result = append(result, dto.VolumePruneReportItem{ID: report.Id, Size: int64(report.Size), Err: errStr})
	}
	return result, nil
}
