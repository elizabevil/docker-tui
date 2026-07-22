//go:build !cgo

package docker

import (
	"context"
)

// Build-tag files select a transport only; REST request construction and
// mapping stay shared so CGO builds can honor TLS and API overrides.
func (c *Client) listContainersPodman(ctx context.Context, options ContainerListOptions) ([]ContainerSummary, error) {
	return c.listContainersPodmanREST(ctx, options)
}
