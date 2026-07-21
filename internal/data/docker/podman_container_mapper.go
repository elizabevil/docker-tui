package docker

import "time"

type podmanContainerSummary struct {
	ID      string            `json:"Id"`
	Names   []string          `json:"Names"`
	Image   string            `json:"Image"`
	Status  string            `json:"Status"`
	State   string            `json:"State"`
	Created time.Time         `json:"Created"`
	Ports   []podmanPort      `json:"Ports"`
	Mounts  []string          `json:"Mounts"`
	Labels  map[string]string `json:"Labels"`
}

type podmanPort struct {
	ContainerPort uint16 `json:"container_port"`
	HostPort      uint16 `json:"host_port"`
	Range         uint16 `json:"range"`
	Protocol      string `json:"protocol"`
	HostIP        string `json:"host_ip"`
}

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
			ID:           shortContainerID(container.ID),
			Name:         name,
			Image:        container.Image,
			Status:       container.Status,
			State:        container.State,
			Created:      container.Created.Unix(),
			PortBindings: ports,
			MountCount:   len(container.Mounts),
			Labels:       container.Labels,
		}
		summary.ComposeProject = container.Labels["com.docker.compose.project"]
		summary.ComposeService = container.Labels["com.docker.compose.service"]
		result = append(result, summary)
	}
	return result
}

func shortContainerID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}
