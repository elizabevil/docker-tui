package dto

// ContainerTopResponse is the Podman Libpod REST response for the container
// top endpoint.
type ContainerTopResponse struct {
	Titles    []string   `json:"Titles"`
	Processes [][]string `json:"Processes"`
}
