package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// ListNetworks fetches networks via the Podman Libpod REST API.
// Returns raw Podman network items; callers map to domain types.
func (c *RESTClient) ListNetworks(ctx context.Context, opts dto.NetworkListOptions) ([]Network, error) {
	query := make(url.Values)
	if len(opts.Filters) > 0 {
		encoded, err := json.Marshal(opts.Filters)
		if err != nil {
			return nil, fmt.Errorf("encode Podman network filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []Network
	if err := c.Get(ctx, PathNetworkList, query, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// InspectNetwork fetches network detail via the Podman Libpod REST API.
// Returns the raw Podman network inspect item; callers map to domain types.
func (c *RESTClient) InspectNetwork(ctx context.Context, id string) (*NetworkInspectItem, error) {
	var raw NetworkInspectItem
	if err := c.Get(ctx, NetworkPath(id, "/json"), nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// RemoveNetwork deletes a network via the Podman Libpod REST API.
func (c *RESTClient) RemoveNetwork(ctx context.Context, id string) error {
	return c.Delete(ctx, NetworkPath(id, ""))
}

// CreateNetwork creates a network via the Podman Libpod REST API.
func (c *RESTClient) CreateNetwork(ctx context.Context, opts dto.Network) (*Network, error) {
	input := dto.NetworkCreateRequest{
		Name: opts.Name, Driver: opts.Driver, Internal: opts.Internal,
		IPv6Enabled: opts.IPv6Enabled, Labels: opts.Labels, Options: opts.Options,
	}
	var raw Network
	if err := c.Post(ctx, PathNetworkCreate, nil, input, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// PruneNetworks removes unused networks via the Podman Libpod REST API.
// Returns raw prune report items; callers assemble domain PruneResult.
func (c *RESTClient) PruneNetworks(ctx context.Context, filters map[string][]string) ([]dto.NetworkPruneReportItem, error) {
	query, err := PodmanFilterQuery(filters)
	if err != nil {
		return nil, err
	}
	var reports []dto.NetworkPruneReportItem
	if err := c.Post(ctx, PathNetworkPrune, query, nil, &reports); err != nil {
		return nil, err
	}
	return reports, nil
}
