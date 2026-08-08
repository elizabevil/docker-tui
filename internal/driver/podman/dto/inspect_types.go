package dto

// ContainerInspectJSON represents the Podman Libpod REST response for
// container inspect. The JSON structure is compatible with the Docker
// inspect schema.
type ContainerInspectJSON struct {
	ID              string                       `json:"Id"`
	Name            string                       `json:"Name"`
	Created         string                       `json:"Created"`
	State           *ContainerInspectState       `json:"State"`
	Platform        string                       `json:"Platform"`
	Config          *ContainerInspectConfig      `json:"Config"`
	HostConfig      *ContainerInspectHostConfig  `json:"HostConfig"`
	NetworkSettings *ContainerInspectNetSettings `json:"NetworkSettings"`
	Mounts          []ContainerInspectMount      `json:"Mounts"`
	RestartCount    int                          `json:"RestartCount"`
}

// ContainerInspectState holds the runtime state of a container.
type ContainerInspectState struct {
	Status     string `json:"Status"`
	Pid        int    `json:"Pid"`
	StartedAt  string `json:"StartedAt"`
	FinishedAt string `json:"FinishedAt"`
}

// ContainerInspectConfig holds the static configuration of a container.
type ContainerInspectConfig struct {
	Image        string              `json:"Image"`
	WorkingDir   string              `json:"WorkingDir"`
	User         string              `json:"User"`
	Entrypoint   []string            `json:"Entrypoint"`
	Cmd          []string            `json:"Cmd"`
	Env          []string            `json:"Env"`
	ExposedPorts map[string]struct{} `json:"ExposedPorts"`
	Labels       map[string]string   `json:"Labels"`
}

// ContainerInspectHostConfig holds the host-side configuration of a container.
type ContainerInspectHostConfig struct {
	CPUShares     int64                          `json:"CpuShares"`
	Memory        int64                          `json:"Memory"`
	NanoCPUs      int64                          `json:"NanoCpus"`
	NetworkMode   string                         `json:"NetworkMode"`
	RestartPolicy *ContainerInspectRestartPolicy `json:"RestartPolicy"`
}

// ContainerInspectRestartPolicy holds the restart policy configuration.
type ContainerInspectRestartPolicy struct {
	Name              string `json:"Name"`
	MaximumRetryCount int    `json:"MaximumRetryCount"`
}

// ContainerInspectNetSettings holds the network settings from inspect.
type ContainerInspectNetSettings struct {
	Networks map[string]ContainerInspectNetworkEntry  `json:"Networks"`
	Ports    map[string][]ContainerInspectPortBinding `json:"Ports"`
}

// ContainerInspectNetworkEntry holds network attachment info for a container.
type ContainerInspectNetworkEntry struct {
	IPAddress  string `json:"IPAddress"`
	Gateway    string `json:"Gateway"`
	MacAddress string `json:"MacAddress"`
}

// ContainerInspectPortBinding holds host-side port mapping.
type ContainerInspectPortBinding struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

// ContainerInspectMount describes a volume or bind mount attached to a container.
type ContainerInspectMount struct {
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Mode        string `json:"Mode"`
	RW          bool   `json:"RW"`
}

// ImageInspectConfig mirrors the Podman Libpod inspect response Config block.
// Only the fields surfaced via ImageDetail are decoded; optional fields use a
// pointer or map type so absent values don't get parsed as zero-valued
// strings or empty slices.
type ImageInspectConfig struct {
	Env          []string          `json:"Env,omitempty"`
	Cmd          []string          `json:"Cmd,omitempty"`
	Entrypoint   []string          `json:"Entrypoint,omitempty"`
	WorkingDir   string            `json:"WorkingDir,omitempty"`
	User         string            `json:"User,omitempty"`
	StopSignal   string            `json:"StopSignal,omitempty"`
	Shell        []string          `json:"Shell,omitempty"`
	OnBuild      []string          `json:"OnBuild,omitempty"`
	ExposedPorts map[string]any    `json:"ExposedPorts,omitempty"`
	Volumes      map[string]any    `json:"Volumes,omitempty"`
	Labels       map[string]string `json:"Labels,omitempty"`
}

// ImageInspectRootFS describes the rootfs section of a Podman image inspect.
type ImageInspectRootFS struct {
	Type   string   `json:"Type"`
	Layers []string `json:"Layers"`
}

// ImageInspectGraphDriver describes the storage driver backing the image.
type ImageInspectGraphDriver struct {
	Name string            `json:"Name"`
	Data map[string]string `json:"Data,omitempty"`
}

// ImageInspectJSON is the typed view of GET /libpod/images/{id}/json. It uses
// the same field shape as go.podman.io/podman/v6 inspect responses so a future
// cgo driver can reuse the same mapping without changes.
type ImageInspectJSON struct {
	ID           string                  `json:"Id"`
	Digest       string                  `json:"Digest,omitempty"`
	RepoTags     []string                `json:"RepoTags,omitempty"`
	RepoDigests  []string                `json:"RepoDigests,omitempty"`
	Parent       string                  `json:"Parent,omitempty"`
	Comment      string                  `json:"Comment,omitempty"`
	Created      string                  `json:"Created,omitempty"`
	Version      string                  `json:"Version,omitempty"`
	Author       string                  `json:"Author,omitempty"`
	Architecture string                  `json:"Architecture,omitempty"`
	Os           string                  `json:"Os,omitempty"`
	OsVersion    string                  `json:"OsVersion,omitempty"`
	Size         int64                   `json:"Size"`
	VirtualSize  int64                   `json:"VirtualSize,omitempty"`
	GraphDriver  ImageInspectGraphDriver `json:"GraphDriver"`
	RootFS       ImageInspectRootFS      `json:"RootFS"`
	Config       *ImageInspectConfig     `json:"Config,omitempty"`
}
