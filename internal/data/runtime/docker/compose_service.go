package docker

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

type DockerComposeService struct {
	Client *Client
}

var _ runtimeapi.ComposeService = DockerComposeService{}

var errDockerComposeUnsupported = errors.New(
	"docker compose: this command is not available via Engine API; use docker compose CLI",
)

func (s DockerComposeService) projectFilter(project string) map[string][]string {
	return map[string][]string{
		"label": {runtimeapi.ComposeProjectLabelValue(project)},
	}
}

func buildFilter(m map[string][]string) filters.Args {
	f := filters.NewArgs()
	for k, vs := range m {
		for _, v := range vs {
			f.Add(k, v)
		}
	}
	return f
}

func (s DockerComposeService) listContainers(ctx context.Context, project string) ([]container.Summary, error) {
	if s.Client == nil || s.Client.cli == nil {
		return nil, fmt.Errorf("docker client not initialized")
	}
	return s.Client.cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: buildFilter(s.projectFilter(project)),
	})
}

func (s DockerComposeService) ListProjects(ctx context.Context, opts runtimeapi.ListProjectsOptions) ([]runtimeapi.ComposeProjectSummary, error) {
	if s.Client == nil || s.Client.cli == nil {
		return nil, fmt.Errorf("docker client not initialized")
	}
	containers, err := s.Client.cli.ContainerList(ctx, container.ListOptions{
		All:     opts.All,
		Filters: buildFilter(map[string][]string{"label": {runtimeapi.ComposeLabelProject}}),
	})
	if err != nil {
		return nil, err
	}
	byProject := make(map[string]*runtimeapi.ComposeProjectSummary)
	for _, ctr := range containers {
		p := ctr.Labels[runtimeapi.ComposeLabelProject]
		if p == "" {
			continue
		}
		entry, ok := byProject[p]
		if !ok {
			entry = &runtimeapi.ComposeProjectSummary{
				Name:     p,
				Source:   "docker",
				Services: []runtimeapi.ComposeServiceSummary{},
			}
			byProject[p] = entry
			// R08-03 F3: capture working_dir / config_files / version from
			// the first container's labels. Docker sets these on every
			// replica; the first container is the simplest correct
			// source. ConfigFiles is the JSON-escaped multi-line
			// compose.yaml path list; split on '\n'.
			if wd := ctr.Labels[runtimeapi.ComposeLabelWorkingDir]; wd != "" {
				entry.WorkingDir = wd
			}
			if v := ctr.Labels[runtimeapi.ComposeLabelVersion]; v != "" {
				entry.Version = v
			}
			if cf := ctr.Labels[runtimeapi.ComposeLabelConfigFiles]; cf != "" {
				for _, line := range strings.Split(cf, "\n") {
					if t := strings.TrimSpace(line); t != "" {
						entry.ConfigFiles = append(entry.ConfigFiles, t)
					}
				}
			}
		}
		entry.Services = append(entry.Services, runtimeapi.ComposeServiceSummary{
			Name:  ctr.Labels[runtimeapi.ComposeLabelService],
			Image: ctr.Image,
		})
		entry.HasRunning = entry.HasRunning || ctr.State == "running"
		entry.HasStopped = entry.HasStopped || (ctr.State == "exited" || ctr.State == "stopped" || ctr.State == "dead")
	}
	out := make([]runtimeapi.ComposeProjectSummary, 0, len(byProject))
	for _, p := range byProject {
		out = append(out, *p)
	}
	return out, nil
}

func (s DockerComposeService) InspectProject(ctx context.Context, project string) (runtimeapi.ComposeProjectSummary, error) {
	if project == "" {
		return runtimeapi.ComposeProjectSummary{}, fmt.Errorf("project name required")
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return runtimeapi.ComposeProjectSummary{}, err
	}
	out := runtimeapi.ComposeProjectSummary{
		Name:     project,
		Source:   "docker",
		Services: []runtimeapi.ComposeServiceSummary{},
	}
	for _, c := range containers {
		// R08-03 F3: copy working_dir / config_files / version from
		// the first container that carries them. Replicas in a project
		// share the same compose.yaml, so the first match is
		// representative.
		if out.WorkingDir == "" {
			out.WorkingDir = c.Labels[runtimeapi.ComposeLabelWorkingDir]
		}
		if out.Version == "" {
			out.Version = c.Labels[runtimeapi.ComposeLabelVersion]
		}
		if len(out.ConfigFiles) == 0 {
			if cf := c.Labels[runtimeapi.ComposeLabelConfigFiles]; cf != "" {
				for _, line := range strings.Split(cf, "\n") {
					if t := strings.TrimSpace(line); t != "" {
						out.ConfigFiles = append(out.ConfigFiles, t)
					}
				}
			}
		}
		out.Services = append(out.Services, runtimeapi.ComposeServiceSummary{
			Name:  c.Labels[runtimeapi.ComposeLabelService],
			Image: c.Image,
		})
		out.HasRunning = out.HasRunning || c.State == "running"
		out.HasStopped = out.HasStopped || (c.State == "exited" || c.State == "stopped" || c.State == "dead")
	}
	return out, nil
}

func (s DockerComposeService) Config(ctx context.Context, project string) (string, error) {
	return "", errDockerComposeUnsupported
}

func (s DockerComposeService) Start(ctx context.Context, project string, services []string) error {
	return s.runContainerAction(ctx, project, services, dockerContainerStart, 0)
}

func (s DockerComposeService) Stop(ctx context.Context, project string, services []string, timeoutSec int) error {
	return s.runContainerAction(ctx, project, services, dockerContainerStop, timeoutSec)
}

func (s DockerComposeService) Restart(ctx context.Context, project string, services []string) error {
	return s.runContainerAction(ctx, project, services, dockerContainerRestart, 10)
}

func (s DockerComposeService) Up(ctx context.Context, project string, opts runtimeapi.UpOptions) error {
	_ = opts
	return s.runContainerAction(ctx, project, nil, dockerContainerStart, 0)
}

func (s DockerComposeService) Down(ctx context.Context, project string, opts runtimeapi.DownOptions) error {
	if s.Client == nil || s.Client.cli == nil {
		return fmt.Errorf("docker client not initialized")
	}
	var imageRefs []string
	if opts.RemoveImages != "" {
		containers, err := s.listContainers(ctx, project)
		if err != nil {
			return err
		}
		seen := make(map[string]struct{})
		for _, c := range containers {
			if opts.RemoveImages == "local" {
				if _, pulled := c.Labels[runtimeapi.ComposeLabelImage]; pulled {
					continue
				}
			}
			if c.Image == "" {
				continue
			}
			if _, ok := seen[c.Image]; ok {
				continue
			}
			seen[c.Image] = struct{}{}
			imageRefs = append(imageRefs, c.Image)
		}
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return err
	}
	for _, c := range containers {
		if err := s.Client.cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true}); err != nil {
			return err
		}
	}
	if opts.RemoveVolumes {
		volumes, err := s.Client.ListVolumesContext(ctx, runtimeapi.VolumeListOptions{
			Filters: runtimeapi.FilterSet{runtimeapi.VolumeFilterLabel: {runtimeapi.ComposeProjectLabelValue(project)}},
		})
		if err != nil {
			return err
		}
		for _, v := range volumes {
			if err := s.Client.cli.VolumeRemove(ctx, v.Name, true); err != nil {
				return err
			}
		}
	}
	for _, ref := range imageRefs {
		if _, err := s.Client.cli.ImageRemove(ctx, ref, image.RemoveOptions{Force: true, PruneChildren: true}); err != nil {
			return err
		}
	}
	networks, err := s.Client.ListNetworksContext(ctx, runtimeapi.NetworkListOptions{
		Filters: runtimeapi.FilterSet{runtimeapi.NetworkFilterLabel: {runtimeapi.ComposeProjectLabelValue(project)}},
	})
	if err != nil {
		return err
	}
	for _, n := range networks {
		if err := s.Client.cli.NetworkRemove(ctx, n.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s DockerComposeService) Run(ctx context.Context, project, service string, command []string, opts runtimeapi.RunOptions) error {
	if s.Client == nil || s.Client.cli == nil {
		return fmt.Errorf("docker client not initialized")
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return err
	}
	var image string
	for _, c := range containers {
		if c.Labels[runtimeapi.ComposeLabelService] == service {
			image = c.Image
			break
		}
	}
	if image == "" {
		return fmt.Errorf("service %q not found in project %q", service, project)
	}
	// R08-08 §F3: container-number is max+1 of existing replicas in this
	// service so the engine's container-name auto-allocator does not
	// collide with pre-existing compose-managed containers. Docker
	// stores the label inside the Summary.Labels map.
	containerNumber := 1
	for _, c := range containers {
		if c.Labels[runtimeapi.ComposeLabelService] != service {
			continue
		}
		if n, err := strconv.Atoi(c.Labels[runtimeapi.ComposeLabelContainerNumber]); err == nil && n >= containerNumber {
			containerNumber = n + 1
		}
	}
	containerName := fmt.Sprintf("%s_%s_oneoff", project, service)
	cfg := &container.Config{
		Image: image,
		Cmd:   command,
		Labels: map[string]string{
			runtimeapi.ComposeLabelProject:         project,
			runtimeapi.ComposeLabelService:         service,
			runtimeapi.ComposeLabelOneoff:          "true",
			runtimeapi.ComposeLabelContainerNumber: strconv.Itoa(containerNumber),
			"dtui.compose.run.timestamp":           time.Now().UTC().Format(time.RFC3339),
		},
	}
	for k, v := range opts.EnvOverrides {
		cfg.Env = append(cfg.Env, k+"="+v)
	}
	if opts.EntrypointOverride != "" {
		cfg.Entrypoint = []string{opts.EntrypointOverride}
	}
	if opts.User != "" {
		cfg.User = opts.User
	}
	hc := &container.HostConfig{
		AutoRemove: opts.RemoveAfter && opts.Detach,
	}
	created, err := s.Client.cli.ContainerCreate(ctx, cfg, hc, nil, nil, containerName)
	if err != nil {
		return err
	}
	if err := s.Client.cli.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
		return err
	}
	if opts.Detach {
		return nil
	}
	wait, errC := s.Client.cli.ContainerWait(ctx, created.ID, container.WaitConditionNotRunning)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errC:
		return err
	case <-wait:
	}
	if opts.RemoveAfter {
		_ = s.Client.cli.ContainerRemove(ctx, created.ID, container.RemoveOptions{Force: true})
	}
	return nil
}

func (s DockerComposeService) Exec(ctx context.Context, project, service string, command []string) error {
	if s.Client == nil || s.Client.cli == nil {
		return fmt.Errorf("docker client not initialized")
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return err
	}
	for _, c := range containers {
		if c.Labels[runtimeapi.ComposeLabelService] != service {
			continue
		}
		execCfg := &container.ExecOptions{
			Cmd:          command,
			AttachStdout: false,
			AttachStderr: false,
			Tty:          false,
		}
		resp, err := s.Client.cli.ContainerExecCreate(ctx, c.ID, *execCfg)
		if err != nil {
			return fmt.Errorf("exec create on %s: %w", c.ID, err)
		}
		if err := s.Client.cli.ContainerExecStart(ctx, resp.ID, container.ExecStartOptions{}); err != nil {
			return fmt.Errorf("exec start on %s: %w", c.ID, err)
		}
		return nil
	}
	return fmt.Errorf("no container found for service %q in project %q", service, project)
}

func (s DockerComposeService) Top(ctx context.Context, project, service string) (runtimeapi.ContainerProcesses, error) {
	if s.Client == nil || s.Client.cli == nil {
		return runtimeapi.ContainerProcesses{}, fmt.Errorf("docker client not initialized")
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return runtimeapi.ContainerProcesses{}, err
	}
	for _, c := range containers {
		if c.Labels[runtimeapi.ComposeLabelService] != service {
			continue
		}
		procs, err := s.Client.cli.ContainerTop(ctx, c.ID, []string{})
		if err == nil {
			return runtimeapi.ContainerProcesses{
				Titles:    procs.Titles,
				Processes: procs.Processes,
			}, nil
		}
	}
	return runtimeapi.ContainerProcesses{}, fmt.Errorf("no running container found for service %q", service)
}

func (s DockerComposeService) Port(ctx context.Context, project, service string, port int) (string, error) {
	if s.Client == nil || s.Client.cli == nil {
		return "", fmt.Errorf("docker client not initialized")
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return "", err
	}
	for _, c := range containers {
		if c.Labels[runtimeapi.ComposeLabelService] != service {
			continue
		}
		info, err := s.Client.cli.ContainerInspect(ctx, c.ID)
		if err != nil || info.NetworkSettings == nil {
			continue
		}
		portKeyStr := strconv.Itoa(port) + "/tcp"
		var matched bool
		var hostIP string
		var hostPort string
		for key, ports := range info.NetworkSettings.Ports {
			if string(key) != portKeyStr {
				continue
			}
			if len(ports) == 0 {
				continue
			}
			hostIP = ports[0].HostIP
			hostPort = ports[0].HostPort
			matched = true
			break
		}
		if !matched {
			continue
		}
		if hostIP == "" {
			hostIP = "0.0.0.0"
		}
		return fmt.Sprintf("%s:%s", hostIP, hostPort), nil
	}
	return "", fmt.Errorf("port %d not exposed for service %q", port, service)
}

func (s DockerComposeService) Stats(ctx context.Context, project, service string) (runtimeapi.ContainerStats, error) {
	if s.Client == nil || s.Client.cli == nil {
		return runtimeapi.ContainerStats{}, fmt.Errorf("docker client not initialized")
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return runtimeapi.ContainerStats{}, err
	}
	for _, c := range containers {
		if c.Labels[runtimeapi.ComposeLabelService] != service {
			continue
		}
		stats, err := s.Client.containerStatsContext(ctx, c.ID)
		if err == nil {
			return stats, nil
		}
	}
	return runtimeapi.ContainerStats{}, fmt.Errorf("no running container for service %q", service)
}

func (s DockerComposeService) Events(ctx context.Context, project string) (<-chan runtimeapi.ComposeEvent, error) {
	out := make(chan runtimeapi.ComposeEvent, 64)
	if s.Client == nil || s.Client.cli == nil {
		close(out)
		return out, fmt.Errorf("docker client not initialized")
	}
	events, errCh := s.Client.cli.Events(ctx, events.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("type", "container"),
			filters.Arg("label", runtimeapi.ComposeProjectLabelValue(project)),
		),
	})
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-events:
				if !ok {
					return
				}
				out <- runtimeapi.ComposeEvent{
					Type:      string(ev.Action),
					Project:   project,
					Service:   ev.Actor.Attributes[runtimeapi.ComposeLabelService],
					Action:    string(ev.Action),
					Timestamp: time.Unix(0, ev.TimeNano),
				}
			case err, ok := <-errCh:
				if !ok {
					return
				}
				_ = err
			}
		}
	}()
	return out, nil
}

type dockerContainerAction int

const (
	dockerContainerStart dockerContainerAction = iota
	dockerContainerStop
	dockerContainerRestart
)

func (s DockerComposeService) runContainerAction(ctx context.Context, project string, services []string, action dockerContainerAction, timeoutSec int) error {
	if s.Client == nil || s.Client.cli == nil {
		return fmt.Errorf("docker client not initialized")
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return err
	}
	filterSet := make(map[string]struct{}, len(services))
	for _, sv := range services {
		filterSet[sv] = struct{}{}
	}
	var stopTimeout int
	if action == dockerContainerStop || action == dockerContainerRestart {
		stopTimeout = timeoutSec
	}
	var failures []string
	for _, c := range containers {
		if len(services) > 0 {
			if _, ok := filterSet[c.Labels[runtimeapi.ComposeLabelService]]; !ok {
				continue
			}
		}
		var err error
		switch action {
		case dockerContainerStart:
			err = s.Client.cli.ContainerStart(ctx, c.ID, container.StartOptions{})
		case dockerContainerStop:
			err = s.Client.cli.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &stopTimeout})
		case dockerContainerRestart:
			err = s.Client.cli.ContainerRestart(ctx, c.ID, container.StopOptions{Timeout: &stopTimeout})
		}
		if err != nil {
			failures = append(failures, c.ID)
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("action failed for: %s", strings.Join(failures, ","))
	}
	return nil
}
