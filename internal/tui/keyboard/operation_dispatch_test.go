package keyboard

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

// operationDispatchModel returns a state.AppModel with one selected
// running container so every Form/Page/Async dispatcher arm has a
// valid target.
func operationDispatchModel() *state.AppModel {
	m := newAppModelWithEngine(newMockEngine())
	m.Navigation.ActivePanel = state.PanelContainers
	m.Resources.Containers.Items = []runtimeapi.ContainerSummary{
		{ID: "one", Name: "api", State: state.ContainerStateRunning},
	}
	return m
}

// TestDispatchOperationRoutesEveryRegistryKeyAction is the R06-03
// acceptance check: for every KeyAction declared in
// operations/registry.jsonc, dispatchOperation must return handled=true.
// If any registry entry fails to route, this test fails — proving the
// keyboard dispatcher is in sync with the JSONC contract.
func TestDispatchOperationRoutesEveryRegistryKeyAction(t *testing.T) {
	ops, err := config.LoadOperations()
	if err != nil {
		t.Fatalf("LoadOperations: %v", err)
	}
	for _, spec := range ops.All {
		m := operationDispatchModel()
		_, _, ok := dispatchOperation(spec.Action, m)
		if !ok {
			t.Errorf("dispatchOperation(%q) = (_, _, false), want true", spec.Action)
		}
	}
}

// TestDispatchOperationIgnoresUnknownKeyAction confirms non-Operation
// actions fall through cleanly to handleAction's switch.
func TestDispatchOperationIgnoresUnknownKeyAction(t *testing.T) {
	m := operationDispatchModel()
	_, _, ok := dispatchOperation("notAnOperation", m)
	if ok {
		t.Errorf("dispatchOperation(\"notAnOperation\") returned handled=true, want false")
	}
}

// TestDispatchOperationFormModesMutateFormKind is the deeper check for
// the four container Form-mode Operations: each must leave
// m.Form.Kind pointing at the matching state.FormKind constant. Rename
// is excluded because it uses the legacy Dialog path (ModeRename), not
// the Form pipeline; that exclusion will go away once R06-04 collapses
// m.Dialog + m.Form into m.Operation.
func TestDispatchOperationFormModesMutateFormKind(t *testing.T) {
	cases := []struct {
		action   keys.KeyAction
		wantKind state.FormKind
	}{
		{keys.ActionContainerCopy, state.FormContainerCopy},
		{keys.ActionContainerUpdate, state.FormContainerUpdate},
		{keys.ActionContainerExport, state.FormContainerExport},
		{keys.ActionContainerCommit, state.FormContainerCommit},
	}
	for _, tc := range cases {
		m := operationDispatchModel()
		_, _, ok := dispatchOperation(string(tc.action), m)
		if !ok {
			t.Errorf("%s: dispatchOperation returned handled=false", tc.action)
			continue
		}
		if got := m.Form.Kind; got != tc.wantKind {
			t.Errorf("%s: Form.Kind = %v, want %v", tc.action, got, tc.wantKind)
		}
	}
}

// TestDispatchOperationPageModesSwitchNavigationMode verifies each
// Page-mode Operation transitions m.Navigation.Mode into the matching
// transient mode so the page renderer picks it up.
func TestDispatchOperationPageModesSwitchNavigationMode(t *testing.T) {
	cases := []struct {
		action   keys.KeyAction
		wantMode state.AppMode
	}{
		{keys.ActionContainerTop, state.ModeTop},
		{keys.ActionContainerPort, state.ModeDetail},
	}
	for _, tc := range cases {
		m := operationDispatchModel()
		_, _, ok := dispatchOperation(string(tc.action), m)
		if !ok {
			t.Errorf("%s: dispatchOperation returned handled=false", tc.action)
			continue
		}
		if got := m.Navigation.Mode; got != tc.wantMode {
			t.Errorf("%s: Navigation.Mode = %v, want %v", tc.action, got, tc.wantMode)
		}
	}
}

// TestDispatchOperationRenameOpensLegacyDialog pins the current
// behaviour: rename is still routed through the legacy ModeRename +
// Dialog path. R06-04 will replace this assertion with a Form-mode
// check once m.Dialog is folded into m.Operation.
func TestDispatchOperationRenameOpensLegacyDialog(t *testing.T) {
	m := operationDispatchModel()
	_, _, ok := dispatchOperation(string(keys.ActionContainerRename), m)
	if !ok {
		t.Fatalf("dispatchOperation(containerRename) returned handled=false")
	}
	if got := m.Navigation.Mode; got != state.ModeRename {
		t.Errorf("Navigation.Mode = %v, want %v (legacy ModeRename)", got, state.ModeRename)
	}
}

// TestDispatchOperationAsyncWaitStopsCleanly verifies the Wait
// dispatcher arm leaves m.ContainerWait.Generation > 0 so the waiter
// can be cancelled on Esc / runtime switch.
func TestDispatchOperationAsyncWaitStopsCleanly(t *testing.T) {
	m := operationDispatchModel()

	before := m.ContainerWait.Generation
	_, _, ok := dispatchOperation(string(keys.ActionContainerWait), m)
	if !ok {
		t.Fatalf("dispatchOperation(containerWait) returned handled=false")
	}
	if m.ContainerWait.Generation <= before {
		t.Errorf("ContainerWait.Generation did not advance (before=%d after=%d)", before, m.ContainerWait.Generation)
	}
}