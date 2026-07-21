package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestRuntimeSelectorNavigationAndCancel(t *testing.T) {
	p := dockerclient.NewPool()
	p.AddHost(dockerclient.HostEntry{Name: "docker", Host: "unix:///docker.sock", Runtime: "docker"})
	p.AddHost(dockerclient.HostEntry{Name: "podman", Host: "unix:///podman.sock", Runtime: "podman"})
	m := state.NewAppModel(config.DefaultConfig(), nil, "test")
	m.Pool = p
	if _, _ = openRuntimeSelector(m); m.Mode != state.ModeRuntimeSelect {
		t.Fatal("selector did not open")
	}
	if m.RuntimeSelectorCursor != 0 {
		t.Fatal("unexpected initial cursor")
	}
	if _, _ = handleRuntimeSelectorKey(keys.KeyDown, m); m.RuntimeSelectorCursor != 1 {
		t.Fatal("down did not move cursor")
	}
	if _, _ = handleRuntimeSelectorKey(keys.KeyEsc, m); m.Mode != state.ModeNormal {
		t.Fatal("esc did not cancel")
	}
}

func TestRuntimeSelectorDisabledForHostOverride(t *testing.T) {
	m := &state.AppModel{RuntimeSelectorDisabled: true}
	updated, cmd := openRuntimeSelector(m)
	if cmd != nil || updated.Mode != state.ModeNormal || updated.ToastMessage == "" {
		t.Fatal("disabled selector changed state unexpectedly")
	}
}
