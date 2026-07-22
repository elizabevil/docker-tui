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
