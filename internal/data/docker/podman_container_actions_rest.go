package docker

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	runtimepodman "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

func (c *Client) executeContainerPodmanREST(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions) error {
	if c.podmanREST == nil {
		return runtimeapi.NewError(runtimeapi.ErrorUnavailable, "container."+string(action), id, errPodmanRESTNotReady)
	}
	query := make(url.Values)
	switch action {
	case runtimeapi.ActionStart:
		return c.podmanREST.Post(ctx, "container.start", runtimepodman.ContainerPath(id, "/start"), nil, nil, nil)
	case runtimeapi.ActionStop:
		setPodmanTimeout(query, options.Timeout)
		return c.podmanREST.Post(ctx, "container.stop", runtimepodman.ContainerPath(id, "/stop"), query, nil, nil)
	case runtimeapi.ActionRestart:
		setPodmanTimeout(query, options.Timeout)
		return c.podmanREST.Post(ctx, "container.restart", runtimepodman.ContainerPath(id, "/restart"), query, nil, nil)
	case runtimeapi.ActionKill:
		if options.Signal != "" {
			query.Set("signal", options.Signal)
		}
		return c.podmanREST.Post(ctx, "container.kill", runtimepodman.ContainerPath(id, "/kill"), query, nil, nil)
	case runtimeapi.ActionPause:
		return c.podmanREST.Post(ctx, "container.pause", runtimepodman.ContainerPath(id, "/pause"), nil, nil, nil)
	case runtimeapi.ActionUnpause:
		return c.podmanREST.Post(ctx, "container.unpause", runtimepodman.ContainerPath(id, "/unpause"), nil, nil, nil)
	case runtimeapi.ActionRename:
		if options.Name == "" {
			return runtimeapi.NewError(runtimeapi.ErrorInvalid, "container.rename", id, fmt.Errorf("name is required"))
		}
		query.Set("name", options.Name)
		return c.podmanREST.Post(ctx, "container.rename", runtimepodman.ContainerPath(id, "/rename"), query, nil, nil)
	case runtimeapi.ActionRemove:
		query.Set("force", strconv.FormatBool(options.Force))
		return c.podmanREST.DeleteWithQuery(ctx, "container.remove", runtimepodman.ContainerPath(id, ""), query)
	default:
		return runtimeapi.UnsupportedError("container." + string(action))
	}
}

func setPodmanTimeout(query url.Values, timeout time.Duration) {
	if timeout > 0 {
		query.Set("timeout", strconv.FormatInt(int64(timeout/time.Second), 10))
	}
}
