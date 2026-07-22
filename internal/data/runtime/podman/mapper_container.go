package podman

import runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"

// MapContainerSummaries converts Podman container list results into
// runtime.ContainerSummary items. Field differences: Created (time.Time) is
// converted to Unix timestamp; Names is a slice where only the first entry
// is used; Ports are expanded from range entries into individual PortBinding items.
func MapContainerSummaries(raw []ContainerItem) []runtimeapi.ContainerSummary {
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
