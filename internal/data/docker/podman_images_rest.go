package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// listImagesPodmanREST fetches images via the Podman Libpod REST API.
func (c *Client) listImagesPodmanREST(ctx context.Context, options runtimeapi.ImageListOptions) ([]ImageSummary, error) {
	if c.podmanREST == nil {
		return nil, errPodmanRESTNotReady
	}
	nativeFilters, err := options.NativeFilters()
	if err != nil {
		return nil, err
	}
	query := make(url.Values)
	if options.All {
		query.Set("all", "true")
	}
	if len(nativeFilters) > 0 {
		encoded, err := json.Marshal(nativeFilters)
		if err != nil {
			return nil, fmt.Errorf("encode Podman image filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []podmanImageSummary
	if err := c.podmanREST.Get(ctx, "image.list", "/images/json", query, &raw); err != nil {
		return nil, fmt.Errorf("podman list: %w", err)
	}
	return mapPodmanImageSummaries(raw), nil
}
