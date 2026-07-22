package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// podmanPruneReport is the per-resource response from Podman's volume prune
// endpoint. The Err field is raw JSON because Podman may return null, an empty
// string, or a stringified message depending on the version.
type podmanPruneReport struct {
	ID   string          `json:"Id"`
	Err  json.RawMessage `json:"Err"`
	Size uint64          `json:"Size"`
}

// listVolumesPodmanREST fetches volumes via the Podman Libpod REST API.
func (c *Client) listVolumesPodmanREST(ctx context.Context, options runtimeapi.VolumeListOptions) ([]runtimeapi.Volume, error) {
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

// inspectVolumePodmanREST fetches volume detail via the Podman Libpod REST API.
func (c *Client) inspectVolumePodmanREST(ctx context.Context, name string) (*runtimeapi.VolumeDetail, error) {
	if c.podmanREST == nil {
		return nil, fmt.Errorf("podman REST transport is not initialized")
	}
	var raw podmanVolumeConfigResponse
	if err := c.podmanREST.Get(ctx, "volume.inspect", "/volumes/"+url.PathEscape(name)+"/json", nil, &raw); err != nil {
		return nil, err
	}
	return mapPodmanVolumeInspect(raw), nil
}

// removeVolumePodmanREST deletes a volume via the Podman Libpod REST API.
func (c *Client) removeVolumePodmanREST(ctx context.Context, name string, force bool) error {
	if c.podmanREST == nil {
		return fmt.Errorf("podman REST transport is not initialized")
	}
	query := url.Values{"force": {strconv.FormatBool(force)}}
	return c.podmanREST.DeleteWithQuery(ctx, "volume.remove", "/volumes/"+url.PathEscape(name), query)
}

// createVolumePodmanREST creates a volume via the Podman Libpod REST API.
func (c *Client) createVolumePodmanREST(ctx context.Context, options runtimeapi.VolumeCreateOptions) (*runtimeapi.Volume, error) {
	if c.podmanREST == nil {
		return nil, fmt.Errorf("podman REST transport is not initialized")
	}
	input := struct {
		Name    string            `json:"Name"`
		Driver  string            `json:"Driver"`
		Labels  map[string]string `json:"Labels"`
		Options map[string]string `json:"Options"`
	}{Name: options.Name, Driver: options.Driver, Labels: options.Labels, Options: options.Options}
	var raw podmanVolumeConfigResponse
	if err := c.podmanREST.Post(ctx, "volume.create", "/volumes/create", nil, input, &raw); err != nil {
		return nil, err
	}
	mapped := mapPodmanVolumes([]podmanVolumeConfigResponse{raw})
	return &mapped[0], nil
}

// pruneVolumesPodmanREST removes unused volumes via the Podman Libpod REST API.
func (c *Client) pruneVolumesPodmanREST(ctx context.Context, options runtimeapi.PruneOptions) (runtimeapi.PruneResult, error) {
	if c.podmanREST == nil {
		return runtimeapi.PruneResult{}, fmt.Errorf("podman REST transport is not initialized")
	}
	query, err := podmanFilterQuery(options.Filters)
	if err != nil {
		return runtimeapi.PruneResult{}, err
	}
	var reports []podmanPruneReport
	if err := c.podmanREST.Post(ctx, "volume.prune", "/volumes/prune", query, nil, &reports); err != nil {
		return runtimeapi.PruneResult{}, err
	}
	result := runtimeapi.PruneResult{Resources: make([]runtimeapi.ResourceResult, 0, len(reports))}
	for _, report := range reports {
		result.Resources = append(result.Resources, runtimeapi.ResourceResult{ID: report.ID, Error: podmanReportError(report.Err)})
		result.SpaceReclaimed += report.Size
	}
	return result, nil
}

// podmanFilterQuery encodes a FilterSet as a JSON query parameter for Podman
// REST endpoints that accept a "filters" query string.
func podmanFilterQuery(filters runtimeapi.FilterSet) (url.Values, error) {
	query := make(url.Values)
	if len(filters) == 0 {
		return query, nil
	}
	encoded, err := json.Marshal(filters)
	if err != nil {
		return nil, fmt.Errorf("encode Podman prune filters: %w", err)
	}
	query.Set("filters", string(encoded))
	return query, nil
}

// podmanReportError extracts an error from a Podman prune report's raw JSON
// Error field. Returns nil when the field is null, empty, or an empty string.
func podmanReportError(raw json.RawMessage) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) || bytes.Equal(trimmed, []byte(`""`)) {
		return nil
	}
	var message string
	if json.Unmarshal(trimmed, &message) == nil && message != "" {
		return fmt.Errorf("%s", message)
	}
	return fmt.Errorf("resource prune failed")
}
