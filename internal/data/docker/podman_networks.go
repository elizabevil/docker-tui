//go:build cgo

// This file compiles only when CGO is available. It uses the official Podman
// Go bindings (go.podman.io/podman/v6/pkg/bindings/network) which require CGO
// for local libpod communication. TLS connections and explicit API overrides
// use the shared REST implementation because bindings cannot express all of
// the configured transport semantics.

package docker

import (
	"context"
	"fmt"
	"net"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	networktypes "go.podman.io/common/libnetwork/types"
	"go.podman.io/podman/v6/pkg/bindings"
	"go.podman.io/podman/v6/pkg/bindings/network"
)

func (c *Client) listNetworksPodman(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	if c.usePodmanRESTTransport() {
		return c.listNetworksPodmanREST(ctx, options)
	}
	nativeFilters, err := options.NativeFilters()
	if err != nil {
		return nil, err
	}
	bindingContext, err := bindings.NewConnection(ctx, c.Host)
	if err != nil {
		return nil, fmt.Errorf("podman connect: %w", err)
	}
	listOptions := new(network.ListOptions).WithFilters(nativeFilters)
	list, err := network.List(bindingContext, listOptions)
	if err != nil {
		return nil, fmt.Errorf("podman list networks: %w", err)
	}
	raw := make([]podmanNetworkItem, 0, len(list))
	for _, n := range list {
		subnets := make([]podmanSubnet, 0, len(n.Subnets))
		for _, s := range n.Subnets {
			subnets = append(subnets, podmanSubnet{
				Subnet:  s.Subnet.String(),
				Gateway: gatewayString(s.Gateway),
			})
		}
		raw = append(raw, podmanNetworkItem{
			Name:        n.Name,
			ID:          n.ID,
			Driver:      n.Driver,
			Created:     n.Created,
			Subnets:     subnets,
			IPv6Enabled: n.IPv6Enabled,
			Internal:    n.Internal,
			Labels:      n.Labels,
			Options:     n.Options,
		})
	}
	return mapPodmanNetworks(raw), nil
}

func (c *Client) inspectNetworkPodman(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	if c.usePodmanRESTTransport() {
		return c.inspectNetworkPodmanREST(ctx, id)
	}
	bindingContext, err := bindings.NewConnection(ctx, c.Host)
	if err != nil {
		return nil, fmt.Errorf("podman connect: %w", err)
	}
	report, err := network.Inspect(bindingContext, id, nil)
	if err != nil {
		return nil, fmt.Errorf("podman inspect network %s: %w", id, err)
	}
	subnets := make([]podmanSubnet, 0, len(report.Subnets))
	for _, s := range report.Subnets {
		subnets = append(subnets, podmanSubnet{
			Subnet:  s.Subnet.String(),
			Gateway: gatewayString(s.Gateway),
		})
	}
	item := podmanNetworkItem{
		Name:        report.Name,
		ID:          report.ID,
		Driver:      report.Driver,
		Created:     report.Created,
		Subnets:     subnets,
		IPv6Enabled: report.IPv6Enabled,
		Internal:    report.Internal,
		Labels:      report.Labels,
		Options:     report.Options,
	}
	return mapPodmanNetworkInspect(item), nil
}

func (c *Client) removeNetworkPodman(ctx context.Context, id string) error {
	if c.usePodmanRESTTransport() {
		return c.removeNetworkPodmanREST(ctx, id)
	}
	bindingContext, err := bindings.NewConnection(ctx, c.Host)
	if err != nil {
		return fmt.Errorf("podman connect: %w", err)
	}
	_, err = network.Remove(bindingContext, id, nil)
	return err
}

func gatewayString(gw net.IP) string {
	if gw == nil {
		return ""
	}
	return gw.String()
}

func (c *Client) createNetworkPodman(ctx context.Context, options runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	if c.usePodmanRESTTransport() {
		return c.createNetworkPodmanREST(ctx, options)
	}
	bindingContext, err := bindings.NewConnection(ctx, c.Host)
	if err != nil {
		return nil, fmt.Errorf("podman connect: %w", err)
	}
	created, err := network.Create(bindingContext, &networktypes.Network{Name: options.Name, Driver: options.Driver, Internal: options.Internal, IPv6Enabled: options.EnableIPv6, Labels: options.Labels, Options: options.Options})
	if err != nil {
		return nil, fmt.Errorf("podman create network: %w", err)
	}
	mapped := mapPodmanNetworks([]podmanNetworkItem{{Name: created.Name, ID: created.ID, Driver: created.Driver, Created: created.Created, IPv6Enabled: created.IPv6Enabled, Internal: created.Internal, Labels: created.Labels, Options: created.Options}})
	return &mapped[0], nil
}

func (c *Client) pruneNetworksPodman(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	if c.usePodmanRESTTransport() {
		return c.pruneNetworksPodmanREST(ctx, options)
	}
	bindingContext, err := bindings.NewConnection(ctx, c.Host)
	if err != nil {
		return runtimeapi.PruneResult{}, fmt.Errorf("podman connect: %w", err)
	}
	reports, err := network.Prune(bindingContext, new(network.PruneOptions).WithFilters(map[string][]string(options.Filters)))
	if err != nil {
		return runtimeapi.PruneResult{}, fmt.Errorf("podman prune networks: %w", err)
	}
	result := runtimeapi.PruneResult{Resources: make([]runtimeapi.ResourceResult, 0, len(reports))}
	for _, report := range reports {
		result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: report.Name, Error: report.Error})
	}
	return result, nil
}
