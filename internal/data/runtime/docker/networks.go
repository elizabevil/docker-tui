package docker

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// ListNetworksContext returns all networks visible to the client with a caller-provided context.
func (c *Client) ListNetworksContext(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
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

// InspectNetworkContext returns detailed network info with a caller-provided context.
func (c *Client) InspectNetworkContext(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	_, raw, err := c.cli.NetworkInspectWithRaw(ctx, id, network.InspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("inspect network %s: %w", id, err)
	}
	var native network.Inspect
	if err := sonic.Unmarshal(raw, &native); err != nil {
		return nil, fmt.Errorf("parse network inspect %s: %w", id, err)
	}

	containers := make(map[string]runtimeapi.NetworkEndpoint, len(native.Containers))
	for epID, ep := range native.Containers {
		containers[epID] = runtimeapi.NetworkEndpoint{
			Name:        ep.Name,
			EndpointID:  ep.EndpointID,
			MacAddress:  ep.MacAddress,
			IPv4Address: ep.IPv4Address,
			IPv6Address: ep.IPv6Address,
		}
	}

	ipamConfigs := make([]runtimeapi.NetworkIPAMConfig, 0, len(native.IPAM.Config))
	for _, cfg := range native.IPAM.Config {
		ipamConfigs = append(ipamConfigs, runtimeapi.NetworkIPAMConfig{
			Subnet:  cfg.Subnet,
			Gateway: cfg.Gateway,
			IPRange: cfg.IPRange,
		})
	}

	return &runtimeapi.NetworkDetail{
		Name:       native.Name,
		ID:         native.ID,
		Created:    native.Created.Format("2006-01-02T15:04:05Z"),
		Scope:      native.Scope,
		Driver:     native.Driver,
		EnableIPv4: native.EnableIPv4,
		EnableIPv6: native.EnableIPv6,
		IPAM:       runtimeapi.NetworkIPAM{Config: ipamConfigs},
		Internal:   native.Internal,
		Containers: containers,
		Options:    native.Options,
		Labels:     native.Labels,
	}, nil
}

// CreateNetworkContext creates a new network with a caller-provided context.
func (c *Client) CreateNetworkContext(ctx context.Context, options runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	if options.Name == "" {
		return nil, runtimeapi.NewError(runtimeapi.ErrorInvalid, "network.create", "", fmt.Errorf("name is required"))
	}
	created, err := c.cli.NetworkCreate(ctx, options.Name, network.CreateOptions{Driver: options.Driver, Internal: options.Internal, EnableIPv6: &options.EnableIPv6, Labels: options.Labels, Options: options.Options})
	if err != nil {
		return nil, runtimeapi.MapRuntimeError(err, "network.create", runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork, ID: options.Name}, runtimeapi.Docker)
	}
	return &runtimeapi.Network{Name: options.Name, ID: created.ID, Driver: options.Driver, Internal: options.Internal, Labels: options.Labels}, nil
}

// PruneNetworksContext removes unused networks with a caller-provided context.
func (c *Client) PruneNetworksContext(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	filterArgs := filters.NewArgs()
	for field, values := range options.Filters {
		for _, value := range values {
			filterArgs.Add(field, value)
		}
	}
	report, err := c.cli.NetworksPrune(ctx, filterArgs)
	if err != nil {
		return runtimeapi.PruneResult{}, runtimeapi.MapRuntimeError(err, "network.prune", runtimeapi.ResourceRef{Type: runtimeapi.ResourceNetwork}, runtimeapi.Docker)
	}
	result := runtimeapi.PruneResult{Resources: make([]runtimeapi.ResourceResult, 0, len(report.NetworksDeleted))}
	for _, name := range report.NetworksDeleted {
		result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: name})
	}
	return result, nil
}
