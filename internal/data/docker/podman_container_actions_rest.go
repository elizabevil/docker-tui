package docker

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func (c *Client) executeContainerPodmanREST(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions) error {
	if c.podmanREST == nil {
		return runtimeapi.NewError(runtimeapi.ErrorUnavailable, "container."+string(action), id, fmt.Errorf("Podman REST transport is not initialized"))
	}
	path := "/containers/" + url.PathEscape(id)
	query := make(url.Values)
	switch action {
	case runtimeapi.ActionStart:
		return c.podmanREST.Post(ctx, "container.start", path+"/start", nil, nil, nil)
	case runtimeapi.ActionStop:
		setPodmanTimeout(query, options.Timeout)
		return c.podmanREST.Post(ctx, "container.stop", path+"/stop", query, nil, nil)
	case runtimeapi.ActionRestart:
		setPodmanTimeout(query, options.Timeout)
		return c.podmanREST.Post(ctx, "container.restart", path+"/restart", query, nil, nil)
	case runtimeapi.ActionKill:
		if options.Signal != "" {
			query.Set("signal", options.Signal)
		}
		return c.podmanREST.Post(ctx, "container.kill", path+"/kill", query, nil, nil)
	case runtimeapi.ActionPause:
		return c.podmanREST.Post(ctx, "container.pause", path+"/pause", nil, nil, nil)
	case runtimeapi.ActionUnpause:
		return c.podmanREST.Post(ctx, "container.unpause", path+"/unpause", nil, nil, nil)
	case runtimeapi.ActionRename:
		if options.Name == "" {
			return runtimeapi.NewError(runtimeapi.ErrorInvalid, "container.rename", id, fmt.Errorf("name is required"))
		}
		query.Set("name", options.Name)
		return c.podmanREST.Post(ctx, "container.rename", path+"/rename", query, nil, nil)
	case runtimeapi.ActionRemove:
		query.Set("force", strconv.FormatBool(options.Force))
		return c.podmanREST.DeleteWithQuery(ctx, "container.remove", path, query)
	default:
		return runtimeapi.UnsupportedError("container." + string(action))
	}
}

func setPodmanTimeout(query url.Values, timeout time.Duration) {
	if timeout > 0 {
		query.Set("timeout", strconv.FormatInt(int64(timeout/time.Second), 10))
	}
}
