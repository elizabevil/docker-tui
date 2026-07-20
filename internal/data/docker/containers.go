package docker

import (
	"bytes"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"
)

func (c *Client) ListContainers(opts ContainerListOptions) ([]ContainerSummary, error) {
	f := filters.NewArgs()
	for k, vs := range opts.Filters {
		for _, v := range vs {
			f.Add(k, v)
		}
	}

	containers, err := c.cli.ContainerList(c.ctx, container.ListOptions{
		All:     opts.All,
		Limit:   opts.Limit,
		Filters: f,
	})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	result := make([]ContainerSummary, 0, len(containers))
	for _, ctr := range containers {
		name := ""
		if len(ctr.Names) > 0 {
			name = ctr.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		portsStr := ""
		for i, p := range ctr.Ports {
			if i > 0 {
				portsStr += ", "
			}
			if p.PublicPort > 0 {
				portsStr += fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PrivatePort, p.Type)
			} else {
				portsStr += fmt.Sprintf("%d/%s", p.PrivatePort, p.Type)
			}
		}

		summary := ContainerSummary{
			ID:         ctr.ID[:12],
			Name:       name,
			Image:      ctr.Image,
			Status:     ctr.Status,
			State:      ctr.State,
			Created:    ctr.Created,
			Ports:      portsStr,
			Labels:     ctr.Labels,
			MountCount: len(ctr.Mounts),
		}

		if ctr.NetworkSettings != nil {
			for _, net := range ctr.NetworkSettings.Networks {
				if net.IPAddress != "" {
					summary.IPs = append(summary.IPs, net.IPAddress)
				}
			}
		}

		if project, ok := ctr.Labels["com.docker.compose.project"]; ok {
			summary.ComposeProject = project
		}
		if service, ok := ctr.Labels["com.docker.compose.service"]; ok {
			summary.ComposeService = service
		}

		result = append(result, summary)
	}
	return result, nil
}

func (c *Client) ContainerStart(id string) error {
	return c.cli.ContainerStart(c.ctx, id, container.StartOptions{})
}

func (c *Client) ContainerStop(id string) error {
	return c.cli.ContainerStop(c.ctx, id, container.StopOptions{})
}

func (c *Client) ContainerRestart(id string) error {
	return c.cli.ContainerRestart(c.ctx, id, container.StopOptions{})
}

func (c *Client) ContainerKill(id string) error {
	return c.cli.ContainerKill(c.ctx, id, "")
}

func (c *Client) ContainerRemove(id string, force bool) error {
	return c.cli.ContainerRemove(c.ctx, id, container.RemoveOptions{Force: force})
}

func (c *Client) ContainerLogs(id string, since, tail string, timestamps bool) (io.ReadCloser, error) {
	resp, err := c.cli.ContainerLogs(c.ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
		Timestamps: timestamps,
		Since:      since,
		Tail:       tail,
	})
	if err != nil {
		return nil, err
	}
	// Demultiplex Docker's multiplexed stream into plain text
	var buf bytes.Buffer
	_, err = stdcopy.StdCopy(&buf, &buf, resp)
	resp.Close()
	if err != nil {
		return nil, fmt.Errorf("demux logs: %w", err)
	}
	return io.NopCloser(&buf), nil
}

func (c *Client) ContainerStats(id string) (io.ReadCloser, error) {
	resp, err := c.cli.ContainerStats(c.ctx, id, false)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
