//go:build !cgo

package docker

import (
	"context"
)

// Non-CGO build: all Podman container operations delegate directly to the shared
// REST transport. The CGO build can also use REST when TLS or API override is
// configured, so these methods are the single fallback path.

// listContainersPodman delegates to the Podman REST transport.
func (c *Client) listContainersPodman(ctx context.Context, options ContainerListOptions) ([]ContainerSummary, error) {
	return c.listContainersPodmanREST(ctx, options)
}
