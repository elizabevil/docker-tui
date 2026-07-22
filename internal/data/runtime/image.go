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

// ImageHistorySource tracks where history data came from.
type ImageHistorySource string

const (
	ImageHistoryPending  ImageHistorySource = "pending"
	ImageHistoryLayerAPI ImageHistorySource = "layer_api"
	ImageHistoryManifest ImageHistorySource = "manifest"
)

// ImageDetail is the full structured inspect model for an image.
type ImageDetail struct {
	ID               string
	RepoTags         []string
	RepoDigests      []string
	Registry         string
	Name             string
	Tag              string
	Created          string
	Size             int64
	Architecture     string
	OS               string
	OSVersion        string
	Author           string
	Comment          string
	Driver           string
	LayerCount       int
	Runtime          ImageRuntimeConfig
	Labels           map[string]string
	IsManifest       bool
	ManifestVariants []ImageManifestEntry
	History          []ImageHistoryLayer
	HistorySource    ImageHistorySource
	HistoryError     string
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
