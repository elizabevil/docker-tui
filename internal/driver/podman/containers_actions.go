package podman

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// ContainerActionOptions carries the parameters of an arbitrary container
// lifecycle action. Fields are action-specific and may be left at their
// zero value:
//
//	ActionStart   — no fields used
//	ActionStop    — Timeout
//	ActionRestart — Timeout
//	ActionKill    — Signal
//	ActionPause   — no fields used
//	ActionUnpause — no fields used
//	ActionRename  — Name (required)
//	ActionRemove  — Force
//
// Bundling the four fields into one struct keeps the dispatcher
// signature within the 5-parameter limit and makes the call site
// self-documenting (opts vs ctx,id,action,timeout,force,signal,name).
type ContainerActionOptions struct {
	// Timeout applies to stop / restart. Zero means "engine default".
	Timeout time.Duration
	// Force applies to remove.
	Force bool
	// Signal applies to kill. Empty defaults to SIGKILL on POSIX.
	Signal string
	// Name applies to rename.
	Name string
}

// ExecuteContainerAction performs a lifecycle action on a container via
// the Podman Libpod REST API. Action is a podman-native action name;
// callers in the runtime package map domain actions to these.
func (c *RESTClient) ExecuteContainerAction(ctx context.Context, id, action string, opts ContainerActionOptions) error {
	query := make(url.Values)
	switch action {
	case ActionStart:
		return c.Post(ctx, ContainerStartPath(id), nil, nil, nil)
	case ActionStop:
		setPodmanTimeout(query, opts.Timeout)
		return c.Post(ctx, ContainerStopPath(id), query, nil, nil)
	case ActionRestart:
		setPodmanTimeout(query, opts.Timeout)
		return c.Post(ctx, ContainerRestartPath(id), query, nil, nil)
	case ActionKill:
		if opts.Signal != "" {
			query.Set("signal", opts.Signal)
		}
		return c.Post(ctx, ContainerKillPath(id), query, nil, nil)
	case ActionPause:
		return c.Post(ctx, ContainerPausePath(id), nil, nil, nil)
	case ActionUnpause:
		return c.Post(ctx, ContainerUnpausePath(id), nil, nil, nil)
	case ActionRename:
		if opts.Name == "" {
			return newPodmanError(KindInvalid, "container.rename", fmt.Errorf("name is required"))
		}
		query.Set("name", opts.Name)
		return c.Post(ctx, ContainerRenamePath(id), query, nil, nil)
	case ActionRemove:
		query.Set("force", strconv.FormatBool(opts.Force))
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
