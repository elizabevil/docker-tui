package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// ListImages fetches images via the Podman Libpod REST API.
// Returns raw Podman image items; callers map to domain types.
func (c *RESTClient) ListImages(ctx context.Context, opts dto.ImageListOptions) ([]ImageItem, error) {
	query := make(url.Values)
	if opts.All {
		query.Set("all", "true")
	}
	if len(opts.Filters) > 0 {
		encoded, err := json.Marshal(opts.Filters)
		if err != nil {
			return nil, fmt.Errorf("encode Podman image filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []ImageItem
	if err := c.Get(ctx, PathImageList, query, &raw); err != nil {
		return nil, fmt.Errorf("podman list: %w", err)
	}
	return raw, nil
}
