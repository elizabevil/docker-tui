package podman

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

type ContainerInspectState struct {
	Status     string `json:"Status"`
	Pid        int    `json:"Pid"`
	StartedAt  string `json:"StartedAt"`
	FinishedAt string `json:"FinishedAt"`
}

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

type ContainerInspectHostConfig struct {
	CPUShares     int64                          `json:"CpuShares"`
	Memory        int64                          `json:"Memory"`
	NanoCPUs      int64                          `json:"NanoCpus"`
	NetworkMode   string                         `json:"NetworkMode"`
	RestartPolicy *ContainerInspectRestartPolicy `json:"RestartPolicy"`
}

type ContainerInspectRestartPolicy struct {
	Name              string `json:"Name"`
	MaximumRetryCount int    `json:"MaximumRetryCount"`
}

type ContainerInspectNetSettings struct {
	Networks map[string]ContainerInspectNetworkEntry  `json:"Networks"`
	Ports    map[string][]ContainerInspectPortBinding `json:"Ports"`
}

type ContainerInspectNetworkEntry struct {
	IPAddress  string `json:"IPAddress"`
	Gateway    string `json:"Gateway"`
	MacAddress string `json:"MacAddress"`
}

type ContainerInspectPortBinding struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type ContainerInspectMount struct {
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Mode        string `json:"Mode"`
	RW          bool   `json:"RW"`
}
