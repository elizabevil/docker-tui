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
