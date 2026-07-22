package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// ListVolumes returns all Docker volumes visible to the client.
func (c *Client) ListVolumes() ([]runtimeapi.Volume, error) {
	return c.ListVolumesContext(c.ctx, runtimeapi.VolumeListOptions{})
}

// ListVolumesContext returns all volumes visible to the client with a caller-provided context.
func (c *Client) ListVolumesContext(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
	if c.RuntimeType == RuntimePodman {
		return c.listVolumesPodman(ctx, options)
	}
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

// InspectVolume returns structured volume detail from the runtime.
func (c *Client) InspectVolume(name string) (*runtimeapi.VolumeDetail, error) {
	return c.InspectVolumeContext(c.ctx, name)
}

// InspectVolumeContext returns detailed volume info with a caller-provided context.
func (c *Client) InspectVolumeContext(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	if c.RuntimeType == RuntimePodman {
		return c.inspectVolumePodman(ctx, name)
	}
	_, raw, err := c.cli.VolumeInspectWithRaw(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("inspect volume %s: %w", name, err)
	}
	var detail runtimeapi.VolumeDetail
	if err := sonicUnmarshal(raw, &detail); err != nil {
		return nil, fmt.Errorf("parse volume inspect %s: %w", name, err)
	}
	return &detail, nil
}

// RemoveVolume removes a Docker volume by name.
func (c *Client) RemoveVolume(id string, force bool) error {
	if c.RuntimeType == RuntimePodman {
		return c.removeVolumePodman(c.ctx, id, force)
	}
	return c.cli.VolumeRemove(c.ctx, id, force)
}

// CreateVolumeContext creates a new volume with a caller-provided context.
func (c *Client) CreateVolumeContext(ctx context.Context, options runtimeapi.VolumeCreateOptions) (*runtimeapi.Volume, error) {
	if options.Name == "" {
		return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, "volume.create", "", fmt.Errorf("name is required"))
	}
	if c.RuntimeType == RuntimePodman {
		created, err := c.createVolumePodman(ctx, options)
		if err != nil {
			return nil, mapRuntimeError(err, "volume.create", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume, ID: options.Name}, c.RuntimeType)
		}
		return created, nil
	}
	created, err := c.cli.VolumeCreate(ctx, volume.CreateOptions{Name: options.Name, Driver: options.Driver, Labels: options.Labels, DriverOpts: options.Options})
	if err != nil {
		return nil, mapRuntimeError(err, "volume.create", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume, ID: options.Name}, c.RuntimeType)
	}
	return &runtimeapi.Volume{Name: created.Name, Driver: created.Driver, Mountpoint: created.Mountpoint, Labels: created.Labels, Scope: created.Scope, CreatedAt: created.CreatedAt}, nil
}

// PruneVolumesContext removes unused volumes with a caller-provided context.
func (c *Client) PruneVolumesContext(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	if c.RuntimeType == RuntimePodman {
		result, err := c.pruneVolumesPodman(ctx, options)
		if err != nil {
			return runtimeapi.PruneResult{}, mapRuntimeError(err, "volume.prune", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume}, c.RuntimeType)
		}
		for index := range result.Resources {
			result.Resources[index].Error = mapRuntimeError(result.Resources[index].Error, "volume.prune", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume, ID: result.Resources[index].ID}, c.RuntimeType)
		}
		return result, nil
	}
	filterArgs := filters.NewArgs()
	for field, values := range options.Filters {
		for _, value := range values {
			filterArgs.Add(field, value)
		}
	}
	report, err := c.cli.VolumesPrune(ctx, filterArgs)
	if err != nil {
		return runtimeapi.PruneResult{}, mapRuntimeError(err, "volume.prune", runtimeapi.ResourceRef{Type: runtimeapi.ResourceVolume}, c.RuntimeType)
	}
	result := runtimeapi.PruneResult{SpaceReclaimed: report.SpaceReclaimed, Resources: make([]runtimeapi.ResourceResult, 0, len(report.VolumesDeleted))}
	for _, name := range report.VolumesDeleted {
		result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: name})
	}
	return result, nil
}
