//go:build !cgo

package docker

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Build-tag files select a transport only; the shared REST implementation is
// also available to CGO builds whose TLS/API options bindings cannot express.
func (c *Client) listNetworksPodman(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	return c.listNetworksPodmanREST(ctx, options)
}

func (c *Client) inspectNetworkPodman(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	return c.inspectNetworkPodmanREST(ctx, id)
}

func (c *Client) removeNetworkPodman(ctx context.Context, id string) error {
	return c.removeNetworkPodmanREST(ctx, id)
}

func (c *Client) createNetworkPodman(ctx context.Context, options runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	return c.createNetworkPodmanREST(ctx, options)
}

func (c *Client) pruneNetworksPodman(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	return c.pruneNetworksPodmanREST(ctx, options)
}
