package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// ListNetworks returns all Docker networks visible to the client.
func (c *Client) ListNetworks() ([]NetworkItem, error) {
	return c.ListNetworksContext(c.ctx, runtimeapi.NetworkListOptions{})
}

func (c *Client) ListNetworksContext(ctx context.Context, options runtimeapi.NetworkListOptions) ([]NetworkItem, error) {
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
	nets, err := c.cli.NetworkList(ctx, network.ListOptions{Filters: filterArgs})
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
