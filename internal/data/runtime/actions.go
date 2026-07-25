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

// ActionOptions carries parameters for an action execution. Not all fields
// apply to every action; the adapter interprets only the relevant subset.
//
// TASK-019 advanced-container fields:
//
//	Memory, NanoCPUs, RestartPolicy, RestartMaxRetries — ActionUpdate
//	SourcePath                    — ActionCopy (path inside the container)
//	Destination                   — ActionExport (writer target)
//	Repository, Tag, Comment, Author, Pause — ActionCommit
//	Condition                     — ActionWait (exit condition)
type ActionOptions struct {
	Force            bool
	Name             string
	Signal           string
	Timeout          time.Duration
	Memory           *int64 // ActionUpdate: memory limit (bytes); nil = unset
	NanoCPUs         *int64 // ActionUpdate: nano-CPUs quota; nil = unset
	RestartPolicy    *string // ActionUpdate: "no"|"always"|"unless-stopped"|"on-failure"; nil = unset
	RestartMaxRetries *int  // ActionUpdate: max retries for on-failure; nil = unset
	SourcePath       string // ActionCopy: container path to read
	Destination      string // ActionExport: local destination
	Repository       string // ActionCommit: repository name for the new image
	Tag              string // ActionCommit: tag for the new image
	Comment          string // ActionCommit: commit message
	Author           string // ActionCommit: author string
	Pause            bool   // ActionCommit: pause container during commit
	Condition        string // ActionWait: "not-running" (default) or "next-exit"
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
