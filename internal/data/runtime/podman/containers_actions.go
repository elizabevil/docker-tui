package podman

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// ExecuteContainerActionREST performs a lifecycle action on a container via
// the Podman Libpod REST API.
func ExecuteContainerActionREST(ctx context.Context, client *Client, id string, action runtimeapi.Action, options runtimeapi.ActionOptions) error {
	if client.REST == nil {
		return runtimeapi.NewError(runtimeapi.ErrorUnavailable, "container."+string(action), id, errPodmanRESTNotReady)
	}
	query := make(url.Values)
	switch action {
	case runtimeapi.ActionStart:
		return client.REST.Post(ctx, "container.start", fmt.Sprintf(PathContainerStart, url.PathEscape(id)), nil, nil, nil)
	case runtimeapi.ActionStop:
		setPodmanTimeout(query, options.Timeout)
		return client.REST.Post(ctx, "container.stop", fmt.Sprintf(PathContainerStop, url.PathEscape(id)), query, nil, nil)
	case runtimeapi.ActionRestart:
		setPodmanTimeout(query, options.Timeout)
		return client.REST.Post(ctx, "container.restart", fmt.Sprintf(PathContainerRestart, url.PathEscape(id)), query, nil, nil)
	case runtimeapi.ActionKill:
		if options.Signal != "" {
			query.Set("signal", options.Signal)
		}
		return client.REST.Post(ctx, "container.kill", fmt.Sprintf(PathContainerKill, url.PathEscape(id)), query, nil, nil)
	case runtimeapi.ActionPause:
		return client.REST.Post(ctx, "container.pause", fmt.Sprintf(PathContainerPause, url.PathEscape(id)), nil, nil, nil)
	case runtimeapi.ActionUnpause:
		return client.REST.Post(ctx, "container.unpause", fmt.Sprintf(PathContainerUnpause, url.PathEscape(id)), nil, nil, nil)
	case runtimeapi.ActionRename:
		if options.Name == "" {
			return runtimeapi.NewError(runtimeapi.ErrorInvalid, "container.rename", id, fmt.Errorf("name is required"))
		}
		query.Set("name", options.Name)
		return client.REST.Post(ctx, "container.rename", fmt.Sprintf(PathContainerRename, url.PathEscape(id)), query, nil, nil)
	case runtimeapi.ActionRemove:
		query.Set("force", strconv.FormatBool(options.Force))
		return client.REST.DeleteWithQuery(ctx, "container.remove", fmt.Sprintf(PathContainerRemove, url.PathEscape(id)), query)
	default:
		return runtimeapi.UnsupportedError("container." + string(action))
	}
}

func setPodmanTimeout(query url.Values, timeout time.Duration) {
	if timeout > 0 {
		query.Set("timeout", strconv.FormatInt(int64(timeout/time.Second), 10))
	}
}
