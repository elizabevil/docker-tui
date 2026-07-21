package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/volume"
)

// ListVolumes returns all Docker volumes visible to the client.
func (c *Client) ListVolumes() ([]VolumeItem, error) {
	resp, err := c.cli.VolumeList(c.ctx, volume.ListOptions{})
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
