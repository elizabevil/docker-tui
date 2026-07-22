//go:build cgo

// This file compiles only when CGO is available. It uses the official Podman
// Go bindings (go.podman.io/podman/v6/pkg/bindings/network) which require CGO
// for libpod client communication. The REST fallback lives in
// podman_networks_nocgo.go. Both files implement the same unexported methods
// on *Client; only one is compiled per build.

package docker

import (
	"context"
	"fmt"
	"net"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"go.podman.io/podman/v6/pkg/bindings"
	"go.podman.io/podman/v6/pkg/bindings/network"
)

func (c *Client) listNetworksPodman(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
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
