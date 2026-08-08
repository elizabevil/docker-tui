package podman

import (
	"context"
	"strconv"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	podmandriver "github.com/elizabevil/docker-tui/internal/driver/podman"
)

// PodmanPodService wraps the Podman REST client for pod list/inspect.
// Implementation note: per R08-15 §7.4 driver parity is a separate
// follow-up; the lowest-risk path uses REST directly. When the cgo
// bindings driver gains ListPods/InspectPod, the same dispatch pattern
// as container actions (prefer Driver, fall through to REST) can be
// applied here.
type PodmanPodService struct {
	Client *podmandriver.Client
}

var _ runtimeapi.PodService = PodmanPodService{}

func (s PodmanPodService) ListPods(ctx context.Context, opts runtimeapi.PodListOptions) ([]runtimeapi.PodSummary, error) {
	if s.Client == nil || s.Client.REST == nil {
		return nil, podmandriver.ErrPodmanRESTNotReady
	}
	native, err := opts.NativeFilters()
	if err != nil {
		return nil, err
	}
	raw, err := s.Client.REST.ListPods(ctx, native)
	if err != nil {
		return nil, err
	}
	out := make([]runtimeapi.PodSummary, 0, len(raw))
	for i := range raw {
		out = append(out, podmanPodToSummary(raw[i]))
	}
	return out, nil
}

func (s PodmanPodService) InspectPod(ctx context.Context, name string) (runtimeapi.PodDetail, error) {
	if s.Client == nil || s.Client.REST == nil {
		return runtimeapi.PodDetail{}, podmandriver.ErrPodmanRESTNotReady
	}
	raw, err := s.Client.REST.InspectPod(ctx, name)
	if err != nil {
		return runtimeapi.PodDetail{}, err
	}
	if raw == nil {
		return runtimeapi.PodDetail{}, nil
	}
	containers := make([]runtimeapi.PodContainer, 0, len(raw.Containers))
	for _, c := range raw.Containers {
		containers = append(containers, runtimeapi.PodContainer{
			ContainerID: c.Id,
			Service:     c.Names,
			Number:      strconv.FormatUint(uint64(c.Number), 10),
		})
	}
	return runtimeapi.PodDetail{
		Name:       raw.Name,
		Source:     "podman",
		Containers: containers,
		Labels:     raw.Labels,
	}, nil
}

func podmanPodToSummary(p podmandriver.ListPodReport) runtimeapi.PodSummary {
	running, stopped := countRunningStopped(p.Containers)
	return runtimeapi.PodSummary{
		Name:    p.Name,
		Source:  "podman",
		Running: running,
		Stopped: stopped,
		Created: p.Created,
	}
}

func countRunningStopped(cs []podmandriver.ListPodContainerItem) (running, stopped int) {
	for _, c := range cs {
		switch c.Status {
		case "running":
			running++
		case "stopped", "exited", "dead", "created":
			stopped++
		}
	}
	return
}
