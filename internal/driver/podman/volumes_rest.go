package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// ListVolumes fetches volumes via the Podman Libpod REST API.
// Returns raw Podman volume items; callers map to domain types.
func (c *RESTClient) ListVolumes(ctx context.Context, opts dto.VolumeListOptions) ([]VolumeItem, error) {
	query := make(url.Values)
	if len(opts.Filters) > 0 {
		encoded, err := json.Marshal(opts.Filters)
		if err != nil {
			return nil, fmt.Errorf("encode Podman volume filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []VolumeItem
	if err := c.Get(ctx, PathVolumeList, query, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// InspectVolume fetches volume detail via the Podman Libpod REST API.
// Returns the raw Podman volume item; callers map to domain types.
func (c *RESTClient) InspectVolume(ctx context.Context, name string) (*VolumeItem, error) {
	var raw VolumeItem
	if err := c.Get(ctx, VolumePath(name, "/json"), nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// RemoveVolume deletes a volume via the Podman Libpod REST API.
func (c *RESTClient) RemoveVolume(ctx context.Context, name string, force bool) error {
	query := url.Values{"force": {strconv.FormatBool(force)}}
	return c.DeleteWithQuery(ctx, VolumePath(name, ""), query)
}

// CreateVolume creates a volume via the Podman Libpod REST API.
func (c *RESTClient) CreateVolume(ctx context.Context, opts dto.VolumeItem) (*VolumeItem, error) {
	input := dto.VolumeCreateRequest{Name: opts.Name, Driver: opts.Driver, Labels: opts.Labels, Options: opts.Options}
	var raw VolumeItem
	if err := c.Post(ctx, PathVolumeCreate, nil, input, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// PruneVolumes removes unused volumes via the Podman Libpod REST API.
// Returns raw prune report items; callers assemble domain PruneResult.
func (c *RESTClient) PruneVolumes(ctx context.Context, filters map[string][]string) ([]dto.VolumePruneReportItem, error) {
	query, err := PodmanFilterQuery(filters)
	if err != nil {
		return nil, err
	}
	var reports []dto.VolumePruneReportItem
	if err := c.Post(ctx, PathVolumePrune, query, nil, &reports); err != nil {
		return nil, err
	}
	return reports, nil
}
