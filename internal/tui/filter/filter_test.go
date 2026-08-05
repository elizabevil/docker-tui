package filter

import (
	"bytes"
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

func TestController_Open_Compose(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.Mode = state.ModeNormal
	m.Navigation.ActivePanel = state.PanelCompose
	m.Compose.ComposeFocus = 1
	m.Compose.ComposeServiceFilter = "svc"
	m.Compose.ComposeProjectFilter = "proj"

	c := New(m)
	c.Open()

	if m.Navigation.Mode != state.ModeFilter {
		t.Fatalf("Mode=%v, want ModeFilter", m.Navigation.Mode)
	}
	if m.Navigation.FilterInput.Text != "svc" {
		t.Fatalf("FilterInput.Text=%q, want svc (active focus=1)", m.Navigation.FilterInput.Text)
	}

	m.Compose.ComposeFocus = 0
	c.Open()
	if m.Navigation.FilterInput.Text != "proj" {
		t.Fatalf("FilterInput.Text=%q, want proj (active focus=0)", m.Navigation.FilterInput.Text)
	}
}

func TestController_Apply_Compose_Project(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelCompose
	m.Compose.ComposeFocus = 0
	c := New(m)
	c.Apply("proj-x")
	if m.Compose.ComposeProjectFilter != "proj-x" {
		t.Fatalf("ComposeProjectFilter=%q, want proj-x", m.Compose.ComposeProjectFilter)
	}
	if c.Active() != "proj-x" {
		t.Fatalf("Active()=%q, want proj-x", c.Active())
	}
}

func TestController_ApplyCurrent(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelContainers
	m.Navigation.FilterInput.Text = "live"
	c := New(m)
	c.ApplyCurrent()
	if m.Resources.Containers.Filter != "live" {
		t.Fatalf("Container.Filter=%q, want live (from FilterInput.Text)", m.Resources.Containers.Filter)
	}
	if c.Active() != "live" {
		t.Fatalf("Active()=%q, want live", c.Active())
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	src := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	src.Navigation.ActivePanel = state.PanelContainers
	src.Resources.Containers.SetFilter("c-f")
	src.Resources.Images.SetFilter("i-f")
	src.Resources.Volumes.SetFilter("v-f")
	src.Resources.Networks.SetFilter("n-f")
	src.Audit.SetFilter("a-f")
	src.Compose.ComposeServiceFilter = "svc"
	src.Compose.ComposeProjectFilter = "proj"

	var buf bytes.Buffer
	if err := New(src).Save(&buf); err != nil {
		t.Fatalf("Save: %v", err)
	}

	dst := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	if err := Load(dst, &buf); err != nil {
		t.Fatalf("Load: %v", err)
	}

	checks := []struct {
		name string
		got  string
		want string
	}{
		{"Containers", dst.Resources.Containers.Filter, "c-f"},
		{"Images", dst.Resources.Images.Filter, "i-f"},
		{"Volumes", dst.Resources.Volumes.Filter, "v-f"},
		{"Networks", dst.Resources.Networks.Filter, "n-f"},
		{"Audit", dst.Audit.FilterText(), "a-f"},
		{"ComposeService", dst.Compose.ComposeServiceFilter, "svc"},
		{"ComposeProject", dst.Compose.ComposeProjectFilter, "proj"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestSave_NilSafe(t *testing.T) {
	var buf bytes.Buffer
	if err := New(nil).Save(&buf); err != nil {
		t.Fatalf("Save(nil): %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("Save(nil) wrote %d bytes, want 0", buf.Len())
	}
	if err := Load(nil, &buf); err != nil {
		t.Fatalf("Load(nil): %v", err)
	}
}

func TestBannerPrefixFor(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.SetFilter("foo")

	cases := []struct {
		name string
		f    state.TableFilter
		want string
	}{
		{"nil", nil, ""},
		{"active", m.Resources.Containers, "filter: foo | "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := BannerPrefixFor(tc.f); got != tc.want {
				t.Fatalf("BannerPrefixFor()=%q, want %q", got, tc.want)
			}
		})
	}

	ctrl := New(m)
	if got := ctrl.BannerPrefix(); got != "filter: foo | " {
		t.Fatalf("Controller.BannerPrefix()=%q, want %q", got, "filter: foo | ")
	}

	countCases := []struct {
		name     string
		filtered int
		total    int
		want     string
	}{
		{"no_total", 3, 0, "filter: foo | "},
		{"with_count", 3, 42, "filter: foo (3/42) | "},
		{"zero_match", 0, 42, "filter: foo (0/42) | "},
	}
	for _, tc := range countCases {
		t.Run("count_"+tc.name, func(t *testing.T) {
			if got := BannerPrefixForCount(m.Resources.Containers, tc.filtered, tc.total); got != tc.want {
				t.Fatalf("BannerPrefixForCount()=%q, want %q", got, tc.want)
			}
		})
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
