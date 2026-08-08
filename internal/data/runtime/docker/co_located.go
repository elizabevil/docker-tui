package docker

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// coLocatedCache is process-global per the docker adapter. 5-second TTL
// (per R08-15 Q2). Invalidation is driven by the events handler wired
// by the docker adapter.
var coLocatedCache = runtimeapi.NewCoLocatedCache(5 * time.Second)

// CoLocatedInvalider exposes Invalidate for the events handler.
func CoLocatedInvalider() func() {
	return coLocatedCache.Invalidate
}

func dockerInspectFn(s *Client) runtimeapi.InspectFn {
	return func(ctx context.Context, id string) (runtimeapi.InspectFields, error) {
		if s == nil || s.cli == nil {
			return runtimeapi.InspectFields{}, errDockerNotInit
		}
		info, err := s.cli.ContainerInspect(ctx, id)
		if err != nil {
			return runtimeapi.InspectFields{}, err
		}
		return runtimeapi.InspectFields{
			NetworkMode: string(info.HostConfig.NetworkMode),
			PidMode:     string(info.HostConfig.PidMode),
			IpcMode:     string(info.HostConfig.IpcMode),
		}, nil
	}
}

// serviceTarget extracts the target service name from a "service:<name>"
// mode value; returns "" when the value is not a service reference.
func serviceTarget(mode string) string {
	if !strings.HasPrefix(mode, "service:") {
		return ""
	}
	return strings.TrimPrefix(mode, "service:")
}

// FillCoLocated populates NetworkMode, PidMode, IpcMode, CoLocatedGroupID
// and CoLocatedGroupSrc on the supplied containers in place. Containers
// with no compose project label are left untouched. The group ID
// follows docker semantics: 2+ containers in a project referencing the
// same target via any of the three modes (Q4 OR) are assigned the same
// "<project>/<target>" group; the source enum reflects which mode first
// matched.
func FillCoLocated(ctx context.Context, s *Client, containers []runtimeapi.ContainerSummary) error {
	inspect := dockerInspectFn(s)
	type groupKey struct {
		project string
		target  string
		src     string
	}
	memberOf := make(map[string]groupKey, len(containers))
	byID := make(map[string]*runtimeapi.ContainerSummary, len(containers))

	for i := range containers {
		c := &containers[i]
		byID[c.ID] = c
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

		var k groupKey
		k.project = c.ComposeProject
		switch {
		case serviceTarget(c.NetworkMode) != "":
			k.target = serviceTarget(c.NetworkMode)
			k.src = "docker_network_mode_service"
		case serviceTarget(c.PidMode) != "":
			k.target = serviceTarget(c.PidMode)
			k.src = "docker_pid_service"
		case serviceTarget(c.IpcMode) != "":
			k.target = serviceTarget(c.IpcMode)
			k.src = "docker_ipc_service"
		}
		if k.target != "" {
			memberOf[c.ID] = k
		}
	}

	for id, k := range memberOf {
		if c, ok := byID[id]; ok {
			c.CoLocatedGroupID = k.project + "/" + k.target
			c.CoLocatedGroupSrc = k.src
		}
	}
	return nil
}

// ContainerInspectResponse is a tiny shim used only for type assertions
// on the docker SDK inspect call; kept private to avoid exporting
// SDK-level types.
type ContainerInspectResponse = container.InspectResponse

var (
	errDockerNotInit = &dockerError{msg: "docker client not initialized"}
)

type dockerError struct{ msg string }

func (e *dockerError) Error() string { return e.msg }

var _ sync.Locker = (*sync.Mutex)(nil)
