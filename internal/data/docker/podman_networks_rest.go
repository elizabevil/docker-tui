package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type podmanNetworkPruneReport struct {
	Name  string          `json:"Name"`
	Error json.RawMessage `json:"Error"`
}

func (c *Client) listNetworksPodmanREST(ctx context.Context, options runtimeapi.NetworkListOptions) ([]runtimeapi.Network, error) {
	if c.podmanREST == nil {
		return nil, fmt.Errorf("podman REST transport is not initialized")
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

func (c *Client) inspectNetworkPodmanREST(ctx context.Context, id string) (*runtimeapi.NetworkDetail, error) {
	if c.podmanREST == nil {
		return nil, fmt.Errorf("podman REST transport is not initialized")
	}
	var raw podmanNetworkItem
	if err := c.podmanREST.Get(ctx, "network.inspect", "/networks/"+url.PathEscape(id)+"/json", nil, &raw); err != nil {
		return nil, err
	}
	return mapPodmanNetworkInspect(raw), nil
}

func (c *Client) removeNetworkPodmanREST(ctx context.Context, id string) error {
	if c.podmanREST == nil {
		return fmt.Errorf("podman REST transport is not initialized")
	}
	return c.podmanREST.Delete(ctx, "network.remove", "/networks/"+url.PathEscape(id))
}

func (c *Client) createNetworkPodmanREST(ctx context.Context, options runtimeapi.NetworkCreateOptions) (*runtimeapi.Network, error) {
	if c.podmanREST == nil {
		return nil, fmt.Errorf("podman REST transport is not initialized")
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

func (c *Client) pruneNetworksPodmanREST(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	if c.podmanREST == nil {
		return runtimeapi.PruneResult{}, fmt.Errorf("podman REST transport is not initialized")
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
