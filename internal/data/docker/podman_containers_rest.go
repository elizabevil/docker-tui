package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// listContainersPodmanREST fetches containers via the Podman Libpod REST API.
func (c *Client) listContainersPodmanREST(ctx context.Context, options ContainerListOptions) ([]ContainerSummary, error) {
	if c.podmanREST == nil {
		return nil, errPodmanRESTNotReady
	}
	nativeFilters, err := options.NativeFilters()
	if err != nil {
		return nil, err
	}
	query := make(url.Values)
	query.Set("all", strconv.FormatBool(options.All))
	if options.Limit > 0 {
		query.Set("last", strconv.Itoa(options.Limit))
	}
	if len(nativeFilters) > 0 {
		encoded, err := json.Marshal(nativeFilters)
		if err != nil {
			return nil, fmt.Errorf("encode Podman container filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []podmanContainerSummary
	if err := c.podmanREST.Get(ctx, "container.list", "/containers/json", query, &raw); err != nil {
		return nil, err
	}
	return mapPodmanContainerSummaries(raw), nil
}
