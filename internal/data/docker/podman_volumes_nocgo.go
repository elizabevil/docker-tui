//go:build !cgo

package docker

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Build-tag files select a transport only; REST request construction and
// mapping stay shared so CGO builds can also use REST for TLS connections.
func (c *Client) listVolumesPodman(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
	return c.listVolumesPodmanREST(ctx, options)
}

func (c *Client) inspectVolumePodman(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	return c.inspectVolumePodmanREST(ctx, name)
}

func (c *Client) removeVolumePodman(ctx context.Context, name string, force bool) error {
	return c.removeVolumePodmanREST(ctx, name, force)
}

func (c *Client) createVolumePodman(ctx context.Context, options runtimeapi.VolumeCreateOptions) (*runtimeapi.Volume, error) {
	return c.createVolumePodmanREST(ctx, options)
}

func (c *Client) pruneVolumesPodman(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	return c.pruneVolumesPodmanREST(ctx, options)
}
