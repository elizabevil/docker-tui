//go:build !cgo

// This file compiles when CGO is unavailable. It uses the Podman REST API
// (libpod endpoints) via the shared RESTClient, which requires no native
// dependencies. Filter encoding, query construction, and error classification
// are handled identically to the CGO path through the shared mapper in
// podman_volume_mapper.go.

package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func (c *Client) listVolumesPodman(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
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
			return nil, fmt.Errorf("encode Podman volume filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []podmanVolumeConfigResponse
	if err := c.podmanREST.Get(ctx, "volume.list", "/volumes/json", query, &raw); err != nil {
		return nil, err
	}
	return mapPodmanVolumes(raw), nil
}

func (c *Client) inspectVolumePodman(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	if c.podmanREST == nil {
		return nil, fmt.Errorf("Podman REST transport is not initialized")
	}
	var raw podmanVolumeConfigResponse
	if err := c.podmanREST.Get(ctx, "volume.inspect", "/volumes/"+url.PathEscape(name)+"/json", nil, &raw); err != nil {
		return nil, err
	}
	return mapPodmanVolumeInspect(raw), nil
}

func (c *Client) removeVolumePodman(ctx context.Context, name string, force bool) error {
	if c.podmanREST == nil {
		return fmt.Errorf("Podman REST transport is not initialized")
	}
	return c.podmanREST.Delete(ctx, "volume.remove", "/volumes/"+url.PathEscape(name))
}
