package docker

// Container operation names are shared by commands, event handling and audit
// projection. Keeping them here prevents UI code from drifting on literals.
const (
	ContainerActionStart   = "start"
	ContainerActionStop    = "stop"
	ContainerActionRestart = "restart"
	ContainerActionKill    = "kill"
	ContainerActionRemove  = "remove"
	ContainerActionCreate  = "create"
	ContainerActionDestroy = "destroy"
	ContainerActionDie     = "die"
)

// RefreshesContainers reports events that can change the container list.
func RefreshesContainers(action string) bool {
	switch action {
	case ContainerActionStart, ContainerActionStop, ContainerActionDie,
		ContainerActionKill, ContainerActionDestroy, ContainerActionCreate:
		return true
	default:
		return false
	}
}
