//go:build !cgo

// This file compiles when CGO is unavailable. It uses the Podman REST API
// (libpod endpoints) via the shared RESTClient, which requires no native
// dependencies. Filter encoding, query construction, and error classification
// are handled identically to the CGO path through the shared mapper in
// podman_network_mapper.go.

package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func (c *Client) listNetworksPodman(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	if c.podmanREST == nil {
		return nil, fmt.Errorf("Podman REST transport is not initialized")
	}
	nativeFilters, err := options.NativeFilters()
	if err != nil {
		return nil, err
	}
	query := make(url.Values)
	if len(nativeFilters) > 0 {
		encoded, err := json.Marshal(nativeFilters)
		if err != nil {
			return nil, fmt.Errorf("encode Podman network filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []podmanNetworkItem
	if err := c.podmanREST.Get(ctx, "network.list", "/networks/json", query, &raw); err != nil {
		return nil, err
	}
	return mapPodmanNetworks(raw), nil
}

func (c *Client) inspectNetworkPodman(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	if c.podmanREST == nil {
		return nil, fmt.Errorf("Podman REST transport is not initialized")
	}
	var raw podmanNetworkItem
	if err := c.podmanREST.Get(ctx, "network.inspect", "/networks/"+url.PathEscape(id)+"/json", nil, &raw); err != nil {
		return nil, err
	}
	return mapPodmanNetworkInspect(raw), nil
}

func (c *Client) removeNetworkPodman(ctx context.Context, id string) error {
	if c.podmanREST == nil {
		return fmt.Errorf("Podman REST transport is not initialized")
	}
	return c.podmanREST.Delete(ctx, "network.remove", "/networks/"+url.PathEscape(id))
}
