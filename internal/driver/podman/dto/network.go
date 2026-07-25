package dto

// NetworkCreateRequest is the Podman Libpod REST request body for creating a
// network.
type NetworkCreateRequest struct {
	Name        string            `json:"name"`
	Driver      string            `json:"driver"`
	Internal    bool              `json:"internal"`
	IPv6Enabled bool              `json:"ipv6_enabled"`
	Labels      map[string]string `json:"labels,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
}
