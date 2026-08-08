package actionbar

import (
	"context"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime/docker"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// capEngine is a minimal runtimeapi.Engine stub for capability-filter
// tests. Only Capabilities carries test data; the rest return zero
// values because buildItemsForScope / VisibleItems never invoke them.
type capEngine struct {
	caps runtimeapi.CapabilitySet
}

func (e *capEngine) Identity() runtimeapi.Identity                       { return runtimeapi.Identity{} }
func (e *capEngine) Capabilities() runtimeapi.CapabilitySet             { return e.caps }
func (e *capEngine) PingContext(context.Context) error                  { return nil }
func (e *capEngine) Containers() runtimeapi.ContainerService             { return nil }
func (e *capEngine) Volumes() runtimeapi.VolumeService                  { return nil }
func (e *capEngine) Networks() runtimeapi.NetworkService                { return nil }
func (e *capEngine) Images() runtimeapi.ImageService                    { return nil }
func (e *capEngine) ImageTransfers() runtimeapi.ImageTransferService    { return nil }
func (e *capEngine) Actions() runtimeapi.ResourceActionService          { return nil }
func (e *capEngine) Exec() runtimeapi.ExecService                       { return nil }
func (e *capEngine) Events() runtimeapi.EventService                    { return nil }
func (e *capEngine) Compose() runtimeapi.ComposeService                 { return nil }
func (e *capEngine) Pods() runtimeapi.PodService                        { return nil }
func (e *capEngine) Close() error                                       { return nil }

func TestContainersExposeOnlyComplexActions(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), &dockerclient.Client{}, "test")
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{ID: "one", Name: "api", State: state.ContainerStateRunning}}
	items := VisibleItems(m)
	want := []keys.KeyAction{keys.ActionContainerRename, keys.ActionContainerTop, keys.ActionContainerPort, keys.ActionContainerDiff, keys.ActionContainerWait, keys.ActionContainerCopy, keys.ActionContainerUpdate, keys.ActionContainerExport, keys.ActionContainerCommit}
	if len(items) != len(want) {
		t.Fatalf("items = %#v", items)
	}
	for index, action := range want {
		if items[index].Action != action || items[index].Key != "" || items[index].Disabled {
			t.Fatalf("item %d = %#v", index, items[index])
		}
	}
}

func TestDirectAndGlobalActionsAreExcluded(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	// Image History is a no-shortcut complex action; panels with no
	// operation scope (audit today) must surface zero items so a stray
	// shortcut cannot leak through.
	m.Navigation.ActivePanel = state.PanelAudit
	if items := VisibleItems(m); len(items) != 0 {
		t.Fatalf("panel %v contains direct/global actions: %#v", state.PanelAudit, items)
	}
}

// TestScopeForPanelMapsPanels locks the per-panel → OperationScope
// dispatch table. Adding a panel without updating this map would
// silently drop the Action Bar for that panel.
func TestScopeForPanelMapsPanels(t *testing.T) {
	cases := []struct {
		panel state.PanelType
		want  config.OperationScope
	}{
		{state.PanelContainers, config.OperationScopeContainer},
		{state.PanelImages, config.OperationScopeImage},
		{state.PanelVolumes, config.OperationScopeVolume},
		{state.PanelNetworks, config.OperationScopeNetwork},
		{state.PanelCompose, config.OperationScopeCompose},
	}
	for _, tc := range cases {
		got, ok := scopeForPanel(tc.panel)
		if !ok || got != tc.want {
			t.Fatalf("scopeForPanel(%v) = %q, %v; want %q", tc.panel, got, ok, tc.want)
		}
	}
	if _, ok := scopeForPanel(state.PanelAudit); ok {
		t.Fatalf("scopeForPanel(audit) should return false")
	}
}

func TestContainerActionAvailabilityAndFiltering(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	items := VisibleItems(m)
	for _, item := range items {
		if !item.Disabled {
			t.Fatalf("missing container action should be disabled: %#v", item)
		}
	}
	m.Navigation.ActionBar.Filter = "rename"
	items = VisibleItems(m)
	if len(items) != 1 || items[0].Action != keys.ActionContainerRename {
		t.Fatalf("filtered items = %#v", items)
	}
}

// TestBuildItemsForScopeSkipsInActionBarFalse pins the C04 contract
// directly on the filter: Operations declaring inActionBar=false must
// never surface as action-bar items, even though they live in the same
// scope slice. ScopeForPanel already keeps Volume/Network panels out of
// the action bar; this unit test locks the filter itself so the
// two-layer guard cannot silently regress.
func TestBuildItemsForScopeSkipsInActionBarFalse(t *testing.T) {
	ops := &config.Operations{
		ByScope: map[config.OperationScope][]config.OperationSpec{
			config.OperationScopeContainer: {
				{Action: "containerRename", Label: "Rename", InActionBar: true},
				{Action: "containerTop", Label: "Top", InActionBar: false},
			},
		},
	}
	m := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	items := buildItemsForScope(m, ops, config.OperationScopeContainer)
	if len(items) != 1 || items[0].Action != keys.ActionContainerRename {
		t.Fatalf("items = %#v, want exactly [containerRename]", items)
	}
}

// TestBuildItemsForScopeHidesByCapability pins the L2 capability gate:
// Operations declaring requiresCapabilities must be hidden when the
// live Engine lacks the listed capability (fail closed), and visible
// when every required capability is supported.
func TestBuildItemsForScopeHidesByCapability(t *testing.T) {
	spec := config.OperationSpec{
		Scope:               config.OperationScopeCompose,
		Kind:                "demo",
		Action:              "compose.demo",
		Mode:                config.OperationModeAsync,
		Async:               &config.AsyncBody{Body: "wait"},
		Label:               "Demo",
		Description:         "demo",
		InActionBar:         true,
		RequiresCapabilities: []string{"compose.pod_scope"},
	}
	ops := &config.Operations{
		ByScope: map[config.OperationScope][]config.OperationSpec{
			config.OperationScopeCompose: {spec},
		},
	}

	// nil engine + non-empty requirement → hidden (fail closed)
	m1 := state.NewAppModel(config.DefaultAppConfig(), nil, "test")
	if got := buildItemsForScope(m1, ops, config.OperationScopeCompose); len(got) != 0 {
		t.Fatalf("nil engine must hide items with capability requirements, got %#v", got)
	}

	// engine lacking the capability → hidden
	m2 := state.NewAppModel(config.DefaultAppConfig(), &capEngine{}, "test")
	if got := buildItemsForScope(m2, ops, config.OperationScopeCompose); len(got) != 0 {
		t.Fatalf("engine without capability must hide item, got %#v", got)
	}

	// engine with the capability → visible
	m3 := state.NewAppModel(config.DefaultAppConfig(), &capEngine{
		caps: runtimeapi.CapabilitySet{
			runtimeapi.CapabilityComposePodScope: runtimeapi.CapabilityInfo{Support: runtimeapi.Available},
		},
	}, "test")
	if got := buildItemsForScope(m3, ops, config.OperationScopeCompose); len(got) != 1 {
		t.Fatalf("engine with capability must show item, got %#v", got)
	}
}

func TestImagesExposeHistoryWhenEligible(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), &dockerclient.Client{}, "test")
	m.Navigation.ActivePanel = state.PanelImages
	m.Resources.Images.Items = []runtimeapi.ImageSummary{{ID: "sha256:one", RepoTags: []string{"nginx:latest"}}}
	items := VisibleItems(m)
	wantActions := []keys.KeyAction{keys.ActionImageHistory, keys.ActionImagePrune}
	if len(items) != len(wantActions) {
		t.Fatalf("image actions = %#v, want %d items", items, len(wantActions))
	}
	for i, action := range wantActions {
		if items[i].Action != action || items[i].Disabled {
			t.Fatalf("item %d = %#v, want action=%q disabled=false", i, items[i], action)
		}
	}
	m.Resources.Images.Items[0].IsManifest = true
	if items = VisibleItems(m); len(items) != len(wantActions) {
		t.Fatalf("image actions after manifest: %#v", items)
	}
	if !items[0].Disabled {
		t.Fatalf("manifest history should be disabled: %#v", items[0])
	}
	if items[1].Disabled {
		t.Fatalf("imagePrune should remain enabled with manifest image: %#v", items[1])
	}
}

// TestComposeRequirementsEvaluateProjectAndService verifies R08-11 F4
// against an AppModel seeded with a small project that has one running
// + one stopped service + one tagged image.
func TestComposeRequirementsEvaluateProjectAndService(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), &dockerclient.Client{}, "test")
	m.Navigation.ActivePanel = state.PanelCompose
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{
		{
			ID:             "web1",
			Name:           "demo_web_1",
			State:          state.ContainerStateRunning,
			ComposeProject: "demo",
			ComposeService: "web",
			Image:          "nginx:1.21",
			Labels: map[string]string{
				runtimeapi.ComposeLabelImage: "nginx:1.21",
			},
			PortBindings: []runtimeapi.PortBinding{{ContainerPort: 80, Protocol: "tcp", HostPort: 18080}},
		},
		{
			ID:             "db1",
			Name:           "demo_db_1",
			State:          state.ContainerStateExited,
			ComposeProject: "demo",
			ComposeService: "db",
			Image:          "postgres:16",
		},
	}

	cases := []struct {
		req  config.Requirement
		want bool
	}{
		{config.RequirementComposeProject, true},
		{config.RequirementComposeRunning, true},
		{config.RequirementComposePaused, false},
		{config.RequirementComposeStopped, false},
		{config.RequirementComposeTagged, true},
		{config.RequirementComposeHasPorts, true},
		{config.RequirementComposeService, false},
	}
	for _, tc := range cases {
		if got := positiveSatisfied(tc.req, m); got != tc.want {
			t.Errorf("positiveSatisfied(%q) = %v, want %v", tc.req, got, tc.want)
		}
	}

	// Stop the running container: Running should drop, Stopped should
	// pass (only stopped containers remain). HasPorts stays true
	// because port-bindings are declared at create time and remain on
	// the summary even after stop — the runtime will resolve the
	// static binding when the user invokes Compose port.
	m.Resources.Containers.Items[0].State = state.ContainerStateExited
	if got := positiveSatisfied(config.RequirementComposeRunning, m); got {
		t.Errorf("running should be false after stopping the only running container")
	}
	if got := positiveSatisfied(config.RequirementComposeStopped, m); !got {
		t.Errorf("stopped should be true after stopping the only running container")
	}
	if got := positiveSatisfied(config.RequirementComposeHasPorts, m); !got {
		t.Errorf("has_ports should still be true (port bindings survive stop)")
	}

	// Move focus to the services pane with a service selected.
	m.Compose.ComposeFocus = 1
	if got := positiveSatisfied(config.RequirementComposeService, m); !got {
		t.Errorf("compose_service should be true with services pane focused")
	}

	// No compose project → every project-scoped token fails.
	m.Resources.Containers.Items = nil
	m.Navigation.ActivePanel = state.PanelCompose
	if got := positiveSatisfied(config.RequirementComposeProject, m); got {
		t.Errorf("compose_project should be false with no containers")
	}
	if got := positiveSatisfied(config.RequirementComposeRunning, m); got {
		t.Errorf("compose_running should be false with no containers")
	}
}

// TestComposeOperationsSurfaceInActionBar pins R08-11 F1: pressing `;`
// on the compose panel surfaces the registered compose operations in the
// expected enabled / disabled shape.
func TestComposeOperationsSurfaceInActionBar(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), &dockerclient.Client{}, "test")
	m.Navigation.ActivePanel = state.PanelCompose
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{
		{
			ID: "web1", State: state.ContainerStateRunning,
			ComposeProject: "demo", ComposeService: "web", Image: "nginx:1.21",
			Labels:        map[string]string{runtimeapi.ComposeLabelImage: "nginx:1.21"},
			PortBindings:  []runtimeapi.PortBinding{{ContainerPort: 80, Protocol: "tcp", HostPort: 18080}},
		},
		{
			ID: "db1", State: state.ContainerStateExited,
			ComposeProject: "demo", ComposeService: "db", Image: "postgres:16",
		},
	}
	items := VisibleItems(m)
	if len(items) == 0 {
		t.Fatalf("compose panel action bar is empty; expected ≥1 items")
	}

	// Spot-check the three semantic groups by action string.
	// Test fixture: 1 running web (with ports, tagged image) + 1
	// exited db. Therefore compose_running=true, compose_stopped=
	// false (running present), compose_paused=false, compose_has_ports=
	// true, compose_tagged=true.
	wantEnabled := []keys.KeyAction{
		keys.ActionComposeProjectStop,    // has running container
		keys.ActionComposeProjectRestart,
		keys.ActionComposeProjectLogs,
		keys.ActionComposeProjectStats,
		keys.ActionComposeProjectDown,
		keys.ActionComposeProjectBuild,
		keys.ActionComposeProjectPull,
		keys.ActionComposeProjectPrune,
		keys.ActionComposeProjectEvents,
		keys.ActionComposeProjectTop,     // has running container
		keys.ActionComposeProjectPort,    // has ports
		keys.ActionComposeProjectPause,   // has running container
		keys.ActionComposeProjectKill,    // has running container
		keys.ActionComposeProjectPush,    // tagged service
	}
	wantDisabled := []keys.KeyAction{
		keys.ActionComposeProjectStart,   // no stopped container (running present so compose_stopped=false)
		keys.ActionComposeProjectUnpause, // no paused container
		keys.ActionComposeProjectRm,      // no stopped-only state (running present)
	}
	itemsByAction := make(map[keys.KeyAction]ActionItem)
	for _, it := range items {
		itemsByAction[it.Action] = it
	}
	for _, a := range wantEnabled {
		it, ok := itemsByAction[a]
		if !ok {
			t.Errorf("action bar missing expected enabled action %q", a)
			continue
		}
		if it.Disabled {
			t.Errorf("action %q should be enabled, got %#v", a, it)
		}
	}
	for _, a := range wantDisabled {
		it, ok := itemsByAction[a]
		if !ok {
			t.Errorf("action bar missing expected disabled action %q", a)
			continue
		}
		if !it.Disabled {
			t.Errorf("action %q should be disabled, got %#v", a, it)
		}
	}
}

// TestComposeTaggedRejectsLatest verifies the tagged heuristic rejects
// both the bare image (no `:`) and the `:latest` form.
func TestComposeTaggedRejectsLatest(t *testing.T) {
	m := state.NewAppModel(config.DefaultAppConfig(), &dockerclient.Client{}, "test")
	m.Navigation.ActivePanel = state.PanelCompose
	cases := []struct {
		image string
		want  bool
	}{
		{"nginx", false},
		{"nginx:latest", false},
		{"nginx:1.21", true},
		{"registry.example.com/team/app:1.0", true},
	}
	for _, tc := range cases {
		m.Resources.Containers.Items = []runtimeapi.ContainerSummary{{
			ID:             "c",
			State:          state.ContainerStateRunning,
			ComposeProject: "demo",
			ComposeService: "svc",
			Image:          tc.image,
			Labels: map[string]string{
				runtimeapi.ComposeLabelImage: tc.image,
			},
		}}
		if got := positiveSatisfied(config.RequirementComposeTagged, m); got != tc.want {
			t.Errorf("compose_tagged(%q) = %v, want %v", tc.image, got, tc.want)
		}
	}
}
