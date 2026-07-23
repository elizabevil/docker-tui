package podman

import (
	"net"
	"sort"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

func MapContainerSummaries(raw []dto.ContainerItem) []runtimeapi.ContainerSummary {
	result := make([]runtimeapi.ContainerSummary, 0, len(raw))
	for _, c := range raw {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
		}
		ports := make([]runtimeapi.PortBinding, 0, len(c.Ports))
		for _, port := range c.Ports {
			portRange := port.Range
			if portRange == 0 {
				portRange = 1
			}
			for offset := uint16(0); offset < portRange; offset++ {
				ports = append(ports, runtimeapi.PortBinding{
					ContainerPort: port.ContainerPort + offset,
					HostPort:      port.HostPort + offset,
					Protocol:      port.Protocol,
					HostIP:        port.HostIP,
				})
			}
		}
		summary := runtimeapi.ContainerSummary{
			ID:           c.ID,
			Name:         name,
			Image:        c.Image,
			Status:       c.Status,
			State:        c.State,
			Created:      c.Created.Unix(),
			PortBindings: ports,
			MountCount:   len(c.Mounts),
			MountNames:   append([]string(nil), c.Mounts...),
			NetworkNames: append([]string(nil), c.Networks...),
			Labels:       c.Labels,
		}
		summary.ComposeProject = c.Labels["com.docker.compose.project"]
		summary.ComposeService = c.Labels["com.docker.compose.service"]
		result = append(result, summary)
	}
	return result
}

func MapImageSummaries(raw []dto.ImageItem) []runtimeapi.ImageSummary {
	out := make([]runtimeapi.ImageSummary, 0, len(raw))
	for _, image := range raw {
		summary := runtimeapi.ImageSummary{
			ID:       image.ID,
			RepoTags: image.RepoTags,
			Created:  image.Created,
			Size:     image.Size,
			Labels:   image.Labels,
			Arch:     image.Arch,
		}
		if summary.Arch == "" {
			summary.Arch = "\u2014"
		}
		summary.IsManifest = image.IsManifestList != nil && *image.IsManifestList
		summary.Registry, _, _ = runtimeapi.SplitImageRef(summary.RepoTags)
		out = append(out, summary)
	}
	return out
}

func MapVolumes(raw []dto.VolumeItem) []runtimeapi.Volume {
	result := make([]runtimeapi.Volume, 0, len(raw))
	for _, v := range raw {
		scope := v.Scope
		if scope == "" {
			scope = "local"
		}
		result = append(result, runtimeapi.Volume{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			Labels:     v.Labels,
			Scope:      scope,
			CreatedAt:  dto.FormatPodmanTime(v.CreatedAt),
		})
	}
	return result
}

func MapVolumeInspect(raw dto.VolumeItem) *runtimeapi.VolumeDetail {
	scope := raw.Scope
	if scope == "" {
		scope = "local"
	}
	return &runtimeapi.VolumeDetail{
		Name:       raw.Name,
		Driver:     raw.Driver,
		Mountpoint: raw.Mountpoint,
		CreatedAt:  dto.FormatPodmanTime(raw.CreatedAt),
		Labels:     raw.Labels,
		Scope:      scope,
		Options:    raw.Options,
		Status:     raw.Status,
	}
}

func MapNetworks(raw []dto.Network) []runtimeapi.Network {
	result := make([]runtimeapi.Network, 0, len(raw))
	for _, n := range raw {
		ipam := make([]string, 0, len(n.Subnets))
		for _, s := range n.Subnets {
			cidr := s.Subnet.String()
			if cidr != "" {
				ipam = append(ipam, cidr)
			}
		}
		result = append(result, runtimeapi.Network{
			Name:     n.Name,
			ID:       n.ID,
			Driver:   n.Driver,
			Scope:    "local",
			IPAM:     ipam,
			Created:  n.Created.Unix(),
			Internal: n.Internal,
			Labels:   n.Labels,
		})
	}
	return result
}

func MapNetworkInspect(raw dto.NetworkInspect) *runtimeapi.NetworkDetail {
	detailContainers := make(map[string]runtimeapi.NetworkEndpoint, len(raw.Containers))
	for id, c := range raw.Containers {
		ep := runtimeapi.NetworkEndpoint{Name: c.Name}
		for _, iface := range c.Interfaces {
			ep.MacAddress = iface.MacAddress.String()
			if len(iface.Subnets) > 0 {
				ep.IPv4Address = iface.Subnets[0].IPNet.String()
				ep.Gateway = netIPString(iface.Subnets[0].Gateway)
			}
			break
		}
		detailContainers[id] = ep
	}

	ipamConfigs := make([]runtimeapi.NetworkIPAMConfig, 0, len(raw.Subnets))
	enableIPv4 := false
	enableIPv6 := raw.IPv6Enabled
	for _, s := range raw.Subnets {
		ipamConfigs = append(ipamConfigs, runtimeapi.NetworkIPAMConfig{
			Subnet:  s.Subnet.String(),
			Gateway: netIPString(s.Gateway),
		})
		if ip, _, err := net.ParseCIDR(s.Subnet.String()); err == nil {
			if ip.To4() != nil {
				enableIPv4 = true
			} else {
				enableIPv6 = true
			}
		}
	}

	return &runtimeapi.NetworkDetail{
		Name:       raw.Name,
		ID:         raw.ID,
		Created:    raw.Created.Format(time.RFC3339),
		Scope:      "local",
		Driver:     raw.Driver,
		EnableIPv4: enableIPv4,
		EnableIPv6: enableIPv6,
		IPAM: runtimeapi.NetworkIPAM{
			Config: ipamConfigs,
		},
		Internal:   raw.Internal,
		Containers: detailContainers,
		Options:    raw.Options,
		Labels:     raw.Labels,
	}
}

func netIPString(ip net.IP) string {
	if ip == nil {
		return ""
	}
	return ip.String()
}

func MapContainerInspectResponse(info dto.ContainerInspectJSON) *runtimeapi.ContainerDetail {
	detail := &runtimeapi.ContainerDetail{
		ID: info.ID, Name: info.Name, Created: info.Created, Platform: info.Platform,
		RestartCount: info.RestartCount, Networks: map[string]runtimeapi.ContainerNetwork{},
		Ports: map[string][]runtimeapi.ContainerPortBinding{},
	}
	if info.State != nil {
		detail.State = runtimeapi.ContainerState{Status: info.State.Status, PID: info.State.Pid, StartedAt: info.State.StartedAt, FinishedAt: info.State.FinishedAt}
	}
	if info.Config != nil {
		detail.Image = info.Config.Image
		detail.Config = runtimeapi.ContainerConfig{WorkingDir: info.Config.WorkingDir, User: info.Config.User, Entrypoint: info.Config.Entrypoint, Command: info.Config.Cmd, Environment: info.Config.Env, Labels: info.Config.Labels}
		for port := range info.Config.ExposedPorts {
			detail.Config.ExposedPorts = append(detail.Config.ExposedPorts, port)
		}
	}
	if info.HostConfig != nil {
		detail.Resources = runtimeapi.ContainerResources{CPUShares: info.HostConfig.CPUShares, Memory: info.HostConfig.Memory, NanoCPUs: info.HostConfig.NanoCPUs, NetworkMode: info.HostConfig.NetworkMode}
		if info.HostConfig.RestartPolicy != nil {
			detail.Resources.RestartPolicy = info.HostConfig.RestartPolicy.Name
			detail.Resources.MaximumRetryCount = info.HostConfig.RestartPolicy.MaximumRetryCount
		}
	}
	if info.NetworkSettings != nil {
		for name, net := range info.NetworkSettings.Networks {
			detail.Networks[name] = runtimeapi.ContainerNetwork{IPAddress: net.IPAddress, Gateway: net.Gateway, MACAddress: net.MacAddress}
		}
		for port, bindings := range info.NetworkSettings.Ports {
			mapped := make([]runtimeapi.ContainerPortBinding, 0, len(bindings))
			for _, b := range bindings {
				mapped = append(mapped, runtimeapi.ContainerPortBinding{HostIP: b.HostIP, HostPort: b.HostPort})
			}
			detail.Ports[port] = mapped
		}
	}
	for _, m := range info.Mounts {
		detail.Mounts = append(detail.Mounts, runtimeapi.ContainerMount{Source: m.Source, Destination: m.Destination, Mode: m.Mode, ReadWrite: m.RW})
	}
	sort.Slice(detail.Config.ExposedPorts, func(i, j int) bool { return detail.Config.ExposedPorts[i] < detail.Config.ExposedPorts[j] })
	return detail
}

// MapImageInspect converts a Podman image inspect response to the runtime's
// ImageDetail. Fields absent from the Podman response are left at their zero
// values so the UI can show "—" without checking for nil on every render.
func MapImageInspect(inspect dto.ImageInspectJSON, history []dto.LayerHistoryEntry) *runtimeapi.ImageDetail {
	detail := &runtimeapi.ImageDetail{
		ID:           inspect.ID,
		RepoTags:     append([]string(nil), inspect.RepoTags...),
		RepoDigests:  append([]string(nil), inspect.RepoDigests...),
		Created:      inspect.Created,
		Size:         inspect.Size,
		Architecture: inspect.Architecture,
		OS:           inspect.Os,
		OSVersion:    inspect.OsVersion,
		Author:       inspect.Author,
		Comment:      inspect.Comment,
		Driver:       inspect.GraphDriver.Name,
		LayerCount:   len(inspect.RootFS.Layers),
	}
	detail.Registry, detail.Name, detail.Tag = runtimeapi.SplitImageRef(detail.RepoTags)
	if detail.Architecture == "" {
		detail.Architecture = "—"
	}
	if detail.OS == "" {
		detail.OS = "—"
	}

	if cfg := inspect.Config; cfg != nil {
		detail.Runtime = runtimeapi.ImageRuntimeConfig{
			WorkingDir:  cfg.WorkingDir,
			User:        cfg.User,
			StopSignal:  cfg.StopSignal,
			Entrypoint:  append([]string(nil), cfg.Entrypoint...),
			Cmd:         append([]string(nil), cfg.Cmd...),
			Shell:       append([]string(nil), cfg.Shell...),
			OnBuild:     append([]string(nil), cfg.OnBuild...),
			Environment: append([]string(nil), cfg.Env...),
		}
		for port := range cfg.ExposedPorts {
			detail.Runtime.ExposedPorts = append(detail.Runtime.ExposedPorts, port)
		}
		for volume := range cfg.Volumes {
			detail.Runtime.Volumes = append(detail.Runtime.Volumes, volume)
		}
		sort.Strings(detail.Runtime.ExposedPorts)
		sort.Strings(detail.Runtime.Volumes)
		detail.Labels = cloneStringMapForImage(cfg.Labels)
	}

	if len(detail.Labels) == 0 {
		detail.Labels = map[string]string{}
	}
	detail.HistorySource = runtimeapi.ImageHistoryLayerAPI
	if len(history) > 0 {
		detail.History = make([]runtimeapi.ImageHistoryLayer, 0, len(history))
		for _, h := range history {
			detail.History = append(detail.History, runtimeapi.ImageHistoryLayer{
				ID: h.ID, Created: h.Created, CreatedBy: h.CreatedBy, Size: h.Size, Comment: h.Comment,
			})
		}
	}
	return detail
}

func cloneStringMapForImage(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	out := make(map[string]string, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}
