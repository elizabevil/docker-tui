package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// ListContainers fetches containers via the Podman Libpod REST API.
// Returns raw Podman container items; callers map to domain types.
func (c *RESTClient) ListContainers(ctx context.Context, opts dto.ContainerListOptions) ([]ContainerItem, error) {
	query := make(url.Values)
	query.Set("all", strconv.FormatBool(opts.All))
	if opts.Limit > 0 {
		query.Set("last", strconv.Itoa(opts.Limit))
	}
	if len(opts.Filters) > 0 {
		encoded, err := json.Marshal(opts.Filters)
		if err != nil {
			return nil, fmt.Errorf("encode Podman container filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []ContainerItem
	if err := c.Get(ctx, PathContainerList, query, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
