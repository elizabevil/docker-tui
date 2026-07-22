package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	runtimepodman "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

// podmanNetworkPruneReport is the per-resource response from Podman's network
// prune endpoint. The Error field is raw JSON because Podman may return null,
// an empty string, or a stringified message depending on the version.
type podmanNetworkPruneReport struct {
	Name  string          `json:"Name"`
	Error json.RawMessage `json:"Error"`
}

// listNetworksPodmanREST fetches networks via the Podman Libpod REST API.
func (c *Client) listNetworksPodmanREST(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	if c.podmanREST == nil {
		return nil, errPodmanRESTNotReady
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

// inspectNetworkPodmanREST fetches network detail via the Podman Libpod REST API.
func (c *Client) inspectNetworkPodmanREST(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	if c.podmanREST == nil {
		return nil, errPodmanRESTNotReady
	}
	var raw podmanNetworkItem
	if err := c.podmanREST.Get(ctx, "network.inspect", runtimepodman.NetworkPath(id, "/json"), nil, &raw); err != nil {
		return nil, err
	}
	return mapPodmanNetworkInspect(raw), nil
}

// removeNetworkPodmanREST deletes a network via the Podman Libpod REST API.
func (c *Client) removeNetworkPodmanREST(ctx context.Context, id string) error {
	if c.podmanREST == nil {
		return errPodmanRESTNotReady
	}
	return c.podmanREST.Delete(ctx, "network.remove", runtimepodman.NetworkPath(id, ""))
}

// createNetworkPodmanREST creates a network via the Podman Libpod REST API.
func (c *Client) createNetworkPodmanREST(ctx context.Context, options runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	if c.podmanREST == nil {
		return nil, errPodmanRESTNotReady
	}
	input := struct {
		Name        string            `json:"name"`
		Driver      string            `json:"driver"`
		Internal    bool              `json:"internal"`
		IPv6Enabled bool              `json:"ipv6_enabled"`
		Labels      map[string]string `json:"labels,omitempty"`
		Options     map[string]string `json:"options,omitempty"`
	}{Name: options.Name, Driver: options.Driver, Internal: options.Internal, IPv6Enabled: options.EnableIPv6, Labels: options.Labels, Options: options.Options}
	var raw podmanNetworkItem
	if err := c.podmanREST.Post(ctx, "network.create", "/networks/create", nil, input, &raw); err != nil {
		return nil, err
	}
	mapped := mapPodmanNetworks([]podmanNetworkItem{raw})
	return &mapped[0], nil
}

// pruneNetworksPodmanREST removes unused networks via the Podman Libpod REST API.
func (c *Client) pruneNetworksPodmanREST(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	if c.podmanREST == nil {
		return runtimeapi.PruneResult{}, errPodmanRESTNotReady
	}
	query, err := podmanFilterQuery(options.Filters)
	if err != nil {
		return runtimeapi.PruneResult{}, err
	}
	var reports []podmanNetworkPruneReport
	if err := c.podmanREST.Post(ctx, "network.prune", "/networks/prune", query, nil, &reports); err != nil {
		return runtimeapi.PruneResult{}, err
	}
	result := runtimeapi.PruneResult{Resources: make([]runtimeapi.ResourceResult, 0, len(reports))}
	for _, report := range reports {
		result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: report.Name, Error: podmanReportError(report.Error)})
	}
	return result, nil
}
