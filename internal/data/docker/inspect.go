package docker

import (
	"fmt"
	"strings"
)

func (c *Client) InspectContainer(id string) (string, error) {
	info, err := c.cli.ContainerInspect(c.ctx, id)
	if err != nil {
		return "", fmt.Errorf("inspect %s: %w", id, err)
	}
	var sb strings.Builder

	fmt.Fprintf(&sb, "ID: %s\n", info.ID[:12])
	fmt.Fprintf(&sb, "Name: %s\n", info.Name)
	fmt.Fprintf(&sb, "Image: %s\n", info.Config.Image)
	fmt.Fprintf(&sb, "Created: %s\n", info.Created)
	if info.State != nil {
		fmt.Fprintf(&sb, "State: %s\n", info.State.Status)
		fmt.Fprintf(&sb, "Pid: %d\n", info.State.Pid)
		if info.State.StartedAt != "" {
			fmt.Fprintf(&sb, "StartedAt: %s\n", info.State.StartedAt)
		}
		if info.State.FinishedAt != "" && info.State.FinishedAt != "0001-01-01T00:00:00Z" {
			fmt.Fprintf(&sb, "FinishedAt: %s\n", info.State.FinishedAt)
		}
		fmt.Fprintf(&sb, "RestartCount: %d\n", info.RestartCount)
	}
	fmt.Fprintf(&sb, "Platform: %s\n", info.Platform)

	if info.HostConfig != nil {
		fmt.Fprintf(&sb, "\n── Resources ──\n")
		if info.HostConfig.CPUShares > 0 {
			fmt.Fprintf(&sb, "CPUShares: %d\n", info.HostConfig.CPUShares)
		}
		if info.HostConfig.Memory > 0 {
			fmt.Fprintf(&sb, "Memory: %d bytes\n", info.HostConfig.Memory)
		}
		if info.HostConfig.NanoCPUs > 0 {
			fmt.Fprintf(&sb, "NanoCPUs: %d\n", info.HostConfig.NanoCPUs)
		}
		fmt.Fprintf(&sb, "NetworkMode: %s\n", info.HostConfig.NetworkMode)
		if info.HostConfig.RestartPolicy.Name != "" {
			fmt.Fprintf(&sb, "RestartPolicy: %s (max %d)\n",
				info.HostConfig.RestartPolicy.Name,
				info.HostConfig.RestartPolicy.MaximumRetryCount)
		}
	}

	if info.NetworkSettings != nil && len(info.NetworkSettings.Networks) > 0 {
		fmt.Fprintf(&sb, "\n── Networks ──\n")
		for name, net := range info.NetworkSettings.Networks {
			fmt.Fprintf(&sb, "  %s:\n", name)
			if net.IPAddress != "" {
				fmt.Fprintf(&sb, "    IP: %s\n", net.IPAddress)
			}
			if net.Gateway != "" {
				fmt.Fprintf(&sb, "    Gateway: %s\n", net.Gateway)
			}
			if net.MacAddress != "" {
				fmt.Fprintf(&sb, "    MAC: %s\n", net.MacAddress)
			}
		}
	}

	if info.NetworkSettings != nil {
		ports := info.NetworkSettings.Ports
		if len(ports) > 0 {
			fmt.Fprintf(&sb, "  Ports:\n")
			for port, bindings := range ports {
				for _, b := range bindings {
					fmt.Fprintf(&sb, "    %s:%s -> %s\n", b.HostIP, b.HostPort, port)
				}
			}
		}
	}

	if len(info.Mounts) > 0 {
		fmt.Fprintf(&sb, "\n── Mounts ──\n")
		for _, m := range info.Mounts {
			fmt.Fprintf(&sb, "  %s -> %s", m.Source, m.Destination)
			if m.Mode != "" {
				fmt.Fprintf(&sb, " (%s)", m.Mode)
			}
			if m.RW {
				fmt.Fprintf(&sb, " [rw]")
			} else {
				fmt.Fprintf(&sb, " [ro]")
			}
			fmt.Fprintf(&sb, "\n")
		}
	}

	if info.Config != nil {
		fmt.Fprintf(&sb, "\n── Config ──\n")
		if info.Config.WorkingDir != "" {
			fmt.Fprintf(&sb, "WorkingDir: %s\n", info.Config.WorkingDir)
		}
		if info.Config.User != "" {
			fmt.Fprintf(&sb, "User: %s\n", info.Config.User)
		}
		if len(info.Config.Entrypoint) > 0 {
			fmt.Fprintf(&sb, "Entrypoint: [%s]\n", strings.Join(info.Config.Entrypoint, ", "))
		}
		if len(info.Config.Cmd) > 0 {
			fmt.Fprintf(&sb, "Cmd: [%s]\n", strings.Join(info.Config.Cmd, ", "))
		}

		if len(info.Config.Env) > 0 {
			fmt.Fprintf(&sb, "Env (%d vars):\n", len(info.Config.Env))
			for _, env := range info.Config.Env {
				fmt.Fprintf(&sb, "  %s\n", env)
			}
		}

		if len(info.Config.ExposedPorts) > 0 {
			ports := make([]string, 0, len(info.Config.ExposedPorts))
			for p := range info.Config.ExposedPorts {
				ports = append(ports, string(p))
			}
			fmt.Fprintf(&sb, "ExposedPorts: %s\n", strings.Join(ports, ", "))
		}
	}

	if len(info.Config.Labels) > 0 {
		fmt.Fprintf(&sb, "\n── Labels ──\n")
		for k, v := range info.Config.Labels {
			fmt.Fprintf(&sb, "  %s=%s\n", k, v)
		}
	}

	return sb.String(), nil
}
