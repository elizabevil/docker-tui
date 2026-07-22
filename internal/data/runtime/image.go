// Package runtime defines the domain types and service interfaces shared by
// all container runtime adapters (Docker, Podman). The UI and state layers
// depend only on these types; runtime-specific wire DTOs stay inside the
// adapter packages and are never leaked upward.
package runtime

import "context"

// ManifestPlatform holds platform info for a single manifest entry.
type ManifestPlatform struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Variant      string `json:"variant,omitempty"`
}

// ImageManifestEntry holds a simplified view of one manifest in a manifest list.
type ImageManifestEntry struct {
	Digest    string           `json:"digest"`
	Platform  ManifestPlatform `json:"platform"`
	Size      int64            `json:"size"`
	Available bool             `json:"available"`
}

// ImageSummary is the lightweight model returned by list operations.
type ImageSummary struct {
	ID         string
	RepoTags   []string
	Created    int64
	Size       int64
	Labels     map[string]string
	Registry   string
	Arch       string
	IsManifest bool
	Manifests  []ImageManifestEntry
}

// ImageRuntimeConfig holds container runtime configuration embedded in an image.
type ImageRuntimeConfig struct {
	WorkingDir   string
	User         string
	StopSignal   string
	Entrypoint   []string
	Cmd          []string
	Shell        []string
	OnBuild      []string
	Environment  []string
	ExposedPorts []string
	Volumes      []string
	Healthcheck  []string
}

// ImageHistoryLayer represents one layer in image build history.
type ImageHistoryLayer struct {
	ID        string
	Created   int64
	CreatedBy string
	Size      int64
	Comment   string
}

// ImageHistorySource tracks where history data came from. The UI uses this to
// decide whether to show layers, manifest variants, a pending spinner, or an
// error message.
type ImageHistorySource string

const (
	// ImageHistoryPending indicates history has not been fetched yet.
	ImageHistoryPending ImageHistorySource = "pending"
	// ImageHistoryLayerAPI indicates history was populated from the image
	// history API (available for non-manifest images).
	ImageHistoryLayerAPI ImageHistorySource = "layer_api"
	// ImageHistoryManifest indicates the image is a manifest list; history
	// layers are not applicable and manifest variants are shown instead.
	ImageHistoryManifest ImageHistorySource = "manifest"
)

// ImageDetail is the full structured inspect model for an image. Both Docker
// and Podman adapters produce this type from their native inspect responses,
// ensuring the UI never parses raw SDK JSON. For manifest lists, History is
// empty and ManifestVariants carries per-platform entries.
type ImageDetail struct {
	ID               string
	RepoTags         []string
	RepoDigests      []string
	Registry         string            // derived from first RepoTag via splitImageRef
	Name             string            // derived from first RepoTag (without registry)
	Tag              string            // derived from first RepoTag (without name)
	Created          string            // RFC3339 timestamp from the runtime
	Size             int64
	Architecture     string
	OS               string
	OSVersion        string
	Author           string
	Comment          string
	Driver           string            // graph driver name (e.g. overlay2)
	LayerCount       int               // number of rootfs layers
	Runtime          ImageRuntimeConfig
	Labels           map[string]string
	IsManifest       bool              // true for multi-arch manifest lists
	ManifestVariants []ImageManifestEntry
	History          []ImageHistoryLayer
	HistorySource    ImageHistorySource
	HistoryError     string            // non-empty when history fetch failed
}

// ImageListOptions carries filter and pagination options for listing images.
type ImageListOptions struct {
	All     bool
	Filters FilterSet
}

// ImageService defines runtime-neutral read operations for images. Mutations
// use ResourceActionService so all resource actions share one result contract.
type ImageService interface {
	List(context.Context, ImageListOptions) ([]ImageSummary, error)
	Inspect(context.Context, ImageSummary) (*ImageDetail, error)
}
