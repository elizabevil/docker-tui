//go:build !cgo

package docker

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Build-tag files select a transport only; REST request construction and
// mapping stay shared so both build modes expose identical behavior.
func (c *Client) listImagesPodman(ctx context.Context, options runtimeapi.ImageListOptions) ([]ImageSummary, error) {
	return c.listImagesPodmanREST(ctx, options)
}
