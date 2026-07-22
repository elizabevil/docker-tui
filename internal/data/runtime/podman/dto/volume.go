package dto

// VolumeCreateRequest is the Podman Libpod REST request body for creating a
// volume.
type VolumeCreateRequest struct {
	Name    string            `json:"Name"`
	Driver  string            `json:"Driver"`
	Labels  map[string]string `json:"Labels"`
	Options map[string]string `json:"Options"`
}
