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
)

// ResourceRef identifies a specific resource by type and ID.
type ResourceRef struct {
	Type ResourceType
	ID   string
}

// ActionOptions carries parameters for an action execution. Not all fields
// apply to every action; the adapter interprets only the relevant subset.
type ActionOptions struct {
	Force   bool
	Name    string
	Signal  string
	Timeout time.Duration
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
