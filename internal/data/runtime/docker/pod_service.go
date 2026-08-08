package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type DockerPodService struct {
	Client *Client
}

var _ runtimeapi.PodService = DockerPodService{}

func (s DockerPodService) ListPods(ctx context.Context, opts runtimeapi.PodListOptions) ([]runtimeapi.PodSummary, error) {
	if s.Client == nil || s.Client.cli == nil {
		return nil, fmt.Errorf("docker client not initialized")
	}
	native, err := opts.NativeFilters()
	if err != nil {
		return nil, err
	}
	groups, err := s.synthesizeGroups(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]runtimeapi.PodSummary, 0, len(groups))
	for name, members := range groups {
		if len(members) < 2 {
			continue
		}
		if !matchPodFilters(name, members, native) {
			continue
		}
		running, stopped := countRunningStopped(members)
		out = append(out, runtimeapi.PodSummary{
			Name:    name,
			Source:  "docker.wrap",
			Running: running,
			Stopped: stopped,
		})
	}
	return out, nil
}

func matchPodFilters(name string, members []container.Summary, filters map[string][]string) bool {
	if vs := filters[runtimeapi.PodFilterName]; len(vs) > 0 && !strings.Contains(name, vs[0]) {
		return false
	}
	if vs := filters[runtimeapi.PodFilterLabel]; len(vs) > 0 {
		key, value, hasValue := strings.Cut(vs[0], "=")
		matched := false
		for _, m := range members {
			v, ok := m.Labels[key]
			if !ok {
				continue
			}
			if hasValue && v != value {
				continue
			}
			matched = true
			break
		}
		if !matched {
			return false
		}
	}
	return true
}

func (s DockerPodService) InspectPod(ctx context.Context, name string) (runtimeapi.PodDetail, error) {
	if s.Client == nil || s.Client.cli == nil {
		return runtimeapi.PodDetail{}, fmt.Errorf("docker client not initialized")
	}
	groups, err := s.synthesizeGroups(ctx)
	if err != nil {
		return runtimeapi.PodDetail{}, err
	}
	members, ok := groups[name]
	if !ok {
		return runtimeapi.PodDetail{}, nil
	}
	containers := make([]runtimeapi.PodContainer, 0, len(members))
	labels := make(map[string]string)
	for _, m := range members {
		containers = append(containers, runtimeapi.PodContainer{
			ContainerID: m.ID,
			Service:     serviceName(m),
			Number:      m.Labels[runtimeapi.ComposeLabelContainerNumber],
		})
		if project, ok := m.Labels[runtimeapi.ComposeLabelProject]; ok {
			labels[runtimeapi.ComposeLabelProject] = project
		}
	}
	return runtimeapi.PodDetail{
		Name:       name,
		Source:     "docker.wrap",
		Containers: containers,
		Labels:     labels,
	}, nil
}

func (s DockerPodService) synthesizeGroups(ctx context.Context) (map[string][]container.Summary, error) {
	opts := container.ListOptions{All: true}
	containers, err := s.Client.cli.ContainerList(ctx, opts)
	if err != nil {
		return nil, err
	}
	groups := make(map[string][]container.Summary)
	for _, ctr := range containers {
		if ctr.Labels[runtimeapi.ComposeLabelProject] == "" {
			continue
		}
		info, err := s.Client.cli.ContainerInspect(ctx, ctr.ID)
		if err != nil {
			continue
		}
		target := coLocatedTarget(info)
		if target == "" {
			continue
		}
		groups[target] = append(groups[target], ctr)
	}
	return groups, nil
}

func coLocatedTarget(info container.InspectResponse) string {
	if t := serviceRef(string(info.HostConfig.NetworkMode)); t != "" {
		return t
	}
	if t := serviceRef(string(info.HostConfig.PidMode)); t != "" {
		return t
	}
	if t := serviceRef(string(info.HostConfig.IpcMode)); t != "" {
		return t
	}
	return ""
}

func serviceRef(mode string) string {
	if after, ok := strings.CutPrefix(mode, "service:"); ok {
		return after
	}
	return ""
}

func countRunningStopped(members []container.Summary) (running, stopped int) {
	for _, m := range members {
		switch string(m.State) {
		case "running":
			running++
		case "exited", "dead", "created", "paused":
			stopped++
		}
	}
	return
}

func serviceName(c container.Summary) string {
	if name := c.Labels[runtimeapi.ComposeLabelService]; name != "" {
		return name
	}
	for _, n := range c.Names {
		return strings.TrimPrefix(n, "/")
	}
	return c.ID[:12]
}
