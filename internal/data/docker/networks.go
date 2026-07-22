package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// ListNetworks returns all Docker networks visible to the client.
func (c *Client) ListNetworks() ([]runtimeapi.Network, error) {
	return c.ListNetworksContext(c.ctx, runtimeapi.NetworkListOptions{})
}

func (c *Client) ListNetworksContext(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	if c.RuntimeType == RuntimePodman {
		return c.listNetworksPodman(ctx, options)
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
	nets, err := c.cli.NetworkList(ctx, network.ListOptions{Filters: filterArgs})
	if err != nil {
		return nil, fmt.Errorf("list networks: %w", err)
	}
	items := make([]runtimeapi.Network, 0, len(nets))
	for _, n := range nets {
		ipam := make([]string, 0)
		for _, pool := range n.IPAM.Config {
			if pool.Subnet != "" {
				ipam = append(ipam, pool.Subnet)
			}
		}
		items = append(items, runtimeapi.Network{
			Name: n.Name, ID: n.ID, Driver: n.Driver,
			Scope: n.Scope, IPAM: ipam,
			Containers: len(n.Containers), Created: n.Created.Unix(),
			Internal: n.Internal, Labels: n.Labels,
		})
	}
	return items, nil
}

// InspectNetwork returns structured network detail from the runtime.
func (c *Client) InspectNetwork(id string) (*runtimeapi.NetworkDetail, error) {
	return c.InspectNetworkContext(c.ctx, id)
}

func (c *Client) InspectNetworkContext(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	if c.RuntimeType == RuntimePodman {
		return c.inspectNetworkPodman(ctx, id)
	}
	_, raw, err := c.cli.NetworkInspectWithRaw(ctx, id, network.InspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("inspect network %s: %w", id, err)
	}
	var detail runtimeapi.NetworkDetail
	if err := sonicUnmarshal(raw, &detail); err != nil {
		return nil, fmt.Errorf("parse network inspect %s: %w", id, err)
	}
	return &detail, nil
}

// RemoveNetwork removes a Docker network by ID or name.
func (c *Client) RemoveNetwork(id string) error {
	if c.RuntimeType == RuntimePodman {
		return c.removeNetworkPodman(c.ctx, id)
	}
	return c.cli.NetworkRemove(c.ctx, id)
}

func (c *Client) CreateNetworkContext(ctx context.Context, options runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	if options.Name == "" {
		return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, "network.create", "", fmt.Errorf("name is required"))
	}
	if c.RuntimeType == RuntimePodman {
		created, err := c.createNetworkPodman(ctx, options)
		if err != nil {
			return nil, mapRuntimeError(err, "network.create", runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork, ID: options.Name}, c.RuntimeType)
		}
		return created, nil
	}
	created, err := c.cli.NetworkCreate(ctx, options.Name, network.CreateOptions{Driver: options.Driver, Internal: options.Internal, EnableIPv6: &options.EnableIPv6, Labels: options.Labels, Options: options.Options})
	if err != nil {
		return nil, mapRuntimeError(err, "network.create", runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork, ID: options.Name}, c.RuntimeType)
	}
	return &runtimeapi.Network{Name: options.Name, ID: created.ID, Driver: options.Driver, Internal: options.Internal, Labels: options.Labels}, nil
}

func (c *Client) PruneNetworksContext(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	if c.RuntimeType == RuntimePodman {
		result, err := c.pruneNetworksPodman(ctx, options)
		if err != nil {
			return runtimeapi.PruneResult{}, mapRuntimeError(err, "network.prune", runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork}, c.RuntimeType)
		}
		for index := range result.Resources {
			result.Resources[index].Error = mapRuntimeError(result.Resources[index].Error, "network.prune", runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork, ID: result.Resources[index].ID}, c.RuntimeType)
		}
		return result, nil
	}
	filterArgs := filters.NewArgs()
	for field, values := range options.Filters {
		for _, value := range values {
			filterArgs.Add(field, value)
		}
	}
	report, err := c.cli.NetworksPrune(ctx, filterArgs)
	if err != nil {
		return runtimeapi.PruneResult{}, mapRuntimeError(err, "network.prune", runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork}, c.RuntimeType)
	}
	result := runtimeapi.PruneResult{Resources: make([]runtimeapi.ResourceResult, 0, len(report.NetworksDeleted))}
	for _, name := range report.NetworksDeleted {
		result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: name})
	}
	return result, nil
}
