package docker

import (
	"fmt"

	"github.com/docker/docker/api/types/network"
)

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

func (c *Client) RemoveNetwork(id string) error {
	return c.cli.NetworkRemove(c.ctx, id)
}
