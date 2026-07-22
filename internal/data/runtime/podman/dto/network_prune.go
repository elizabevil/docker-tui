package dto

import "encoding/json"

// NetworkPruneReport is the Podman Libpod network prune response item.
// Error is json.RawMessage so callers can pass it directly to podmanReportError.
type NetworkPruneReport struct {
	Name  string          `json:"Name"`
	Error json.RawMessage `json:"Error"`
}
