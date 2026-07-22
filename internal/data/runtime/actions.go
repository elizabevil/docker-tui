package runtime

import (
	"context"
	"time"
)

type ResourceType string

const (
	ResourceContainer ResourceType = "container"
	ResourceImage     ResourceType = "image"
	ResourceVolume    ResourceType = "volume"
	ResourceNetwork   ResourceType = "network"
)

type Action string

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

type ResourceRef struct {
	Type ResourceType
	ID   string
}

type ActionOptions struct {
	Force   bool
	Name    string
	Signal  string
	Timeout time.Duration
}

type ActionResult struct {
	Resource       ResourceRef
	Action         Action
	SpaceReclaimed uint64
}

type ResourceActionService interface {
	Execute(context.Context, ResourceRef, Action, ActionOptions) (ActionResult, error)
}
