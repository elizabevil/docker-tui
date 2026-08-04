package keyboard

import "strings"

const (
	composeVerbStart = "start"
	composeVerbStop  = "stop"
	composeVerbDown  = "down"
)

func ContainerBatchScope(scope, verb string, actions []string) string {
	if len(actions) == 0 {
		return scope + "." + verb
	}
	return scope + "." + verb + "." + strings.Join(actions, ".")
}

func ComposeScope(verb string) string {
	return "compose." + verb
}
