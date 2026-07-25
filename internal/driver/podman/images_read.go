package podman

import (
	"context"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// InspectImage fetches image detail via the Podman Libpod REST API. Returns
// the raw Podman inspect JSON; callers map to domain types.
func (c *RESTClient) InspectImage(ctx context.Context, id string) (*dto.ImageInspectJSON, error) {
	var raw dto.ImageInspectJSON
	if err := c.Get(ctx, ImageInspectPath(id), nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// ImageHistory fetches the per-layer history for an image via the Podman
// Libpod REST API. The history endpoint accepts image references (ID, name,
// or digest); the caller is responsible for passing a valid reference.
func (c *RESTClient) ImageHistory(ctx context.Context, id string) ([]dto.LayerHistoryEntry, error) {
	var raw []dto.LayerHistoryEntry
	path := "/libpod/images/" + escape(id) + "/history"
	if err := c.Get(ctx, path, nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
