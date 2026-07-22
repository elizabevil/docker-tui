//go:build cgo

package podman

// Type aliases to Podman native types.
// The HTTP REST JSON responses use the same field names and types as these
// Go structs, so they can be deserialized directly without custom DTOs.
//
// Aliases avoid duplication: the CGO bindings path and the REST path share
// the same struct definitions from the Podman driver.

import (
	"encoding/json"
	"time"

	"go.podman.io/podman/v6/libpod/define"
	"go.podman.io/podman/v6/pkg/domain/entities/types"
)

// ImageItem is an alias for the Podman image list DTO.
// Verified compatible with Podman 5.4.2 /images/json response.
type ImageItem = types.ImageSummary

// VolumeItem is an alias for the Podman volume inspect/list DTO.
// Verified compatible with Podman 5.4.2 /volumes/json response.
type VolumeItem = define.InspectVolumeData

// VolumePruneReport is the Podman volume prune response.
// The CGO path uses reports.PruneReport but the REST path returns
// a simpler structure with ID, Size, and error string.
//
// Err is json.RawMessage so callers can pass it directly to
// podmanReportError. Size is uint64 to match runtimeapi.PruneResult.SpaceReclaimed.
type VolumePruneReport struct {
	ID   string          `json:"ID"`
	Size uint64          `json:"Size"`
	Err  json.RawMessage `json:"Err"`
}

// NetworkPruneReport is the Podman network prune response.
// Error is json.RawMessage so callers can pass it directly to podmanReportError.
type NetworkPruneReport struct {
	Name  string          `json:"Name"`
	Error json.RawMessage `json:"Error"`
}

// EventItem is the Podman Libpod REST event DTO.
// The HTTP response has a flat structure (Type, Status, Action, ID, Name,
// Attributes) which differs from the CGO types.Event (nested Actor.Attributes).
// We keep a custom struct for REST deserialization.
type EventItem struct {
	Type       string            `json:"Type"`
	Status     string            `json:"Status"`
	Action     string            `json:"Action"`
	ID         string            `json:"ID"`
	Name       string            `json:"Name"`
	Time       time.Time         `json:"Time"`
	TimeNano   int64             `json:"TimeNano"`
	Attributes map[string]string `json:"Attributes"`
}

// FormatPodmanTime formats a Podman time.Time as RFC3339. Returns "" for zero
// times (e.g. when Podman does not populate the field).
func FormatPodmanTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}
