package podman

import (
	"context"
	"net/url"
	"strconv"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// InspectContainer fetches container detail via the Podman Libpod REST API.
// Returns the raw Podman inspect JSON; callers map to domain types.
func (c *RESTClient) InspectContainer(ctx context.Context, id string) (*dto.ContainerInspectJSON, error) {
	var raw dto.ContainerInspectJSON
	if err := c.Get(ctx, ContainerInspectPath(id), nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// ContainerTop fetches the process listing for a container via the
// Podman Libpod REST API.
func (c *RESTClient) ContainerTop(ctx context.Context, id string) (*dto.ContainerTopResponse, error) {
	var raw dto.ContainerTopResponse
	if err := c.Get(ctx, ContainerTopPath(id), nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// ContainerStats fetches container stats via the Podman Libpod REST API.
func (c *RESTClient) ContainerStats(ctx context.Context, id string) (*dto.StatsJSON, error) {
	query := url.Values{"stream": {strconv.FormatBool(false)}}
	var raw dto.StatsJSON
	if err := c.Get(ctx, ContainerStatsPath(id), query, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}
