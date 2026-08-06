package runtime

import (
	"context"
	"time"
)

// ResourceType identifies a class of manageable resource.
type ResourceType string

// Resource type constants used in ResourceRef and action routing.
const (
	ResourceContainer ResourceType = "container"
	ResourceImage     ResourceType = "image"
	ResourceVolume    ResourceType = "volume"
	ResourceNetwork   ResourceType = "network"
	ResourceEvent     ResourceType = "events"
)

// Action identifies a lifecycle operation that can be executed on a resource.
type Action string

// Well-known action identifiers. The set is intentionally minimal; runtime-
// specific actions are routed through ResourceActionService.Execute.
const (
	ActionStart   Action = "start"
	ActionStop    Action = "stop"
	ActionRestart Action = "restart"
	ActionKill    Action = "kill"
	ActionPause   Action = "pause"
	ActionUnpause Action = "unpause"
	ActionRename  Action = "rename"
	ActionRemove  Action = "remove"
	ActionPull    Action = "pull"
	ActionPrune   Action = "prune"
	ActionCreate  Action = "create"
	ActionDestroy Action = "destroy"
	ActionDie     Action = "die"

	// TASK-019: advanced container actions. These have richer payloads
	// (configs, diff streams, file copies) than the lifecycle set above
	// and are dispatched through dedicated ContainerService methods.
	ActionUpdate Action = "update" // apply resource-limit changes
	ActionDiff   Action = "diff"   // inspect filesystem changes vs image
	ActionExport Action = "export" // produce a tar of the container fs
	ActionCommit Action = "commit" // snapshot the container as a new image
	ActionWait   Action = "wait"   // block until the container exits
	ActionCopy   Action = "copy"   // extract a file from the container
)

// ResourceRef identifies a specific resource by type and ID.
type ResourceRef struct {
	Type ResourceType
	ID   string
}

// LifecycleOptions is the parameter set for the simple container / image
// lifecycle actions (start, stop, restart, kill, pause, unpause, rename,
// remove, pull, prune, create, destroy). It is a flat value (not a pointer)
// because every field has a meaningful zero value for callers that do
// not care about it.
type LifecycleOptions struct {
	// Force applies to remove/kill. False is the safe default.
	Force bool
	// Name applies to rename. Empty means "no rename requested".
	Name string
	// Signal applies to kill / stop. Empty defaults to the engine default
	// (SIGKILL for kill, SIGTERM for stop).
	Signal string
	// Timeout applies to stop / restart. Zero means "engine default".
	Timeout time.Duration
	// RemoveVolumes applies to ActionRemove on a container. Docker only.
	RemoveVolumes bool
	// RemoveLinks applies to ActionRemove on a container. Docker only.
	RemoveLinks bool
	// PruneChildren applies to ActionRemove on an image. Docker only.
	PruneChildren bool
	// Platforms applies to ActionRemove on an image. Docker only; empty = all.
	Platforms []string
}

// UpdateOptions carries parameters for ActionUpdate. Each field is a
// pointer so the caller can distinguish "do not change" (nil) from "set
// to zero". Fields with engine-side zero values (e.g. policy "no") are
// encoded as plain strings; the adapter maps them to SDK enums.
type UpdateOptions struct {
	Memory            *int64  // bytes; nil = leave unchanged
	NanoCPUs          *int64  // nil = leave unchanged
	RestartPolicy     *string // "no"|"always"|"unless-stopped"|"on-failure"; nil = leave unchanged
	RestartMaxRetries *int    // only meaningful with RestartPolicy == "on-failure"
}

// DiffOptions is empty for now — ContainerDiff takes no parameters today
// but the type is reserved for symmetry with the other action families
// and for future filters (e.g. by path glob).
type DiffOptions struct{}

// ExportOptions identifies where the exported tar should be written.
// Destination is a host path; the adapter is responsible for opening it.
type ExportOptions struct {
	Destination string
}

// CommitOptions configures a container-to-image commit. Pause is the
// container pause flag (true = pause while committing to get a clean
// snapshot). The other fields map directly to Docker / Podman labels.
type CommitOptions struct {
	Repository string // image repository name (e.g. "myapp"); empty = engine default
	Tag        string // image tag; empty = "latest"
	Comment    string
	Author     string
	Pause      bool
}

// WaitOptions carries parameters for ActionWait. Condition matches
// Docker / Podman's documented values:
//   - "" or "not-running" — default; blocks until the container exits
//   - "next-exit" — blocks until the next time the container exits
//   - "removed" — blocks until the container is removed
type WaitOptions struct {
	Condition string
}

// CopyOptions is the parameter for ActionCopy (extract a file from a
// container). SourcePath is the container-side path; the adapter returns
// an io.ReadCloser containing a tar archive.
type CopyOptions struct {
	SourcePath string
}

// ActionOptions is a discriminated union: for each call, exactly one
// sub-payload (Lifecycle or one of the six typed pointers) is populated,
// matching the Action passed to Execute.
//
// Lifecycle actions (start, stop, restart, kill, pause, unpause, rename,
// remove, pull, prune, create, destroy) populate Lifecycle directly:
//
//	opts := ActionOptions{Lifecycle: LifecycleOptions{Force: true}}
//	svc.Execute(ctx, ref, ActionRemove, opts)
//
// Advanced actions (update, diff, export, commit, wait, copy) populate
// the corresponding typed pointer:
//
//	opts := ActionOptions{Update: &UpdateOptions{Memory: &mem}}
//	svc.Execute(ctx, ref, ActionUpdate, opts)
//
// The adapter uses a type switch on the action and reads only the
// matching sub-payload. Fields that do not match the action are ignored,
// which makes the type system self-documenting and removes the
// "pause the container" vs "pause while committing" ambiguity.
type ActionOptions struct {
	Lifecycle LifecycleOptions

	Update *UpdateOptions
	Diff   *DiffOptions
	Export *ExportOptions
	Commit *CommitOptions
	Wait   *WaitOptions
	Copy   *CopyOptions
}

// ActionResult is the outcome of a successful action execution.
type ActionResult struct {
	Resource       ResourceRef
	Action         Action
	SpaceReclaimed uint64
}

// ResourceActionService executes lifecycle actions across resource types.
// Adapters translate the generic action into SDK-specific calls.
type ResourceActionService interface {
	Execute(context.Context, ResourceRef, Action, ActionOptions) (ActionResult, error)
}

// Operation composes the diagnostic operation identifier for a runtime
// operation, in the canonical "<resource>.<verb>" form used by Error.Operation.
func Operation(resource ResourceType, verb string) string {
	return string(resource) + "." + verb
}

// RefreshesContainers reports actions that should trigger a container list
// refresh. Restart is excluded because it is not emitted as a discrete list
// event (start/stop pair covers it).
func RefreshesContainers(action Action) bool {
	switch action {
	case ActionStart, ActionStop, ActionDie,
		ActionKill, ActionPause, ActionUnpause,
		ActionRename, ActionDestroy, ActionCreate:
		return true
	default:
		return false
	}
}
