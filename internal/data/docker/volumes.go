package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// ListVolumes returns all Docker volumes visible to the client.
func (c *Client) ListVolumes() ([]VolumeItem, error) {
	return c.ListVolumesContext(c.ctx, runtimeapi.VolumeListOptions{})
}

func (c *Client) ListVolumesContext(ctx context.Context, options runtimeapi.VolumeListOptions) ([]VolumeItem, error) {
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
	items := make([]VolumeItem, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		items = append(items, VolumeItem{
			Name: v.Name, Driver: v.Driver, Mountpoint: v.Mountpoint,
			Labels: v.Labels, Scope: v.Scope, CreatedAt: v.CreatedAt,
		})
	}
	return items, nil
}

// InspectVolume returns the raw JSON from docker volume inspect.
func (c *Client) InspectVolume(name string) ([]byte, error) {
	_, raw, err := c.cli.VolumeInspectWithRaw(context.Background(), name)
	if err != nil {
		return nil, fmt.Errorf("inspect volume %s: %w", name, err)
	}
	return raw, nil
}

// RemoveVolume removes a Docker volume by name.
func (c *Client) RemoveVolume(id string, force bool) error {
	return c.cli.VolumeRemove(c.ctx, id, force)
}
