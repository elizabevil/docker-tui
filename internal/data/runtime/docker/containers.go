package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func (c *Client) listContainersDocker(ctx context.Context, opts runtimeapi.ContainerListOptions) ([]runtimeapi.ContainerSummary, error) {
	nativeFilters, err := opts.NativeFilters()
	if err != nil {
		return nil, err
	}
	f := filters.NewArgs()
	for k, vs := range nativeFilters {
		for _, v := range vs {
			f.Add(k, v)
		}
	}

	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:     opts.All,
		Limit:   opts.Limit,
		Filters: f,
	})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	result := make([]runtimeapi.ContainerSummary, 0, len(containers))
	for _, ctr := range containers {
		name := ""
		if len(ctr.Names) > 0 {
			name = ctr.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		portBindings := make([]runtimeapi.PortBinding, 0, len(ctr.Ports))
		for _, p := range ctr.Ports {
			portBindings = append(portBindings, runtimeapi.PortBinding{
				ContainerPort: p.PrivatePort,
				Protocol:      p.Type,
				HostIP:        p.IP,
				HostPort:      p.PublicPort,
			})
		}

		summary := runtimeapi.ContainerSummary{
			ID:           ctr.ID,
			Name:         name,
			Image:        ctr.Image,
			Status:       ctr.Status,
			State:        ctr.State,
			Created:      ctr.Created,
			PortBindings: portBindings,
			Labels:       ctr.Labels,
			MountCount:   len(ctr.Mounts),
		}
		for _, mount := range ctr.Mounts {
			if mount.Name != "" {
				summary.MountNames = append(summary.MountNames, mount.Name)
			} else if mount.Source != "" {
				summary.MountNames = append(summary.MountNames, mount.Source)
			}
		}

		if ctr.NetworkSettings != nil {
			summary.Networks = map[string]string{}
			for name, net := range ctr.NetworkSettings.Networks {
				summary.NetworkNames = append(summary.NetworkNames, name)
				if net.IPAddress != "" {
					summary.Networks[name] = net.IPAddress
					summary.IPs = append(summary.IPs, net.IPAddress)
				}
			}
		}

		if project, ok := ctr.Labels[runtimeapi.ComposeLabelProject]; ok {
			summary.ComposeProject = project
		}
		if service, ok := ctr.Labels[runtimeapi.ComposeLabelService]; ok {
			summary.ComposeService = service
		}
		runtimeapi.ApplyComposeLabels(ctr.Labels, &summary)

		result = append(result, summary)
	}
	return result, nil
}

func (c *Client) containerTopContext(ctx context.Context, id string) (runtimeapi.ContainerProcesses, error) {
	response, err := c.cli.ContainerTop(ctx, id, nil)
	if err != nil {
		return runtimeapi.ContainerProcesses{}, fmt.Errorf("top container: %w", err)
	}
	return runtimeapi.ContainerProcesses{Titles: response.Titles, Processes: response.Processes}, nil
}

func (c *Client) containerLogsContext(ctx context.Context, id string, options runtimeapi.ContainerLogOptions) (io.ReadCloser, error) {
	resp, err := c.cli.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
		Timestamps: options.Timestamps,
		Since:      options.Since,
		Tail:       options.Tail,
	})
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	_, err = stdcopy.StdCopy(&buf, &buf, resp)
	_ = resp.Close() //nolint:errcheck // demux stream consumed; closing best-effort.
	if err != nil {
		return nil, fmt.Errorf("demux logs: %w", err)
	}
	return io.NopCloser(&buf), nil
}
