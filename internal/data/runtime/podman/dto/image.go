package dto

// ImagePruneReport is the Podman Libpod REST response item for image prune
// operations.
type ImagePruneReport struct {
	ID   string `json:"Id"`
	Size uint64 `json:"Size"`
	Err  string `json:"Err"`
}

// ImageLoadReport is the Podman Libpod REST response for image load operations.
type ImageLoadReport struct {
	Names []string `json:"Names"`
}
