package dto

// ContainerListOptions mirrors Podman's containers.ListOptions.
// Used by the CGO driver and REST path.
type ContainerListOptions struct {
	All     bool
	Limit   int
	Filters map[string][]string
}

// ImageListOptions mirrors Podman's images.ListOptions.
type ImageListOptions struct {
	All     bool
	Filters map[string][]string
}

// NetworkListOptions mirrors Podman's network.ListOptions.
type NetworkListOptions struct {
	Filters map[string][]string
}

// VolumeListOptions mirrors Podman's volumes.ListOptions.
type VolumeListOptions struct {
	Filters map[string][]string
}
