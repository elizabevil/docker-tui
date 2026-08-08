package podman

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	podmandriver "github.com/elizabevil/docker-tui/internal/driver/podman"
	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

type PodmanComposeService struct {
	Client *podmandriver.Client
}

var _ runtimeapi.ComposeService = PodmanComposeService{}

var errComposeUnsupported = errors.New(
	"podman compose: this command is not available via Engine API; use podman compose CLI",
)

// projectFilter returns the label-only filter for listing a compose
// project's containers. "All" (including stopped containers) is NOT a
// valid filter key per podman swagger — it is a separate top-level
// query parameter, handled by the driver via ContainerListOptions.All
// (see containers_rest.go:query.Set("all", ...)).
func (s PodmanComposeService) projectFilter(project string) map[string][]string {
	return map[string][]string{
		"label": {runtimeapi.ComposeProjectLabelValue(project)},
	}
}

// maxContainerNumber returns the highest com.docker.compose.container-number
// label among the project's containers of the given service (R08-08 F3).
func maxContainerNumber(containers []runtimeapi.ContainerSummary, service string) int {
	max := 0
	for _, c := range containers {
		if c.ComposeService != service {
			continue
		}
		n, err := strconv.Atoi(c.ContainerNumber)
		if err == nil && n > max {
			max = n
		}
	}
	return max
}

func (s PodmanComposeService) listContainers(ctx context.Context, project string) ([]runtimeapi.ContainerSummary, error) {
	if s.Client == nil || s.Client.REST == nil {
		return nil, podmandriver.ErrPodmanRESTNotReady
	}
	filters, err := podmandriver.PodmanFilterQuery(s.projectFilter(project))
	if err != nil {
		return nil, err
	}
	raw, err := s.Client.REST.ListContainers(ctx, dto.ContainerListOptions{
		All:     true,
		Limit:   0,
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}
	return MapContainerSummaries(raw), nil
}

func (s PodmanComposeService) ListProjects(ctx context.Context, opts runtimeapi.ListProjectsOptions) ([]runtimeapi.ComposeProjectSummary, error) {
	if s.Client == nil || s.Client.REST == nil {
		return nil, podmandriver.ErrPodmanRESTNotReady
	}
	filters, err := podmandriver.PodmanFilterQuery(map[string][]string{"label": {runtimeapi.ComposeLabelProject}})
	if err != nil {
		return nil, err
	}
	raw, err := s.Client.REST.ListContainers(ctx, dto.ContainerListOptions{All: opts.All, Limit: 0, Filters: filters})
	if err != nil {
		return nil, err
	}
	summaries := MapContainerSummaries(raw)
	byProject := make(map[string]*runtimeapi.ComposeProjectSummary)
	for _, c := range summaries {
		p := c.ComposeProject
		if p == "" {
			continue
		}
		entry, ok := byProject[p]
		if !ok {
			entry = &runtimeapi.ComposeProjectSummary{
				Name:        p,
				Source:      "podman",
				Services:    []runtimeapi.ComposeServiceSummary{},
				WorkingDir:  c.WorkingDir,
				ConfigFiles: c.ConfigFiles,
				Version:     c.Version,
			}
			byProject[p] = entry
		}
		entry.Services = append(entry.Services, runtimeapi.ComposeServiceSummary{
			Name: c.ComposeService,
			Image: c.Image,
		})
		entry.HasRunning = entry.HasRunning || c.State == "running"
		entry.HasStopped = entry.HasStopped || (c.State == "exited" || c.State == "stopped" || c.State == "dead")
	}
	out := make([]runtimeapi.ComposeProjectSummary, 0, len(byProject))
	for _, p := range byProject {
		out = append(out, *p)
	}
	return out, nil
}

func (s PodmanComposeService) InspectProject(ctx context.Context, project string) (runtimeapi.ComposeProjectSummary, error) {
	if project == "" {
		return runtimeapi.ComposeProjectSummary{}, fmt.Errorf("project name required")
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return runtimeapi.ComposeProjectSummary{}, err
	}
	out := runtimeapi.ComposeProjectSummary{
		Name:     project,
		Source:   "podman",
		Services: []runtimeapi.ComposeServiceSummary{},
	}
	for _, c := range containers {
		if out.WorkingDir == "" && c.WorkingDir != "" {
			out.WorkingDir = c.WorkingDir
		}
		if len(out.ConfigFiles) == 0 && len(c.ConfigFiles) > 0 {
			out.ConfigFiles = c.ConfigFiles
		}
		if out.Version == "" && c.Version != "" {
			out.Version = c.Version
		}
		out.Services = append(out.Services, runtimeapi.ComposeServiceSummary{
			Name:  c.ComposeService,
			Image: c.Image,
		})
		out.HasRunning = out.HasRunning || c.State == "running"
		out.HasStopped = out.HasStopped || (c.State == "exited" || c.State == "stopped" || c.State == "dead")
	}
	return out, nil
}

func (s PodmanComposeService) Config(ctx context.Context, project string) (string, error) {
	return "", errComposeUnsupported
}

func (s PodmanComposeService) Start(ctx context.Context, project string, services []string) error {
	return s.runContainerAction(ctx, project, services, podmandriver.ActionStart, podmandriver.ContainerActionOptions{})
}

func (s PodmanComposeService) Stop(ctx context.Context, project string, services []string, timeoutSec int) error {
	return s.runContainerAction(ctx, project, services, podmandriver.ActionStop, podmandriver.ContainerActionOptions{Timeout: time.Duration(timeoutSec) * time.Second})
}

func (s PodmanComposeService) Restart(ctx context.Context, project string, services []string) error {
	return s.runContainerAction(ctx, project, services, podmandriver.ActionRestart, podmandriver.ContainerActionOptions{})
}

func (s PodmanComposeService) Up(ctx context.Context, project string, opts runtimeapi.UpOptions) error {
	return s.runContainerAction(ctx, project, nil, podmandriver.ActionStart, podmandriver.ContainerActionOptions{})
}

func (s PodmanComposeService) Down(ctx context.Context, project string, opts runtimeapi.DownOptions) error {
	if s.Client == nil || s.Client.REST == nil {
		return podmandriver.ErrPodmanRESTNotReady
	}
	// Collect project image references before containers are removed — the
	// container list is the only source of image usage, and removal destroys
	// it (R08-04 F2 --rmi).
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
					continue // pulled image with explicit image: field
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
	if err := s.runContainerAction(ctx, project, nil, podmandriver.ActionRemove, podmandriver.ContainerActionOptions{Force: true}); err != nil {
		return err
	}
	if opts.RemoveVolumes {
		filters, err := podmandriver.PodmanFilterQuery(s.projectFilter(project))
		if err != nil {
			return err
		}
		vols, err := s.Client.REST.ListVolumes(ctx, dto.VolumeListOptions{Filters: filters})
		if err != nil {
			return err
		}
		for _, v := range vols {
			if err := s.Client.REST.RemoveVolume(ctx, v.Name, true); err != nil {
				return err
			}
		}
	}
	for _, ref := range imageRefs {
		if _, err := s.Client.REST.ExecuteImageAction(ctx, ref, podmandriver.ActionRemove, true, nil); err != nil {
			return err
		}
	}
	filters, err := podmandriver.PodmanFilterQuery(s.projectFilter(project))
	if err != nil {
		return err
	}
	nets, err := s.Client.REST.ListNetworks(ctx, dto.NetworkListOptions{Filters: filters})
	if err != nil {
		return err
	}
	for _, n := range nets {
		if err := s.Client.REST.RemoveNetwork(ctx, n.Name); err != nil {
			return err
		}
	}
	return nil
}

func (s PodmanComposeService) Run(ctx context.Context, project, service string, command []string, opts runtimeapi.RunOptions) error {
	if s.Client == nil || s.Client.REST == nil {
		return podmandriver.ErrPodmanRESTNotReady
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return err
	}
	var image string
	var configHash string
	for _, c := range containers {
		if c.ComposeService == service {
			image = c.Image
			configHash = c.ConfigHash
			break
		}
	}
	if image == "" {
		return fmt.Errorf("service %q not found in project %q", service, project)
	}
	labels := map[string]string{
		runtimeapi.ComposeLabelProject:         project,
		runtimeapi.ComposeLabelService:         service,
		runtimeapi.ComposeLabelOneoff:          "true",
		runtimeapi.ComposeLabelContainerNumber: strconv.Itoa(maxContainerNumber(containers, service) + 1),
		"dtui.compose.run.timestamp":           time.Now().UTC().Format(time.RFC3339),
	}
	if configHash != "" {
		labels[runtimeapi.ComposeLabelConfigHash] = configHash
	}
	spec := &dto.SpecGenerator{
		Name:    fmt.Sprintf("%s_%s_oneoff", project, service),
		Image:   image,
		Command: command,
		Labels:  labels,
	}
	if len(opts.EnvOverrides) > 0 {
		spec.Env = opts.EnvOverrides
	}
	if opts.EntrypointOverride != "" {
		spec.Entrypoint = []string{opts.EntrypointOverride}
	}
	if opts.User != "" {
		spec.User = opts.User
	}
	// Remove is delegated to the engine only for detached runs; attached
	// runs remove explicitly after wait (R08-08 F2 step 5).
	spec.Remove = opts.RemoveAfter && opts.Detach
	created, err := s.Client.REST.ContainerCreate(ctx, spec)
	if err != nil {
		return err
	}
	if err := s.Client.REST.ExecuteContainerAction(ctx, created.ID, podmandriver.ActionStart, podmandriver.ContainerActionOptions{}); err != nil {
		return err
	}
	if opts.Detach {
		return nil
	}
	if _, err := s.Client.REST.ContainerWait(ctx, created.ID, ""); err != nil {
		return err
	}
	if opts.RemoveAfter {
		return s.Client.REST.ExecuteContainerAction(ctx, created.ID, podmandriver.ActionRemove, podmandriver.ContainerActionOptions{Force: true})
	}
	return nil
}

func (s PodmanComposeService) Exec(ctx context.Context, project, service string, command []string) error {
	return s.runContainerAction(ctx, project, []string{service}, podmandriver.ActionExec, podmandriver.ContainerActionOptions{})
}

func (s PodmanComposeService) Top(ctx context.Context, project, service string) (runtimeapi.ContainerProcesses, error) {
	if s.Client == nil || s.Client.REST == nil {
		return runtimeapi.ContainerProcesses{}, podmandriver.ErrPodmanRESTNotReady
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return runtimeapi.ContainerProcesses{}, err
	}
	for _, c := range containers {
		if c.ComposeService != service {
			continue
		}
		raw, err := s.Client.REST.ContainerTop(ctx, c.ID)
		if err == nil {
			return runtimeapi.ContainerProcesses{Titles: raw.Titles, Processes: raw.Processes}, nil
		}
	}
	return runtimeapi.ContainerProcesses{}, fmt.Errorf("no running container found for service %q", service)
}

func (s PodmanComposeService) Port(ctx context.Context, project, service string, port int) (string, error) {
	if s.Client == nil || s.Client.REST == nil {
		return "", podmandriver.ErrPodmanRESTNotReady
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return "", err
	}
	key := strconv.Itoa(port) + "/tcp"
	for _, c := range containers {
		if c.ComposeService != service {
			continue
		}
		info, err := s.Client.REST.InspectContainer(ctx, c.ID)
		if err != nil || info.NetworkSettings == nil {
			continue
		}
		bindings := info.NetworkSettings.Ports[key]
		if len(bindings) == 0 {
			continue
		}
		hostIP := bindings[0].HostIP
		if hostIP == "" {
			hostIP = "0.0.0.0"
		}
		return fmt.Sprintf("%s:%s", hostIP, bindings[0].HostPort), nil
	}
	return "", fmt.Errorf("port %d not exposed for service %q", port, service)
}

func (s PodmanComposeService) Stats(ctx context.Context, project, service string) (runtimeapi.ContainerStats, error) {
	if s.Client == nil || s.Client.REST == nil {
		return runtimeapi.ContainerStats{}, podmandriver.ErrPodmanRESTNotReady
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return runtimeapi.ContainerStats{}, err
	}
	for _, c := range containers {
		if c.ComposeService != service {
			continue
		}
		raw, err := s.Client.REST.ContainerStats(ctx, c.ID)
		if err == nil {
			return mapPodmanStats(raw), nil
		}
	}
	return runtimeapi.ContainerStats{}, fmt.Errorf("no running container for service %q", service)
}

func (s PodmanComposeService) Events(ctx context.Context, project string) (<-chan runtimeapi.ComposeEvent, error) {
	out := make(chan runtimeapi.ComposeEvent, 64)
	if s.Client == nil || s.Client.REST == nil {
		close(out)
		return out, podmandriver.ErrPodmanRESTNotReady
	}
	events, err := (PodmanEventService{Client: s.Client}).Subscribe(ctx, runtimeapi.EventOptions{
		// /libpod/events filter spec is generic in podman swagger; podman
		// rejects "label" and "type" as filter keys (these are valid only
		// on the docker compat /events endpoint). Server-side filter is
		// limited to container/event/image/pod/volume/network/daemon/status/
		// time — none of which can express "compose project label". We
		// subscribe to all container events and filter client-side by
		// ComposeLabelProject in Attributes (see event handler below).
	})
	if err != nil {
		close(out)
		return out, err
	}
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-events:
				if !ok {
					return
				}
				if item.Error != nil {
					continue
				}
				// /libpod/events rejects label/type filter keys; we subscribe
				// to all container events and filter client-side by the
				// compose project label set on the source container.
				if item.Event.Attributes[runtimeapi.ComposeLabelProject] != project {
					continue
				}
				out <- runtimeapi.ComposeEvent{
					Type:      item.Event.ResourceType,
					Project:   project,
					Service:   item.Event.Attributes[runtimeapi.ComposeLabelService],
					Action:    item.Event.Action,
					Timestamp: time.Unix(0, item.Event.TimeNano),
				}
			}
		}
	}()
	return out, nil
}

func (s PodmanComposeService) runContainerAction(ctx context.Context, project string, services []string, action string, opts podmandriver.ContainerActionOptions) error {
	if s.Client == nil || s.Client.REST == nil {
		return podmandriver.ErrPodmanRESTNotReady
	}
	containers, err := s.listContainers(ctx, project)
	if err != nil {
		return err
	}
	filterSet := make(map[string]struct{}, len(services))
	for _, sv := range services {
		filterSet[sv] = struct{}{}
	}
	var failures []string
	for _, c := range containers {
		if len(services) > 0 {
			if _, ok := filterSet[c.ComposeService]; !ok {
				continue
			}
		}
		if err := s.Client.REST.ExecuteContainerAction(ctx, c.ID, action, opts); err != nil {
			failures = append(failures, c.ID)
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("action %s failed for: %s", action, strings.Join(failures, ","))
	}
	return nil
}
