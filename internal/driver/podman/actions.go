package podman

// Container lifecycle action constants. These are the verbs accepted by
// the Podman Libpod REST API and used by the transport's switch dispatchers
// in containers_actions.go / images_actions.go.
//
// Vocabulary is intentionally kept as raw string constants (not a separate
// typed enum): the runtime layer adapts via simple string conversion, and
// the simpler shape avoids the "type wrapper that adds no compile-time
// safety" trap. Future migration to a typed dto.Action would only happen
// if a real second source of action vocabulary (e.g. CGO bindings) creates
// a conversion bug at the boundary.
const (
	ActionStart   = "start"
	ActionStop    = "stop"
	ActionRestart = "restart"
	ActionKill    = "kill"
	ActionPause   = "pause"
	ActionUnpause = "unpause"
	ActionRename  = "rename"
	ActionRemove  = "remove"
	ActionCreate  = "create"
	ActionExec    = "exec"
	ActionPrune   = "prune"
	ActionPull    = "pull"
	ActionTag     = "tag"
	ActionPush    = "push"
)
