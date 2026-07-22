//go:build !cgo

package docker

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Non-CGO build: all Podman volume operations delegate directly to the shared
// REST transport. The CGO build can also use REST when TLS or API override is
// configured, so these methods are the single fallback path.

// listVolumesPodman delegates to the Podman REST transport.
func (c *Client) listVolumesPodman(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
	return c.listVolumesPodmanREST(ctx, options)
}

// inspectVolumePodman delegates to the Podman REST transport.
func (c *Client) inspectVolumePodman(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	return c.inspectVolumePodmanREST(ctx, name)
}

// removeVolumePodman delegates to the Podman REST transport.
func (c *Client) removeVolumePodman(ctx context.Context, name string, force bool) error {
	return c.removeVolumePodmanREST(ctx, name, force)
}

// createVolumePodman delegates to the Podman REST transport.
func (c *Client) createVolumePodman(ctx context.Context, options runtimeapi.VolumeCreateOptions) (*runtimeapi.Volume, error) {
	return c.createVolumePodmanREST(ctx, options)
}

// pruneVolumesPodman delegates to the Podman REST transport.
func (c *Client) pruneVolumesPodman(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	return c.pruneVolumesPodmanREST(ctx, options)
}
