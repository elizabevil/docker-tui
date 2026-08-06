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
//	ActionStop    — Timeout, Signal
//	ActionRestart — Timeout
//	ActionKill    — Signal
//	ActionPause   — no fields used
//	ActionUnpause — no fields used
//	ActionRename  — Name (required)
//	ActionRemove  — Force, RemoveVolumes
//
// Bundling the fields into one struct keeps the dispatcher signature
// readable and makes the call site self-documenting. Use Query to
// produce the URL query values for a given action.
type ContainerActionOptions struct {
	// Timeout applies to stop / restart. Zero means "engine default".
	Timeout time.Duration
	// Force applies to remove.
	Force bool
	// Signal applies to kill. Empty defaults to SIGKILL on POSIX.
	Signal string
	// Name applies to rename.
	Name string
	// RemoveVolumes applies to remove.
	RemoveVolumes bool
}

// Query returns the URL query parameters built from the fields relevant
// to the given action. Zero values are skipped; Force is always emitted
// for ActionRemove because the endpoint treats an absent value as
// implementation-defined.
func (o ContainerActionOptions) Query(action string) url.Values {
	q := make(url.Values)
	switch action {
	case ActionStop:
		o.setTimeout(q)
		o.setSignal(q)
	case ActionRestart:
		o.setTimeout(q)
	case ActionKill:
		o.setSignal(q)
	case ActionRename:
		o.setName(q)
	case ActionRemove:
		o.setForce(q)
		o.setRemoveVolumes(q)
	}
	return q
}

// setTimeout sets the "timeout" parameter when Timeout is non-zero.
func (o ContainerActionOptions) setTimeout(q url.Values) {
	if o.Timeout > 0 {
		q.Set("timeout", strconv.FormatInt(int64(o.Timeout/time.Second), 10))
	}
}

// setSignal sets the "signal" parameter when Signal is non-empty.
func (o ContainerActionOptions) setSignal(q url.Values) {
	if o.Signal != "" {
		q.Set("signal", o.Signal)
	}
}

// setName sets the "name" parameter when Name is non-empty.
func (o ContainerActionOptions) setName(q url.Values) {
	if o.Name != "" {
		q.Set("name", o.Name)
	}
}

// setForce always sets the "force" parameter; zero Force renders "false".
func (o ContainerActionOptions) setForce(q url.Values) {
	q.Set("force", strconv.FormatBool(o.Force))
}

// setRemoveVolumes sets the "v" parameter when RemoveVolumes is true.
func (o ContainerActionOptions) setRemoveVolumes(q url.Values) {
	if o.RemoveVolumes {
		q.Set("v", "1")
	}
}

// ExecuteContainerAction performs a lifecycle action on a container via
// the Podman Libpod REST API. Action is a podman-native action name;
// callers in the runtime package map domain actions to these.
func (c *RESTClient) ExecuteContainerAction(ctx context.Context, id, action string, opts ContainerActionOptions) error {
	switch action {
	case ActionStart:
		return c.Post(ctx, ContainerStartPath(id), nil, nil, nil)
	case ActionStop:
		return c.Post(ctx, ContainerStopPath(id), opts.Query(action), nil, nil)
	case ActionRestart:
		return c.Post(ctx, ContainerRestartPath(id), opts.Query(action), nil, nil)
	case ActionKill:
		return c.Post(ctx, ContainerKillPath(id), opts.Query(action), nil, nil)
	case ActionPause:
		return c.Post(ctx, ContainerPausePath(id), nil, nil, nil)
	case ActionUnpause:
		return c.Post(ctx, ContainerUnpausePath(id), nil, nil, nil)
	case ActionRename:
		if opts.Name == "" {
			return newPodmanError(KindInvalid, "container.rename", fmt.Errorf("name is required"))
		}
		return c.Post(ctx, ContainerRenamePath(id), opts.Query(action), nil, nil)
	case ActionRemove:
		return c.DeleteWithQuery(ctx, ContainerRemovePath(id), opts.Query(action))
	default:
		return newPodmanError(KindInvalid, "container."+action, fmt.Errorf("unsupported action: %s", action))
	}
}