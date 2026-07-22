package podman

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

// InspectContainerREST fetches container detail via the Podman Libpod REST API.
// Returns the raw Podman inspect JSON; callers map to domain types.
func InspectContainerREST(ctx context.Context, client *Client, id string) (*ContainerInspectJSON, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	var raw ContainerInspectJSON
	path := fmt.Sprintf(PathContainerInspect, url.PathEscape(id))
	if err := client.REST.Get(ctx, "container.inspect", path, nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// ContainerTopREST fetches the process listing for a container via the
// Podman Libpod REST API.
func ContainerTopREST(ctx context.Context, client *Client, id string) (*dto.ContainerTopResponse, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	var raw dto.ContainerTopResponse
	path := fmt.Sprintf(PathContainerTop, url.PathEscape(id))
	if err := client.REST.Get(ctx, "container.top", path, nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// ContainerStatsREST fetches container stats via the Podman Libpod REST API.
func ContainerStatsREST(ctx context.Context, client *Client, id string) (*dto.StatsJSON, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	query := url.Values{"stream": {strconv.FormatBool(false)}}
	var raw dto.StatsJSON
	path := fmt.Sprintf(PathContainerStats, url.PathEscape(id))
	if err := client.REST.Get(ctx, "container.stats", path, query, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}
