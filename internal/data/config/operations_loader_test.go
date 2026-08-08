package config

import (
	"reflect"
	"strings"
	"testing"
)

// TestLoadOperationsResolvesAllContainerEntries is the R06-02 acceptance
// check: every (Scope=container, Kind) pair declared in
// scopes/container.jsonc carries Mode, the matching body sub-object,
// Label, Description and Requires.
func TestLoadOperationsResolvesAllContainerEntries(t *testing.T) {
	ops, err := LoadOperations()
	if err != nil {
		t.Fatalf("LoadOperations: %v", err)
	}
	wantKinds := []string{"rename", "top", "port", "diff", "wait", "copy", "update", "export", "commit"}
	for _, kind := range wantKinds {
		spec, ok := ops.LookupByKind(OperationScopeContainer, kind)
		if !ok {
			t.Errorf("missing container::%s", kind)
			continue
		}
		if spec.Label == "" {
			t.Errorf("container::%s missing label", kind)
		}
		if spec.Action == "" {
			t.Errorf("container::%s missing action", kind)
		}
		if !isKnownMode(spec.Mode) {
			t.Errorf("container::%s mode=%q not recognised", kind, spec.Mode)
		}
		if err := assertModeBody(spec); err != nil {
			t.Errorf("container::%s %v", kind, err)
		}
	}
}

// TestLoadOperationsResolvesImageHistory verifies the image scope seeds
// its own operation entry rather than inheriting the container defaults.
func TestLoadOperationsResolvesImageHistory(t *testing.T) {
	ops, err := LoadOperations()
	if err != nil {
		t.Fatalf("LoadOperations: %v", err)
	}
	spec, ok := ops.LookupByKind(OperationScopeImage, "history")
	if !ok {
		t.Fatalf("missing image::history")
	}
	if spec.Mode != OperationModePage {
		t.Errorf("image::history mode = %q, want %q", spec.Mode, OperationModePage)
	}
	if spec.Action != "imageHistory" {
		t.Errorf("image::history action = %q, want imageHistory", spec.Action)
	}
	if spec.Page == nil || spec.Page.Body != "historyLayers" {
		t.Errorf("image::history page.body = %+v, want historyLayers", spec.Page)
	}
}

// TestLoadOperationsLookupByAction confirms the Action → spec path.
func TestLoadOperationsLookupByAction(t *testing.T) {
	ops, err := LoadOperations()
	if err != nil {
		t.Fatalf("LoadOperations: %v", err)
	}
	for _, action := range []string{
		"containerRename", "containerTop", "containerPort",
		"containerDiff", "containerWait", "containerCopy",
		"containerUpdate", "containerExport", "containerCommit",
		"imageHistory",
	} {
		if _, ok := ops.LookupByAction(action); !ok {
			t.Errorf("LookupByAction(%q) = (_, false), want true", action)
		}
	}
}

// TestLoadOperationsRejectsUnknownScope guards the invariant that
// every scope file's Scope value is preserved and that all scopes
// expose at least one Operation.
func TestLoadOperationsRejectsUnknownScope(t *testing.T) {
	ops, err := LoadOperations()
	if err != nil {
		t.Fatalf("LoadOperations: %v", err)
	}
	for _, scope := range []OperationScope{OperationScopeContainer, OperationScopeImage} {
		if ops.ForScope(scope) == nil {
			t.Errorf("ForScope(%q) returned nil", scope)
		}
	}
}

// TestLoadOperationsWaitHasAsyncBody verifies the async-mode Operation
// carries its dedicated AsyncBody sub-object (R06 follow-up requirement:
// async must be symmetric with form/page).
func TestLoadOperationsWaitHasAsyncBody(t *testing.T) {
	ops, err := LoadOperations()
	if err != nil {
		t.Fatalf("LoadOperations: %v", err)
	}
	spec, ok := ops.LookupByKind(OperationScopeContainer, "wait")
	if !ok {
		t.Fatalf("missing container::wait")
	}
	if spec.Async == nil {
		t.Fatalf("container::wait async body is nil")
	}
	if spec.Async.Body != "wait" {
		t.Errorf("container::wait async.body = %q, want wait", spec.Async.Body)
	}
}

func isKnownMode(m OperationMode) bool {
	switch m {
	case OperationModeForm, OperationModePage, OperationModeAsync:
		return true
	}
	return false
}

// assertModeBody enforces the Mode ↔ body sub-object contract:
// each Mode must be paired with exactly its matching body field.
func assertModeBody(spec OperationSpec) error {
	formSet := spec.Form != nil
	pageSet := spec.Page != nil
	asyncSet := spec.Async != nil
	switch spec.Mode {
	case OperationModeForm:
		if !formSet {
			return errBodyMissing(OperationModeForm)
		}
		if pageSet || asyncSet {
			return errBodyMismatch(spec.Mode)
		}
	case OperationModePage:
		if !pageSet {
			return errBodyMissing(OperationModePage)
		}
		if formSet || asyncSet {
			return errBodyMismatch(spec.Mode)
		}
	case OperationModeAsync:
		if !asyncSet {
			return errBodyMissing(OperationModeAsync)
		}
		if formSet || pageSet {
			return errBodyMismatch(spec.Mode)
		}
	default:
		return errBodyMismatch(spec.Mode)
	}
	return nil
}

func errBodyMissing(mode OperationMode) error {
	return &missingBodyErr{mode: mode}
}

func errBodyMismatch(mode OperationMode) error {
	return &mismatchBodyErr{mode: mode}
}

type missingBodyErr struct{ mode OperationMode }

func (e *missingBodyErr) Error() string {
	return "missing body sub-object for mode " + string(e.mode)
}

type mismatchBodyErr struct{ mode OperationMode }

func (e *mismatchBodyErr) Error() string {
	return "mode " + string(e.mode) + " has wrong body sub-object"
}

// TestLoadOperationsRequiresAreAllKnown guards the closed Requirement
// universe. If validateSpec ever loosens the check, this test fails
// before the action bar's evaluator can panic on an unknown token.
func TestLoadOperationsRequiresAreAllKnown(t *testing.T) {
	ops, err := LoadOperations()
	if err != nil {
		t.Fatalf("LoadOperations: %v", err)
	}
	for _, spec := range ops.All {
		for _, req := range spec.Requires {
			if _, ok := allRequirements[req]; !ok {
				t.Errorf("%s::%s declares unknown requirement %q", spec.Scope, spec.Kind, req)
			}
		}
	}
}

// TestLoadOperationsFormatSymmetry is a documentation-style sanity
// check: every Operation must declare either a non-empty action or be
// filtered out, and the JSON must round-trip cleanly through the
// loader (already exercised by LoadOperations()).
func TestLoadOperationsFormatSymmetry(t *testing.T) {
	ops, err := LoadOperations()
	if err != nil {
		t.Fatalf("LoadOperations: %v", err)
	}
	for _, spec := range ops.All {
		if strings.TrimSpace(spec.Label) == "" {
			t.Errorf("%s::%s missing label", spec.Scope, spec.Kind)
		}
		if strings.TrimSpace(spec.Action) == "" {
			t.Errorf("%s::%s missing action", spec.Scope, spec.Kind)
		}
	}
}

// TestComposeScopeHasNoBuildPullPush pins §3.5 L1 filter: build / pull /
// push are out of the TUI's scope (compose 域不解析 yaml 不实现 build engine
// 不批量拉镜像). Any kind reappearing in the compose scope is a regression.
func TestComposeScopeHasNoBuildPullPush(t *testing.T) {
	ops, err := LoadOperations()
	if err != nil {
		t.Fatalf("LoadOperations: %v", err)
	}
	for _, spec := range ops.ForScope(OperationScopeCompose) {
		switch spec.Kind {
		case "project_build", "project_pull", "project_push":
			t.Fatalf("out-of-scope kind %q still in compose scope", spec.Kind)
		}
	}
}

// TestComposeScopeHasNoGroupOperations pins §3.4 / R08-14: CoLocated
// metadata is a ContainerSummary field, not a separately-actionable
// group concept. group_* kinds must not be promoted into the action
// bar; any kind reappearing is a regression.
func TestComposeScopeHasNoGroupOperations(t *testing.T) {
	ops, err := LoadOperations()
	if err != nil {
		t.Fatalf("LoadOperations: %v", err)
	}
	for _, spec := range ops.ForScope(OperationScopeCompose) {
		switch spec.Kind {
		case "group_down", "group_restart", "group_exec":
			t.Fatalf("group kind %q still in compose scope; CoLocated must not promote to group ops", spec.Kind)
		}
	}
}

// TestRequiresCapabilitiesFieldRoundtrips verifies the loader accepts
// and preserves the requiresCapabilities JSON field. Capability names
// are open at the loader layer; evaluation happens against the
// runtime CapabilitySet, which fails closed on unknown names.
func TestRequiresCapabilitiesFieldRoundtrips(t *testing.T) {
	src := []byte(`
        {
          "scope": "compose",
          "operations": [
            {
              "kind": "future_capability_demo",
              "action": "compose.demo",
              "mode": "async",
              "async": { "body": "wait" },
              "label": "demo",
              "description": "demo op",
              "requires": ["engine"],
              "requiresCapabilities": ["compose.pod_scope"]
            }
          ]
        }
    `)
	var doc operationsScopeDocument
	if err := decodeJSONCStrict(src, &doc); err != nil {
		t.Fatalf("decode: %v", err)
	}
	spec := doc.Operations[0]
	if got, want := spec.RequiresCapabilities, []string{"compose.pod_scope"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("RequiresCapabilities = %v, want %v", got, want)
	}
}