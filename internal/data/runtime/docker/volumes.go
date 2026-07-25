package docker

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// ListVolumesContext returns all volumes visible to the client with a caller-provided context.
func (c *Client) ListVolumesContext(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
	nativeFilters, err := options.NativeFilters()
	if err != nil {
		return nil, err
	}
	filterArgs := filters.NewArgs()
	for field, values := range nativeFilters {
		for _, value := range values {
			filterArgs.Add(field, value)
		}
	}
	resp, err := c.cli.VolumeList(ctx, volume.ListOptions{Filters: filterArgs})
	if err != nil {
		return nil, fmt.Errorf("list volumes: %w", err)
	}
	items := make([]runtimeapi.Volume, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		items = append(items, runtimeapi.Volume{
			Name: v.Name, Driver: v.Driver, Mountpoint: v.Mountpoint,
			Labels: v.Labels, Scope: v.Scope, CreatedAt: v.CreatedAt,
		})
	}
	return items, nil
}

// InspectVolumeContext returns detailed volume info with a caller-provided context.
func (c *Client) InspectVolumeContext(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	_, raw, err := c.cli.VolumeInspectWithRaw(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("inspect volume %s: %w", name, err)
	}
	var native volume.Volume
	if err := sonic.Unmarshal(raw, &native); err != nil {
		return nil, fmt.Errorf("parse volume inspect %s: %w", name, err)
	}
	return &runtimeapi.VolumeDetail{
		Name:       native.Name,
		Driver:     native.Driver,
		Mountpoint: native.Mountpoint,
		CreatedAt:  native.CreatedAt,
		Labels:     native.Labels,
		Scope:      native.Scope,
		Options:    native.Options,
		Status:     native.Status,
	}, nil
}

// CreateVolumeContext creates a new volume with a caller-provided context.
func (c *Client) CreateVolumeContext(ctx context.Context, options runtimeapi.VolumeCreateOptions) (*runtimeapi.Volume, error) {
	if options.Name == "" {
		return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, "volume.create", "", fmt.Errorf("name is required"))
	}
	created, err := c.cli.VolumeCreate(ctx, volume.CreateOptions{Name: options.Name, Driver: options.Driver, Labels: options.Labels, DriverOpts: options.Options})
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, "volume.create", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume, ID: options.Name}, runtimeapi.Docker)
	}
	return &runtimeapi.Volume{Name: created.Name, Driver: created.Driver, Mountpoint: created.Mountpoint, Labels: created.Labels, Scope: created.Scope, CreatedAt: created.CreatedAt}, nil
}

// PruneVolumesContext removes unused volumes with a caller-provided context.
func (c *Client) PruneVolumesContext(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	filterArgs := filters.NewArgs()
	for field, values := range options.Filters {
		for _, value := range values {
			filterArgs.Add(field, value)
		}
	}
	report, err := c.cli.VolumesPrune(ctx, filterArgs)
	if err != nil {
		return runtimeapi.PruneResult{}, runtimeapi.MapRuntimeError(err, "volume.prune", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume}, runtimeapi.Docker)
	}
	result := runtimeapi.PruneResult{SpaceReclaimed: report.SpaceReclaimed, Resources: make([]runtimeapi.ResourceResult, 0, len(report.VolumesDeleted))}
	for _, name := range report.VolumesDeleted {
		result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: name})
	}
	return result, nil
}
