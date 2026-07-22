package docker

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// inspectContainerPodmanREST fetches container detail via the Podman Libpod REST API.
func (c *Client) inspectContainerPodmanREST(ctx context.Context, id string) (*runtimeapi.ContainerDetail, error) {
	if c.podmanREST == nil {
		return nil, fmt.Errorf("podman REST transport is not initialized")
	}
	var raw inspectContainer
	if err := c.podmanREST.Get(ctx, "container.inspect", "/containers/"+url.PathEscape(id)+"/json", nil, &raw); err != nil {
		return nil, err
	}
	return mapContainerInspectResponse(raw), nil
}

func (c *Client) containerTopPodmanREST(ctx context.Context, id string) (runtimeapi.ContainerProcesses, error) {
	if c.podmanREST == nil {
		return runtimeapi.ContainerProcesses{}, fmt.Errorf("podman REST transport is not initialized")
	}
	var raw struct {
		Titles    []string   `json:"Titles"`
		Processes [][]string `json:"Processes"`
	}
	if err := c.podmanREST.Get(ctx, "container.top", "/containers/"+url.PathEscape(id)+"/top", nil, &raw); err != nil {
		return runtimeapi.ContainerProcesses{}, err
	}
	return runtimeapi.ContainerProcesses{Titles: raw.Titles, Processes: raw.Processes}, nil
}

func (c *Client) containerStatsPodmanREST(ctx context.Context, id string) (runtimeapi.ContainerStats, error) {
	if c.podmanREST == nil {
		return runtimeapi.ContainerStats{}, fmt.Errorf("podman REST transport is not initialized")
	}
	query := url.Values{"stream": {strconv.FormatBool(false)}}
	// The per-container Libpod route deliberately exposes the Docker-compatible
	// stats schema, so both adapters share the same calculation semantics.
	var raw statsJSON
	if err := c.podmanREST.Get(ctx, "container.stats", "/containers/"+url.PathEscape(id)+"/stats", query, &raw); err != nil {
		return runtimeapi.ContainerStats{}, err
	}
	cpu, usage, limit, memory, rx, tx := computeStats(raw)
	return runtimeapi.ContainerStats{
		ReadAt: time.Now(), CPUPercent: cpu,
		MemoryUsage: usage, MemoryLimit: limit,
		MemoryPercent: memory, NetworkRx: rx, NetworkTx: tx,
	}, nil
}
