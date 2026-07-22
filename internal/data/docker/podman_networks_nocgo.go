//go:build !cgo

package docker

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Non-CGO build: all Podman network operations delegate directly to the shared
// REST transport. The CGO build can also use REST when TLS or API override is
// configured, so these methods are the single fallback path.

// listNetworksPodman delegates to the Podman REST transport.
func (c *Client) listNetworksPodman(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	return c.listNetworksPodmanREST(ctx, options)
}

// inspectNetworkPodman delegates to the Podman REST transport.
func (c *Client) inspectNetworkPodman(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	return c.inspectNetworkPodmanREST(ctx, id)
}

// removeNetworkPodman delegates to the Podman REST transport.
func (c *Client) removeNetworkPodman(ctx context.Context, id string) error {
	return c.removeNetworkPodmanREST(ctx, id)
}

// createNetworkPodman delegates to the Podman REST transport.
func (c *Client) createNetworkPodman(ctx context.Context, options runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	return c.createNetworkPodmanREST(ctx, options)
}

// pruneNetworksPodman delegates to the Podman REST transport.
func (c *Client) pruneNetworksPodman(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	return c.pruneNetworksPodmanREST(ctx, options)
}
