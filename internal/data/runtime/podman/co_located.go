package podman

import (
	"context"
	"strings"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// coLocatedCache is process-global per the podman adapter. 5-second TTL
// (per R08-15 Q2). Invalidation is driven by the podman events handler.
var coLocatedCache = runtimeapi.NewCoLocatedCache(5 * time.Second)

// CoLocatedInvalider exposes Invalidate for the events handler.
func CoLocatedInvalider() func() {
	return coLocatedCache.Invalidate
}

func podmanInspectFn(s *PodmanComposeService) runtimeapi.InspectFn {
	return func(ctx context.Context, id string) (runtimeapi.InspectFields, error) {
		if s == nil || s.Client == nil || s.Client.REST == nil {
			return runtimeapi.InspectFields{}, errPodmanNotInit
		}
		info, err := s.Client.REST.InspectContainer(ctx, id)
		if err != nil {
			return runtimeapi.InspectFields{}, err
		}
		if info == nil || info.HostConfig == nil {
			return runtimeapi.InspectFields{}, nil
		}
		return runtimeapi.InspectFields{
			NetworkMode: info.HostConfig.NetworkMode,
		}, nil
	}
}

// FillCoLocated populates NetworkMode, PidMode, IpcMode, CoLocatedGroupID
// and CoLocatedGroupSrc on the supplied containers in place. Podman
// semantics differ from docker: podman-compose's default behaviour
// puts all containers of a project into a single project pod, so
// CoLocatedGroupID = "pod_<project>" and Source = "podman_default_pod".
// Inspect-derived mode fields are still filled because the podman
// driver can also surface service-mode references (rare under default
// podman-compose but possible under explicit podman setups).
func FillCoLocated(ctx context.Context, s *PodmanComposeService, containers []runtimeapi.ContainerSummary) error {
	inspect := podmanInspectFn(s)
	for i := range containers {
		c := &containers[i]
		if c.ComposeProject == "" {
			continue
		}
		fields, err := coLocatedCache.Get(ctx, c.ID, inspect)
		if err != nil {
			return err
		}
		c.NetworkMode = fields.NetworkMode
		c.PidMode = fields.PidMode
		c.IpcMode = fields.IpcMode
		c.CoLocatedGroupID = "pod_" + c.ComposeProject
		c.CoLocatedGroupSrc = "podman_default_pod"
	}
	_ = strings.TrimSpace
	return nil
}

var errPodmanNotInit = &podmanError{msg: "podman client not initialized"}

type podmanError struct{ msg string }

func (e *podmanError) Error() string { return e.msg }
