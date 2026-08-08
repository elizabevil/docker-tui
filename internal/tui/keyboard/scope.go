package keyboard

import "strings"

const (
	composeVerbStart     = "start"
	composeVerbStop      = "stop"
	composeVerbRestart   = "restart"
	composeVerbDown      = "down"
	composeVerbPause     = "pause"
	composeVerbUnpause   = "unpause"
	composeVerbKill      = "kill"
	composeVerbRm        = "rm"
	composeVerbPrune     = "prune"
	composeVerbPush      = "push"
	composeVerbLogs      = "logs"
	composeVerbTop       = "top"
	composeVerbPort      = "port"
	composeVerbStats     = "stats"
	composeVerbRun       = "run"
	composeVerbExec      = "exec"
	composeVerbGroupDown = "group.down"
)

// composeScopePrefix is the dotted prefix shared by every BatchActioned
// scope emitted by compose-级 actions (R08-04 .. R08-14).
const composeScopePrefix = "compose."

// ComposeScope returns the scope string for a compose-级 verb (e.g.
// "compose.start", "compose.group.down"). The verb argument is one
// of the composeVerb* constants in this file.
func ComposeScope(verb string) string {
	return composeScopePrefix + verb
}

// ContainerBatchScope builds a scope string for a container-级 batch
// operation (e.g. "container.start.web.db"). The trailing segments
// are only appended when the caller wants per-target visibility.
func ContainerBatchScope(scope, verb string, actions []string) string {
	if len(actions) == 0 {
		return scope + "." + verb
	}
	return scope + "." + verb + "." + strings.Join(actions, ".")
}
