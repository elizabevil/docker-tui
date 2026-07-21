package docker

type ContainerListOptions struct {
	All     bool
	Limit   int
	Filters map[string][]string
}

type ContainerSummary struct {
	ID             string
	Name           string
	Image          string
	Status         string
	State          string
	Created        int64
	Ports          string
	IPs            []string
	MountCount     int
	Labels         map[string]string
	ComposeProject string
	ComposeService string
}

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

type ImageSummary struct {
	ID         string
	RepoTags   []string
	Created    int64
	Size       int64
	Labels     map[string]string
	Registry   string
	Arch       string
	IsManifest bool // true if image is a manifest list (multi-arch)
	Manifests  []ImageManifestEntry
}

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

type ImageHistoryLayer struct {
	ID        string
	Created   int64
	CreatedBy string
	Size      int64
	Comment   string
}

type ImageHistorySource string

const (
	ImageHistoryPending  ImageHistorySource = "pending"
	ImageHistoryLayerAPI ImageHistorySource = "layer_api"
	ImageHistoryManifest ImageHistorySource = "manifest"
)

type ImageDetailData struct {
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

type VolumeItem struct {
	Name       string
	Driver     string
	Mountpoint string
	Labels     map[string]string
	Scope      string
	CreatedAt  string
}

type NetworkItem struct {
	Name       string
	ID         string
	Driver     string
	Scope      string
	IPAM       []string
	Containers int
	Created    int64 // Unix timestamp
	Internal   bool
	Labels     map[string]string
}
