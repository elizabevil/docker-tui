//go:build cgo

package docker

import (
	"context"
	"fmt"

	"go.podman.io/podman/v6/pkg/bindings"
	"go.podman.io/podman/v6/pkg/bindings/containers"
)

func (c *Client) listContainersPodman(ctx context.Context, options ContainerListOptions) ([]ContainerSummary, error) {
	if c.usePodmanRESTTransport() {
		return c.listContainersPodmanREST(ctx, options)
	}
	nativeFilters, err := options.NativeFilters()
	if err != nil {
		return nil, err
	}
	bindingContext, err := bindings.NewConnection(ctx, c.Host)
	if err != nil {
		return nil, fmt.Errorf("podman connect: %w", err)
	}
	listOptions := new(containers.ListOptions).WithAll(options.All).WithFilters(nativeFilters)
	if options.Limit > 0 {
		listOptions.WithLast(options.Limit)
	}
	list, err := containers.List(bindingContext, listOptions)
	if err != nil {
		return nil, fmt.Errorf("podman list containers: %w", err)
	}
	raw := make([]podmanContainerSummary, 0, len(list))
	for _, container := range list {
		ports := make([]podmanPort, 0, len(container.Ports))
		for _, port := range container.Ports {
			ports = append(ports, podmanPort{
				ContainerPort: port.ContainerPort, HostPort: port.HostPort, Range: port.Range,
				Protocol: port.Protocol, HostIP: port.HostIP,
			})
		}
		raw = append(raw, podmanContainerSummary{
			ID: container.ID, Names: container.Names, Image: container.Image,
			Status: container.Status, State: container.State, Created: container.Created,
			Ports: ports, Mounts: container.Mounts, Labels: container.Labels,
		})
	}
	return mapPodmanContainerSummaries(raw), nil
}
