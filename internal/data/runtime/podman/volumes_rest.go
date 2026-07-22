package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"
)

// ListVolumesREST fetches volumes via the Podman Libpod REST API.
// Returns raw Podman volume items; callers map to domain types.
func ListVolumesREST(ctx context.Context, client *Client, filters map[string][]string) ([]VolumeItem, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	query := make(url.Values)
	if len(filters) > 0 {
		encoded, err := json.Marshal(filters)
		if err != nil {
			return nil, fmt.Errorf("encode Podman volume filters: %w", err)
		}
		query.Set("filters", string(encoded))
	}
	var raw []VolumeItem
	if err := client.REST.Get(ctx, "volume.list", PathVolumeList, query, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// InspectVolumeREST fetches volume detail via the Podman Libpod REST API.
// Returns the raw Podman volume item; callers map to domain types.
func InspectVolumeREST(ctx context.Context, client *Client, name string) (*VolumeItem, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	var raw VolumeItem
	if err := client.REST.Get(ctx, "volume.inspect", VolumePath(name, "/json"), nil, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// RemoveVolumeREST deletes a volume via the Podman Libpod REST API.
func RemoveVolumeREST(ctx context.Context, client *Client, name string, force bool) error {
	if client.REST == nil {
		return errPodmanRESTNotReady
	}
	query := url.Values{"force": {strconv.FormatBool(force)}}
	return client.REST.DeleteWithQuery(ctx, "volume.remove", VolumePath(name, ""), query)
}

// CreateVolumeREST creates a volume via the Podman Libpod REST API.
// Returns the raw Podman volume item; callers map to domain types.
func CreateVolumeREST(ctx context.Context, client *Client, name, driver string, labels, options map[string]string) (*VolumeItem, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	input := dto.VolumeCreateRequest{Name: name, Driver: driver, Labels: labels, Options: options}
	var raw VolumeItem
	if err := client.REST.Post(ctx, "volume.create", PathVolumeCreate, nil, input, &raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

// PruneVolumesREST removes unused volumes via the Podman Libpod REST API.
// Returns raw prune report items; callers assemble domain PruneResult.
func PruneVolumesREST(ctx context.Context, client *Client, filters map[string][]string) ([]dto.VolumePruneReportItem, error) {
	if client.REST == nil {
		return nil, errPodmanRESTNotReady
	}
	query, err := PodmanFilterQuery(filters)
	if err != nil {
		return nil, err
	}
	var reports []dto.VolumePruneReportItem
	if err := client.REST.Post(ctx, "volume.prune", PathVolumePrune, query, nil, &reports); err != nil {
		return nil, err
	}
	return reports, nil
}
