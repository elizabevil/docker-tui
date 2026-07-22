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
