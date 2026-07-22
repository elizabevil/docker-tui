//go:build cgo

// This file compiles only when CGO is available. It uses the official Podman
// Go bindings (go.podman.io/podman/v6/pkg/bindings/volumes) which require CGO
// for local libpod communication. TLS connections and explicit API overrides
// use the shared REST implementation because bindings cannot express all of
// the configured transport semantics.

package docker

import (
	"context"
	"fmt"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"go.podman.io/podman/v6/pkg/bindings"
	"go.podman.io/podman/v6/pkg/bindings/volumes"
	entitytypes "go.podman.io/podman/v6/pkg/domain/entities/types"
)

func (c *Client) listVolumesPodman(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
	if c.usePodmanRESTTransport() {
		return c.listVolumesPodmanREST(ctx, options)
	}
	nativeFilters, err := options.NativeFilters()
	if err != nil {
		return nil, err
	}
	bindingContext, err := bindings.NewConnection(ctx, c.Host)
	if err != nil {
		return nil, fmt.Errorf("podman connect: %w", err)
	}
	listOptions := new(volumes.ListOptions).WithFilters(nativeFilters)
	list, err := volumes.List(bindingContext, listOptions)
	if err != nil {
		return nil, fmt.Errorf("podman list volumes: %w", err)
	}
	raw := make([]podmanVolumeConfigResponse, 0, len(list))
	for _, v := range list {
		raw = append(raw, podmanVolumeConfigResponse{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			CreatedAt:  v.CreatedAt,
			Labels:     v.Labels,
			Scope:      v.Scope,
			Options:    v.Options,
			Status:     v.Status,
		})
	}
	return mapPodmanVolumes(raw), nil
}

func (c *Client) inspectVolumePodman(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	if c.usePodmanRESTTransport() {
		return c.inspectVolumePodmanREST(ctx, name)
	}
	bindingContext, err := bindings.NewConnection(ctx, c.Host)
	if err != nil {
		return nil, fmt.Errorf("podman connect: %w", err)
	}
	vol, err := volumes.Inspect(bindingContext, name, nil)
	if err != nil {
		return nil, fmt.Errorf("podman inspect volume %s: %w", name, err)
	}
	return mapPodmanVolumeInspect(podmanVolumeConfigResponse{
		Name:       vol.Name,
		Driver:     vol.Driver,
		Mountpoint: vol.Mountpoint,
		CreatedAt:  vol.CreatedAt,
		Labels:     vol.Labels,
		Scope:      vol.Scope,
		Options:    vol.Options,
		Status:     vol.Status,
	}), nil
}

func (c *Client) removeVolumePodman(ctx context.Context, name string, force bool) error {
	if c.usePodmanRESTTransport() {
		return c.removeVolumePodmanREST(ctx, name, force)
	}
	bindingContext, err := bindings.NewConnection(ctx, c.Host)
	if err != nil {
		return fmt.Errorf("podman connect: %w", err)
	}
	opts := new(volumes.RemoveOptions).WithForce(force)
	return volumes.Remove(bindingContext, name, opts)
}

func (c *Client) createVolumePodman(ctx context.Context, options runtimeapi.VolumeCreateOptions) (*runtimeapi.Volume, error) {
	if c.usePodmanRESTTransport() {
		return c.createVolumePodmanREST(ctx, options)
	}
	bindingContext, err := bindings.NewConnection(ctx, c.Host)
	if err != nil {
		return nil, fmt.Errorf("podman connect: %w", err)
	}
	created, err := volumes.Create(bindingContext, entitytypes.VolumeCreateOptions{Name: options.Name, Driver: options.Driver, Labels: options.Labels, Options: options.Options}, nil)
	if err != nil {
		return nil, fmt.Errorf("podman create volume: %w", err)
	}
	mapped := mapPodmanVolumes([]podmanVolumeConfigResponse{{Name: created.Name, Driver: created.Driver, Mountpoint: created.Mountpoint, CreatedAt: created.CreatedAt, Labels: created.Labels, Scope: created.Scope}})
	return &mapped[0], nil
}

func (c *Client) pruneVolumesPodman(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	if c.usePodmanRESTTransport() {
		return c.pruneVolumesPodmanREST(ctx, options)
	}
	bindingContext, err := bindings.NewConnection(ctx, c.Host)
	if err != nil {
		return runtimeapi.PruneResult{}, fmt.Errorf("podman connect: %w", err)
	}
	reports, err := volumes.Prune(bindingContext, new(volumes.PruneOptions).WithFilters(map[string][]string(options.Filters)))
	if err != nil {
		return runtimeapi.PruneResult{}, fmt.Errorf("podman prune volumes: %w", err)
	}
	result := runtimeapi.PruneResult{Resources: make([]runtimeapi.ResourceResult, 0, len(reports))}
	for _, report := range reports {
		result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: report.Id, Error: report.Err})
		result.SpaceReclaimed += report.Size
	}
	return result, nil
}
