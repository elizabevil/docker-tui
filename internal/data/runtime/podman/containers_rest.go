package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListContainersREST fetches containers via the Podman Libpod REST API.
// Returns raw Podman container items; callers map to domain types.
func ListContainersREST(ctx context.Context, client *Client, all bool, limit int, filters map[string][]string) ([]ContainerItem, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	query := make(url.Values)
	query.Set("all", strconv.FormatBool(all))
	if limit > 0 {
		query.Set("last", strconv.Itoa(limit))
	}
	if len(filters) > 0 {
		encoded, err := json.Marshal(filters)
		if err != nil {
			return nil, fmt.Errorf("encode Podman container filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []ContainerItem
	if err := client.REST.Get(ctx, "container.list", PathContainerList, query, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
