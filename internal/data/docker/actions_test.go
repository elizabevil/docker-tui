package docker

import "testing"

func TestRefreshesContainers(t *testing.T) {
	for _, action := range []string{ContainerActionStart, ContainerActionStop, ContainerActionDie, ContainerActionKill, ContainerActionDestroy, ContainerActionCreate} {
		if !RefreshesContainers(action) {
			t.Errorf("%q should refresh containers", action)
		}
	}
	if RefreshesContainers(ContainerActionRestart) {
		t.Error("restart is not emitted as a list event")
	}
}
