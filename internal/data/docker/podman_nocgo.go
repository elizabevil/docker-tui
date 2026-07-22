//go:build !cgo

package docker

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Non-CGO build: all Podman image operations delegate directly to the shared
// REST transport. The CGO build can also use REST when TLS or API override is
// configured, so these methods are the single fallback path.

// listImagesPodman delegates to the Podman REST transport.
func (c *Client) listImagesPodman(ctx context.Context, options runtimeapi.ImageListOptions) ([]ImageSummary, error) {
	return c.listImagesPodmanREST(ctx, options)
}
