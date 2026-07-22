package podman

import (
	"sort"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// MapContainerInspectResponse converts a Podman container inspect response
// into the runtime-neutral ContainerDetail model.
func MapContainerInspectResponse(info ContainerInspectJSON) *runtimeapi.ContainerDetail {
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
