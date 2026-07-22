package docker

import "time"

// podmanContainerSummary is the adapter's stable representation of the Libpod
// container list response. Both CGO bindings and non-CGO REST normalize into
// this type before mapping to ContainerSummary.
type podmanContainerSummary struct {
	ID       string            `json:"Id"`
	Names    []string          `json:"Names"`
	Image    string            `json:"Image"`
	Status   string            `json:"Status"`
	State    string            `json:"State"`
	Created  time.Time         `json:"Created"`
	Ports    []podmanPort      `json:"Ports"`
	Mounts   []string          `json:"Mounts"`
	Networks []string          `json:"Networks"`
	Labels   map[string]string `json:"Labels"`
}

// podmanPort is a single port mapping entry in the Podman container list response.
type podmanPort struct {
	ContainerPort uint16 `json:"container_port"`
	HostPort      uint16 `json:"host_port"`
	Range         uint16 `json:"range"`
	Protocol      string `json:"protocol"`
	HostIP        string `json:"host_ip"`
}

// mapPodmanContainerSummaries converts Podman container list results into
// ContainerSummary items. Field differences: Created (time.Time) is converted
// to Unix timestamp; Names is a slice where only the first entry is used;
// Ports are expanded from range entries into individual PortBinding items.
func mapPodmanContainerSummaries(raw []podmanContainerSummary) []ContainerSummary {
	result := make([]ContainerSummary, 0, len(raw))
	for _, container := range raw {
		name := ""
		if len(container.Names) > 0 {
			name = container.Names[0]
		}
		ports := make([]PortBinding, 0, len(container.Ports))
		for _, port := range container.Ports {
			portRange := port.Range
			if portRange == 0 {
				portRange = 1
			}
			for offset := uint16(0); offset < portRange; offset++ {
				ports = append(ports, PortBinding{
					ContainerPort: port.ContainerPort + offset,
					HostPort:      port.HostPort + offset,
					Protocol:      port.Protocol,
					HostIP:        port.HostIP,
				})
			}
		}
		summary := ContainerSummary{
			ID:           container.ID,
			Name:         name,
			Image:        container.Image,
			Status:       container.Status,
			State:        container.State,
			Created:      container.Created.Unix(),
			PortBindings: ports,
			MountCount:   len(container.Mounts),
			MountNames:   append([]string(nil), container.Mounts...),
			NetworkNames: append([]string(nil), container.Networks...),
			Labels:       container.Labels,
		}
		summary.ComposeProject = container.Labels["com.docker.compose.project"]
		summary.ComposeService = container.Labels["com.docker.compose.service"]
		result = append(result, summary)
	}
	return result
}
