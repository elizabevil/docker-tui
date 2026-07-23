//go:build !cgo

package dto

// ImageItem is the Podman Libpod image list response item.
// Field shape and JSON tags match go.podman.io/podman/v6/pkg/domain/entities/types.ImageSummary.
type ImageItem struct {
	ID             string            `json:"Id"`
	ParentId       string            `json:"ParentId,omitempty"`
	RepoTags       []string          `json:"RepoTags,omitempty"`
	RepoDigests    []string          `json:"RepoDigests,omitempty"`
	Created        int64             `json:"Created"`
	Size           int64             `json:"Size"`
	SharedSize     int               `json:"SharedSize"`
	VirtualSize    int64             `json:"VirtualSize,omitempty"`
	Labels         map[string]string `json:"Labels,omitempty"`
	Containers     int               `json:"Containers"`
	ReadOnly       bool              `json:"ReadOnly,omitempty"`
	Dangling       bool              `json:"Dangling,omitempty"`
	Arch           string            `json:"Arch,omitempty"`
	Digest         string            `json:"Digest,omitempty"`
	History        []string          `json:"History,omitempty"`
	IsManifestList *bool             `json:"IsManifestList,omitempty"`
	Names          []string          `json:"Names,omitempty"`
	Os             string            `json:"Os,omitempty"`
}
