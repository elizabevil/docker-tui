package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

// ListNetworksREST fetches networks via the Podman Libpod REST API.
// Returns raw Podman network items; callers map to domain types.
func ListNetworksREST(ctx context.Context, client *Client, filters map[string][]string) ([]Network, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	query := make(url.Values)
	if len(filters) > 0 {
		encoded, err := json.Marshal(filters)
		if err != nil {
			return nil, fmt.Errorf("encode Podman network filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []Network
	if err := client.REST.Get(ctx, "network.list", PathNetworkList, query, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// InspectNetworkREST fetches network detail via the Podman Libpod REST API.
// Returns the raw Podman network inspect item; callers map to domain types.
func InspectNetworkREST(ctx context.Context, client *Client, id string) (*NetworkInspectItem, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	var raw NetworkInspectItem
	if err := client.REST.Get(ctx, "network.inspect", NetworkPath(id, "/json"), nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// RemoveNetworkREST deletes a network via the Podman Libpod REST API.
func RemoveNetworkREST(ctx context.Context, client *Client, id string) error {
	if client.REST == nil {
		return errPodmanRESTNotReady
	}
	return client.REST.Delete(ctx, "network.remove", NetworkPath(id, ""))
}

// CreateNetworkREST creates a network via the Podman Libpod REST API.
// Returns the raw Podman network item; callers map to domain types.
func CreateNetworkREST(ctx context.Context, client *Client, name, driver string, internal, ipv6 bool, labels, options map[string]string) (*Network, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	input := dto.NetworkCreateRequest{
		Name: name, Driver: driver, Internal: internal,
		IPv6Enabled: ipv6, Labels: labels, Options: options,
	}
	var raw Network
	if err := client.REST.Post(ctx, "network.create", PathNetworkCreate, nil, input, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// PruneNetworksREST removes unused networks via the Podman Libpod REST API.
// Returns raw prune report items; callers assemble domain PruneResult.
func PruneNetworksREST(ctx context.Context, client *Client, filters map[string][]string) ([]dto.NetworkPruneReportItem, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	query, err := PodmanFilterQuery(filters)
	if err != nil {
		return nil, err
	}
	var reports []dto.NetworkPruneReportItem
	if err := client.REST.Post(ctx, "network.prune", PathNetworkPrune, query, nil, &reports); err != nil {
		return nil, err
	}
	return reports, nil
}
