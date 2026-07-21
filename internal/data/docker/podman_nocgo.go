//go:build !cgo

package docker

import (
	"fmt"
)

func (c *Client) listImagesPodman() ([]ImageSummary, error) {
	if c.podmanREST == nil {
		return nil, fmt.Errorf("Podman REST transport is not initialized")
	}
	var raw []podmanImageSummary
	if err := c.podmanREST.Get(c.ctx, "image.list", "/images/json", nil, &raw); err != nil {
		return nil, fmt.Errorf("podman list: %w", err)
	}
	return mapPodmanImageSummaries(raw), nil
}
