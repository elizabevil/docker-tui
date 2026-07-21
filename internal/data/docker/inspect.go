package docker

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"

	"github.com/elizabevil/docker-tui/internal/data/i18n"
)

// ── Named types for Docker inspect JSON unmarshalling ──────────

// inspectContainer represents the JSON returned by docker container inspect.
type inspectContainer struct {
	ID         string                    `json:"Id"`
	Name       string                    `json:"Name"`
	Created    string                    `json:"Created"`
	State      *inspectContainerState    `json:"State"`
	Platform   string                    `json:"Platform"`
	Config     *inspectContainerConfig   `json:"Config"`
	HostConfig *inspectHostConfig        `json:"HostConfig"`
	NetworkSettings *inspectNetSettings  `json:"NetworkSettings"`
	Mounts     []inspectMount            `json:"Mounts"`
	RestartCount int                     `json:"RestartCount"`
}

// inspectContainerState holds the running state of a container.
type inspectContainerState struct {
	Status     string `json:"Status"`
	Pid        int    `json:"Pid"`
	StartedAt  string `json:"StartedAt"`
	FinishedAt string `json:"FinishedAt"`
}

// inspectContainerConfig holds the container configuration.
type inspectContainerConfig struct {
	Image        string              `json:"Image"`
	WorkingDir   string              `json:"WorkingDir"`
	User         string              `json:"User"`
	Entrypoint   []string            `json:"Entrypoint"`
	Cmd          []string            `json:"Cmd"`
	Env          []string            `json:"Env"`
	ExposedPorts map[string]struct{} `json:"ExposedPorts"`
	Labels       map[string]string   `json:"Labels"`
}

// inspectHostConfig holds resource limits and host-level settings.
type inspectHostConfig struct {
	CPUShares     int64                  `json:"CpuShares"`
	Memory        int64                  `json:"Memory"`
	NanoCPUs      int64                  `json:"NanoCpus"`
	NetworkMode   string                 `json:"NetworkMode"`
	RestartPolicy *inspectRestartPolicy  `json:"RestartPolicy"`
}

// inspectRestartPolicy holds the container restart policy.
type inspectRestartPolicy struct {
	Name              string `json:"Name"`
	MaximumRetryCount int    `json:"MaximumRetryCount"`
}

// inspectNetSettings holds network configuration for a container.
type inspectNetSettings struct {
	Networks map[string]inspectNetworkEntry `json:"Networks"`
	Ports    map[string][]inspectPortBinding `json:"Ports"`
}

// inspectNetworkEntry holds a single network attachment.
type inspectNetworkEntry struct {
	IPAddress  string `json:"IPAddress"`
	Gateway    string `json:"Gateway"`
	MacAddress string `json:"MacAddress"`
}

// inspectPortBinding holds a host-to-container port mapping.
type inspectPortBinding struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

// inspectMount holds a volume or bind mount definition.
type inspectMount struct {
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Mode        string `json:"Mode"`
	RW          bool   `json:"RW"`
}

// inspectNetwork represents the JSON returned by docker network inspect.
type inspectNetwork struct {
	Name       string                            `json:"Name"`
	ID         string                            `json:"Id"`
	Created    string                            `json:"Created"`
	Scope      string                            `json:"Scope"`
	Driver     string                            `json:"Driver"`
	EnableIPv6 bool                              `json:"EnableIPv6"`
	Internal   bool                              `json:"Internal"`
	IPAM       inspectNetworkIPAM                `json:"IPAM"`
	Containers map[string]inspectNetworkContainer `json:"Containers"`
	Labels     map[string]string                 `json:"Labels"`
}

// inspectNetworkIPAM holds IP Address Management configuration.
type inspectNetworkIPAM struct {
	Driver string                   `json:"Driver"`
	Config []inspectNetworkIPAMConf `json:"Config"`
}

// inspectNetworkIPAMConf holds a single IPAM subnet configuration.
type inspectNetworkIPAMConf struct {
	Subnet  string `json:"Subnet"`
	Gateway string `json:"Gateway"`
}

// inspectNetworkContainer holds info about a container attached to a network.
type inspectNetworkContainer struct {
	Name        string `json:"Name"`
	EndpointID  string `json:"EndpointID"`
	MacAddress  string `json:"MacAddress"`
	IPv4Address string `json:"IPv4Address"`
	IPv6Address string `json:"IPv6Address"`
}

// inspectVolume represents the JSON returned by docker volume inspect.
type inspectVolume struct {
	Name       string                 `json:"Name"`
	Driver     string                 `json:"Driver"`
	Mountpoint string                 `json:"Mountpoint"`
	Scope      string                 `json:"Scope"`
	CreatedAt  string                 `json:"CreatedAt"`
	Labels     map[string]string      `json:"Labels"`
	Options    map[string]string      `json:"Options"`
	Status     map[string]interface{} `json:"Status"`
}

// ── Public API ────────────────────────────────────────────────

// InspectContainer returns the raw JSON from docker container inspect.
func (c *Client) InspectContainer(id string) ([]byte, error) {
	_, raw, err := c.cli.ContainerInspectWithRaw(context.Background(), id, false)
	if err != nil {
		return nil, fmt.Errorf("inspect %s: %w", id, err)
	}
	return raw, nil
}

// DetailSection is a section in the detail view.
type DetailSection struct {
	Title    string
	Subtitle string
	Lines    []string
}

// BuildContainerDetailSections parses raw JSON from docker container inspect
// and returns sections with i18n-translated labels for the detail view.
func BuildContainerDetailSections(jsonData []byte) []DetailSection {
	if len(jsonData) == 0 {
		return nil
	}

	var info inspectContainer
	if err := sonicUnmarshal(jsonData, &info); err != nil {
		return nil
	}

	sections := make([]DetailSection, 0, 7)

	// ── Basic Info ──
	basic := DetailSection{Title: i18n.T("inspect.section_basic")}
	shortID := info.ID
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}
	appendValue(&basic.Lines, "ID", shortID)
	appendValue(&basic.Lines, "Name", info.Name)
	if info.Config != nil {
		appendValue(&basic.Lines, "Image", info.Config.Image)
	}
	appendValue(&basic.Lines, i18n.T("inspect.created"), info.Created)
	if info.State != nil {
		appendValue(&basic.Lines, "State", info.State.Status)
		appendValue(&basic.Lines, "Pid", fmt.Sprintf("%d", info.State.Pid))
		if info.State.StartedAt != "" {
			appendValue(&basic.Lines, i18n.T("inspect.container.started_at"), info.State.StartedAt)
		}
		if info.State.FinishedAt != "" && info.State.FinishedAt != "0001-01-01T00:00:00Z" {
			appendValue(&basic.Lines, i18n.T("inspect.container.finished_at"), info.State.FinishedAt)
		}
		appendValue(&basic.Lines, i18n.T("inspect.container.restart_count"), fmt.Sprintf("%d", info.RestartCount))
	}
	appendValue(&basic.Lines, "Platform", info.Platform)
	sections = append(sections, basic)

	// ── Resources ──
	if info.HostConfig != nil {
		resources := DetailSection{Title: i18n.T("inspect.section_resources")}
		if info.HostConfig.CPUShares > 0 {
			appendValue(&resources.Lines, i18n.T("inspect.container.cpu_shares"), fmt.Sprintf("%d", info.HostConfig.CPUShares))
		}
		if info.HostConfig.Memory > 0 {
			appendValue(&resources.Lines, i18n.T("inspect.container.memory"), fmt.Sprintf("%d bytes", info.HostConfig.Memory))
		}
		if info.HostConfig.NanoCPUs > 0 {
			appendValue(&resources.Lines, i18n.T("inspect.container.nano_cpus"), fmt.Sprintf("%d", info.HostConfig.NanoCPUs))
		}
		appendValue(&resources.Lines, i18n.T("inspect.container.network_mode"), info.HostConfig.NetworkMode)
		if info.HostConfig.RestartPolicy != nil && info.HostConfig.RestartPolicy.Name != "" {
			appendValue(&resources.Lines, i18n.T("inspect.container.restart_policy"),
				fmt.Sprintf("%s (max %d)", info.HostConfig.RestartPolicy.Name, info.HostConfig.RestartPolicy.MaximumRetryCount))
		}
		if len(resources.Lines) > 0 {
			sections = append(sections, resources)
		}
	}

	// ── Networks ──
	if info.NetworkSettings != nil && len(info.NetworkSettings.Networks) > 0 {
		networks := DetailSection{Title: i18n.T("inspect.section_networks")}
		for name, net := range info.NetworkSettings.Networks {
			networks.Lines = append(networks.Lines, "  "+name+":")
			appendValue(&networks.Lines, "    "+i18n.T("inspect.container.ip"), net.IPAddress)
			appendValue(&networks.Lines, "    "+i18n.T("inspect.container.gateway"), net.Gateway)
			appendValue(&networks.Lines, "    "+i18n.T("inspect.container.mac"), net.MacAddress)
		}
		if info.NetworkSettings.Ports != nil && len(info.NetworkSettings.Ports) > 0 {
			networks.Lines = append(networks.Lines, "  "+i18n.T("inspect.container.ports")+":")
			for port, bindings := range info.NetworkSettings.Ports {
				for _, b := range bindings {
					networks.Lines = append(networks.Lines, fmt.Sprintf("    %s:%s -> %s", b.HostIP, b.HostPort, port))
				}
			}
		}
		sections = append(sections, networks)
	}

	// ── Mounts ──
	if len(info.Mounts) > 0 {
		mounts := DetailSection{Title: i18n.T("inspect.section_mounts")}
		for _, m := range info.Mounts {
			line := fmt.Sprintf("  %s -> %s", m.Source, m.Destination)
			if m.Mode != "" {
				line += fmt.Sprintf(" (%s)", m.Mode)
			}
			if m.RW {
				line += " [rw]"
			} else {
				line += " [ro]"
			}
			mounts.Lines = append(mounts.Lines, line)
		}
		sections = append(sections, mounts)
	}

	// ── Config ──
	if info.Config != nil {
		config := DetailSection{Title: i18n.T("inspect.section_container_config")}
		appendValue(&config.Lines, i18n.T("inspect.container.working_dir"), info.Config.WorkingDir)
		appendValue(&config.Lines, i18n.T("inspect.container.user"), info.Config.User)
		if len(info.Config.Entrypoint) > 0 {
			appendValue(&config.Lines, i18n.T("inspect.container.entrypoint"), fmt.Sprintf("[%s]", joinStrings(info.Config.Entrypoint)))
		}
		if len(info.Config.Cmd) > 0 {
			appendValue(&config.Lines, i18n.T("inspect.container.cmd"), fmt.Sprintf("[%s]", joinStrings(info.Config.Cmd)))
		}
		if len(info.Config.Env) > 0 {
			config.Lines = append(config.Lines, i18n.T("inspect.container.env")+fmt.Sprintf(" (%d vars):", len(info.Config.Env)))
			for _, env := range info.Config.Env {
				config.Lines = append(config.Lines, "  "+env)
			}
		}
		if len(info.Config.ExposedPorts) > 0 {
			ports := make([]string, 0, len(info.Config.ExposedPorts))
			for p := range info.Config.ExposedPorts {
				ports = append(ports, string(p))
			}
			appendValue(&config.Lines, i18n.T("inspect.container.exposed_ports"), joinStrings(ports))
		}
		if len(config.Lines) > 0 {
			sections = append(sections, config)
		}
	}

	// ── Labels ──
	if info.Config != nil && len(info.Config.Labels) > 0 {
		labels := DetailSection{Title: i18n.T("inspect.section_labels")}
		for k, v := range info.Config.Labels {
			labels.Lines = append(labels.Lines, fmt.Sprintf("  %s=%s", k, v))
		}
		sections = append(sections, labels)
	}

	return sections
}

// BuildNetworkDetailSections parses raw JSON from docker network inspect
// and returns sections with i18n-translated labels for the detail view.
func BuildNetworkDetailSections(jsonData []byte) []DetailSection {
	if len(jsonData) == 0 {
		return nil
	}

	var info inspectNetwork
	if err := sonicUnmarshal(jsonData, &info); err != nil {
		return nil
	}

	sections := make([]DetailSection, 0, 5)

	// ── Network Info ──
	netInfo := DetailSection{Title: i18n.T("inspect.section_network_info")}
	shortID := info.ID
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}
	appendValue(&netInfo.Lines, i18n.T("inspect.network.name"), info.Name)
	appendValue(&netInfo.Lines, "ID", shortID)
	appendValue(&netInfo.Lines, i18n.T("inspect.network.driver"), info.Driver)
	appendValue(&netInfo.Lines, i18n.T("inspect.network.scope"), info.Scope)
	appendValue(&netInfo.Lines, i18n.T("inspect.network.created"), info.Created)
	if info.Internal {
		appendValue(&netInfo.Lines, i18n.T("inspect.network.internal"), "true")
	}
	if info.EnableIPv6 {
		appendValue(&netInfo.Lines, i18n.T("inspect.network.ipv6"), "true")
	}
	sections = append(sections, netInfo)

	// ── IP Configuration ──
	if len(info.IPAM.Config) > 0 {
		ipam := DetailSection{Title: i18n.T("inspect.section_network_ipam")}
		for _, cfg := range info.IPAM.Config {
			appendValue(&ipam.Lines, "    "+i18n.T("inspect.network.subnet"), cfg.Subnet)
			appendValue(&ipam.Lines, "    "+i18n.T("inspect.network.gateway"), cfg.Gateway)
		}
		sections = append(sections, ipam)
	}

	// ── Connected Containers ──
	if len(info.Containers) > 0 {
		containers := DetailSection{Title: i18n.T("inspect.section_network_containers")}
		for id, ctr := range info.Containers {
			shortCtrID := id
			if len(shortCtrID) > 12 {
				shortCtrID = shortCtrID[:12]
			}
			containers.Lines = append(containers.Lines, fmt.Sprintf("  %s (%s):", ctr.Name, shortCtrID))
			appendValue(&containers.Lines, "    IPv4", ctr.IPv4Address)
			appendValue(&containers.Lines, "    IPv6", ctr.IPv6Address)
			appendValue(&containers.Lines, "    MAC", ctr.MacAddress)
		}
		sections = append(sections, containers)
	}

	// ── Labels ──
	if len(info.Labels) > 0 {
		labels := DetailSection{Title: i18n.T("inspect.section_labels")}
		for k, v := range info.Labels {
			labels.Lines = append(labels.Lines, fmt.Sprintf("  %s=%s", k, v))
		}
		sections = append(sections, labels)
	}

	return sections
}

// BuildVolumeDetailSections parses raw JSON from docker volume inspect
// and returns sections with i18n-translated labels for the detail view.
func BuildVolumeDetailSections(jsonData []byte) []DetailSection {
	if len(jsonData) == 0 {
		return nil
	}

	var info inspectVolume
	if err := sonicUnmarshal(jsonData, &info); err != nil {
		return nil
	}

	sections := make([]DetailSection, 0, 4)

	// ── Volume Info ──
	volInfo := DetailSection{Title: i18n.T("inspect.section_volume_info")}
	appendValue(&volInfo.Lines, i18n.T("inspect.volume.name"), info.Name)
	appendValue(&volInfo.Lines, i18n.T("inspect.volume.driver"), info.Driver)
	appendValue(&volInfo.Lines, i18n.T("inspect.volume.mountpoint"), info.Mountpoint)
	appendValue(&volInfo.Lines, i18n.T("inspect.volume.scope"), info.Scope)
	appendValue(&volInfo.Lines, i18n.T("inspect.volume.created"), info.CreatedAt)
	sections = append(sections, volInfo)

	// ── Options ──
	if len(info.Options) > 0 {
		opts := DetailSection{Title: i18n.T("inspect.section_volume_options")}
		for k, v := range info.Options {
			opts.Lines = append(opts.Lines, fmt.Sprintf("  %s=%s", k, v))
		}
		sections = append(sections, opts)
	}

	// ── Labels ──
	if len(info.Labels) > 0 {
		labels := DetailSection{Title: i18n.T("inspect.section_labels")}
		for k, v := range info.Labels {
			labels.Lines = append(labels.Lines, fmt.Sprintf("  %s=%s", k, v))
		}
		sections = append(sections, labels)
	}

	return sections
}

// ── helpers ───────────────────────────────────────────────────

func appendValue(lines *[]string, key, value string) {
	if value != "" {
		*lines = append(*lines, key+": "+value)
	}
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

func sonicUnmarshal(data []byte, v interface{}) error {
	return sonic.Unmarshal(data, v)
}
