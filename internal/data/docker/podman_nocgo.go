//go:build !cgo

package docker

import (
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// Build-tag files select a transport only; REST request construction and
// mapping stay shared so both build modes expose identical behavior.
func (c *Client) listImagesPodman(options runtimeapi.ImageListOptions) ([]ImageSummary, error) {
	return c.listImagesPodmanREST(options)
}
