package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/network"
)

// ListNetworks returns all Docker networks visible to the client.
func (c *Client) ListNetworks() ([]NetworkItem, error) {
	nets, err := c.cli.NetworkList(c.ctx, network.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list networks: %w", err)
	}
	items := make([]NetworkItem, 0, len(nets))
	for _, n := range nets {
		ipam := make([]string, 0)
		for _, pool := range n.IPAM.Config {
			if pool.Subnet != "" {
				ipam = append(ipam, pool.Subnet)
			}
		}
		items = append(items, NetworkItem{
			Name: n.Name, ID: n.ID, Driver: n.Driver,
			Scope: n.Scope, IPAM: ipam,
			Containers: len(n.Containers), Created: n.Created.Unix(),
			Internal: n.Internal, Labels: n.Labels,
		})
	}
	return items, nil
}

// InspectNetwork returns the raw JSON from docker network inspect.
func (c *Client) InspectNetwork(id string) ([]byte, error) {
	_, raw, err := c.cli.NetworkInspectWithRaw(context.Background(), id, network.InspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("inspect network %s: %w", id, err)
	}
	return raw, nil
}

// RemoveNetwork removes a Docker network by ID or name.
func (c *Client) RemoveNetwork(id string) error {
	return c.cli.NetworkRemove(c.ctx, id)
}
