package filter

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func TestController_Active_Empty(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	c := New(m)
	if got := c.Active(); got != "" {
		t.Fatalf("Active()=%q, want empty", got)
	}
	if c.HasActive() {
		t.Fatal("HasActive()=true, want false")
	}
}

func TestController_OpenAndApply(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelContainers
	c := New(m)

	c.Open()
	if !c.IsInputMode() {
		t.Fatal("IsInputMode()=false after Open, want true")
	}
	if m.Navigation.Mode != state.ModeFilter {
		t.Fatalf("Mode=%v after Open, want ModeFilter", m.Navigation.Mode)
	}

	c.Apply("foo")
	if m.Resources.Containers.Filter != "foo" {
		t.Fatalf("Container.Filter=%q, want foo", m.Resources.Containers.Filter)
	}
	if c.Active() != "foo" {
		t.Fatalf("Active()=%q, want foo", c.Active())
	}
	if !c.HasActive() {
		t.Fatal("HasActive()=false, want true")
	}
}

func TestController_ClampCursor(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelContainers
	// simulate stale cursor from before filter
	m.Resources.Containers.Cursor = 5
	m.Resources.Containers.ViewOffset = 3

	c := New(m)
	c.ClampCursor()

	if m.Resources.Containers.Cursor != 0 || m.Resources.Containers.ViewOffset != 0 {
		t.Fatalf("cursor/offset not reset: cursor=%d offset=%d",
			m.Resources.Containers.Cursor, m.Resources.Containers.ViewOffset)
	}
}

func TestController_Clear(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelContainers
	c := New(m)
	c.Apply("bar")
	c.ClampCursor()
	m.Resources.Containers.Cursor = 4
	m.Resources.Containers.ViewOffset = 2

	c.Clear()

	if m.Resources.Containers.Filter != "" {
		t.Fatalf("Filter=%q after Clear, want empty", m.Resources.Containers.Filter)
	}
	if c.HasActive() {
		t.Fatal("HasActive()=true after Clear, want false")
	}
	if m.Navigation.Mode == state.ModeFilter {
		t.Fatal("Clear changed Mode to non-Filter; expected to stay put")
	}
}

func TestController_Close(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelContainers
	c := New(m)
	c.Open()
	c.Apply("baz")
	m.Navigation.FilterExitPending = true // simulate pending double-Esc

	c.Close()

	if m.Navigation.Mode == state.ModeFilter {
		t.Fatal("Mode still Filter after Close")
	}
	if m.Navigation.FilterExitPending {
		t.Fatal("FilterExitPending still true after Close")
	}
	if m.Resources.Containers.Filter != "" {
		t.Fatalf("Filter=%q after Close, want empty", m.Resources.Containers.Filter)
	}
}

func TestController_Apply_Compose(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelCompose
	m.Compose.ComposeFocus = 1
	c := New(m)

	c.Apply("svc-x")
	if m.Compose.ComposeServiceFilter != "svc-x" {
		t.Fatalf("ComposeServiceFilter=%q, want svc-x", m.Compose.ComposeServiceFilter)
	}
	if c.Active() != "svc-x" {
		t.Fatalf("Active()=%q, want svc-x", c.Active())
	}
}

func TestController_NilSafe(t *testing.T) {
	c := New(nil)
	// none of these should panic
	c.Open()
	c.Apply("x")
	c.ClampCursor()
	c.Clear()
	c.Close()
	if c.IsInputMode() {
		t.Fatal("IsInputMode()=true for nil model")
	}
	if c.HasActive() {
		t.Fatal("HasActive()=true for nil model")
	}
}

func TestController_HasActive_Compose(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelCompose
	m.Compose.ComposeFocus = 1
	m.Compose.ComposeServiceFilter = "svc"
	c := New(m)
	if !c.HasActive() {
		t.Fatal("HasActive()=false, want true for compose with filter")
	}
	if c.Active() != "svc" {
		t.Fatalf("Active()=%q, want svc", c.Active())
	}
}

func TestController_Clear_Compose(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelCompose
	m.Compose.ComposeFocus = 1
	m.Compose.ComposeServiceFilter = "svc"
	c := New(m)
	c.Clear()
	if m.Compose.ComposeServiceFilter != "" {
		t.Fatalf("ComposeServiceFilter=%q after Clear, want empty", m.Compose.ComposeServiceFilter)
	}
}

func TestController_ClearAll(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.SetFilter("c-filter")
	m.Resources.Images.SetFilter("i-filter")
	m.Resources.Volumes.SetFilter("v-filter")
	m.Resources.Networks.SetFilter("n-filter")
	m.Audit.SetFilter("a-filter")
	m.Compose.ComposeServiceFilter = "svc"
	m.Compose.ComposeProjectFilter = "proj"
	m.Resources.Containers.Cursor = 5
	m.Resources.Images.Cursor = 3
	m.Navigation.FilterInput.Text = "stale"

	c := New(m)
	c.ClearAll()

	checks := []struct {
		name  string
		got   string
	}{
		{"Containers", m.Resources.Containers.Filter},
		{"Images", m.Resources.Images.Filter},
		{"Volumes", m.Resources.Volumes.Filter},
		{"Networks", m.Resources.Networks.Filter},
		{"Audit", m.Audit.FilterText()},
	}
	for _, c := range checks {
		if c.got != "" {
			t.Errorf("%s.Filter=%q after ClearAll, want empty", c.name, c.got)
		}
	}
	if m.Compose.ComposeServiceFilter != "" {
		t.Errorf("ComposeServiceFilter=%q, want empty", m.Compose.ComposeServiceFilter)
	}
	if m.Compose.ComposeProjectFilter != "" {
		t.Errorf("ComposeProjectFilter=%q, want empty", m.Compose.ComposeProjectFilter)
	}
	if m.Resources.Containers.Cursor != 0 || m.Resources.Images.Cursor != 0 {
		t.Errorf("cursors not reset: containers=%d images=%d",
			m.Resources.Containers.Cursor, m.Resources.Images.Cursor)
	}
	if m.Navigation.FilterInput.Text != "" {
		t.Errorf("FilterInput.Text=%q, want empty", m.Navigation.FilterInput.Text)
	}
	if c.HasActive() {
		t.Error("HasActive()=true after ClearAll")
	}
}
