package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func (c *Client) listVolumesPodmanREST(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
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

func (c *Client) inspectVolumePodmanREST(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	if c.podmanREST == nil {
		return nil, fmt.Errorf("Podman REST transport is not initialized")
	}
	var raw podmanVolumeConfigResponse
	if err := c.podmanREST.Get(ctx, "volume.inspect", "/volumes/"+url.PathEscape(name)+"/json", nil, &raw); err != nil {
		return nil, err
	}
	return mapPodmanVolumeInspect(raw), nil
}

func (c *Client) removeVolumePodmanREST(ctx context.Context, name string, force bool) error {
	if c.podmanREST == nil {
		return fmt.Errorf("Podman REST transport is not initialized")
	}
	query := url.Values{"force": {strconv.FormatBool(force)}}
	return c.podmanREST.DeleteWithQuery(ctx, "volume.remove", "/volumes/"+url.PathEscape(name), query)
}
