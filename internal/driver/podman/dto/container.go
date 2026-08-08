package dto

// ContainerTopResponse is the Podman Libpod REST response for the container
// top endpoint.
type ContainerTopResponse struct {
	Titles    []string   `json:"Titles"`
	Processes [][]string `json:"Processes"`
}

// SpecGenerator carries the Podman Libpod container create request body
// (POST /libpod/containers/create). Field shapes and JSON tags match the
// swagger SpecGenerator definition; only the fields surfaced via
// ContainerCreate are declared.
type SpecGenerator struct {
	Name       string            `json:"name,omitempty"`
	Image      string            `json:"image,omitempty"`
	Command    []string          `json:"command,omitempty"`
	Entrypoint []string          `json:"entrypoint,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	User       string            `json:"user,omitempty"`
	Terminal   bool              `json:"terminal,omitempty"`
	Remove     bool              `json:"remove,omitempty"`
}

// ContainerCreateResponse is the Podman Libpod REST response for the
// container create endpoint (201 Created).
type ContainerCreateResponse struct {
	ID       string   `json:"Id"`
	Warnings []string `json:"Warnings"`
}
