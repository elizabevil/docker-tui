package docker

import (
	"fmt"

	"github.com/docker/docker/api/types/volume"
)

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

func (c *Client) RemoveVolume(id string, force bool) error {
	return c.cli.VolumeRemove(c.ctx, id, force)
}
