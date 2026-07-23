package dto

import "encoding/json"

// NetworkPruneReport is the Podman Libpod network prune response item.
// Error is json.RawMessage so callers can pass it directly to podmanReportError.
type NetworkPruneReport struct {
	Name  string          `json:"Name"`
	Error json.RawMessage `json:"Error"`
}

// VolumePruneReport is the Podman volume prune response item.
//
// Err is json.RawMessage because the prune endpoint returns Err as either a
// JSON string or null; callers pass it to podmanReportError which expects
// json.RawMessage. Size is uint64 so it can be accumulated into
// runtimeapi.PruneResult.SpaceReclaimed (also uint64) without conversion.
type VolumePruneReport struct {
	ID   string          `json:"ID"`
	Size uint64          `json:"Size"`
	Err  json.RawMessage `json:"Err"`
}
