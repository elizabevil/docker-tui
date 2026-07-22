package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ListImagesREST fetches images via the Podman Libpod REST API.
// Returns raw Podman image items; callers map to domain types.
func ListImagesREST(ctx context.Context, client *Client, all bool, filters map[string][]string) ([]ImageItem, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	query := make(url.Values)
	if all {
		query.Set("all", "true")
	}
	if len(filters) > 0 {
		encoded, err := json.Marshal(filters)
		if err != nil {
			return nil, fmt.Errorf("encode Podman image filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []ImageItem
	if err := client.REST.Get(ctx, "image.list", PathImageList, query, &raw); err != nil {
		return nil, fmt.Errorf("podman list: %w", err)
	}
	return raw, nil
}
