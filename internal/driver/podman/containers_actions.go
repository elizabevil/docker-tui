package podman

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// ExecuteContainerAction performs a lifecycle action on a container via
// the Podman Libpod REST API. Action is a podman-native action name;
// callers in the runtime package map domain actions to these.
func (c *RESTClient) ExecuteContainerAction(ctx context.Context, id string, action string, timeout time.Duration, force bool, signal, name string) error {
	query := make(url.Values)
	switch action {
	case ActionStart:
		return c.Post(ctx, ContainerStartPath(id), nil, nil, nil)
	case ActionStop:
		setPodmanTimeout(query, timeout)
		return c.Post(ctx, ContainerStopPath(id), query, nil, nil)
	case ActionRestart:
		setPodmanTimeout(query, timeout)
		return c.Post(ctx, ContainerRestartPath(id), query, nil, nil)
	case ActionKill:
		if signal != "" {
			query.Set("signal", signal)
		}
		return c.Post(ctx, ContainerKillPath(id), query, nil, nil)
	case ActionPause:
		return c.Post(ctx, ContainerPausePath(id), nil, nil, nil)
	case ActionUnpause:
		return c.Post(ctx, ContainerUnpausePath(id), nil, nil, nil)
	case ActionRename:
		if name == "" {
			return newPodmanError(KindInvalid, "container.rename", fmt.Errorf("name is required"))
		}
		query.Set("name", name)
		return c.Post(ctx, ContainerRenamePath(id), query, nil, nil)
	case ActionRemove:
		query.Set("force", strconv.FormatBool(force))
		return c.DeleteWithQuery(ctx, ContainerRemovePath(id), query)
	default:
		return newPodmanError(KindInvalid, "container."+action, fmt.Errorf("unsupported action: %s", action))
	}
}

func setPodmanTimeout(query url.Values, timeout time.Duration) {
	if timeout > 0 {
		query.Set("timeout", strconv.FormatInt(int64(timeout/time.Second), 10))
	}
}
