package config

// OperationScope is the resource an Operation acts on. The set is open
// (any string is accepted) but the loader cross-checks against the
// `theme.action.<scope>` theme contract so a scope without a theme
// section renders as a transparent default. Defined values are seeded
// from `theme.action.<scope>` in `theme.go`.
type OperationScope string

const (
	OperationScopeContainer OperationScope = "container"
	OperationScopeImage     OperationScope = "image"
	// OperationScopeVolume    OperationScope = "volume"  // future
	// OperationScopeNetwork   OperationScope = "network" // future
)

// OperationMode is the rendering mode of an Operation's body. Each
// mode carries its own typed body sub-object (Form / Page / Async) so
// the JSON schema is self-documenting: a reader sees exactly which
// extra fields a given mode requires without consulting Go code.
type OperationMode string

const (
	OperationModeForm  OperationMode = "form"
	OperationModePage  OperationMode = "page"
	OperationModeAsync OperationMode = "async"
)

// FormBody is the per-mode body for mode=form. Its `Kind` value must
// match a `state.FormKind` constant name (without the package prefix).
type FormBody struct {
	// Kind is the FormKind identifier (e.g. "containerCopy",
	// "imageSave"). The dispatcher switches on this string.
	Kind string `json:"kind"`
}

// PageBody is the per-mode body for mode=page. Its `Body` value names
// the page renderer (e.g. "processes", "portBindings",
// "historyLayers"). The dispatcher switches on this string.
type PageBody struct {
	Body string `json:"body"`
}

// AsyncBody is the per-mode body for mode=async. Its `Body` value
// names the async renderer (e.g. "wait", "imagePrune"). The dispatcher
// switches on this string. Mode=async currently has one consumer but
// the body sub-object keeps the schema symmetric with form/page so
// future async operations (imagePrune, volumePrune, ...) plug in
// without a schema change.
type AsyncBody struct {
	Body string `json:"body"`
}

// Requirement is one named predicate the loader / action bar evaluates
// against the AppModel. The set is closed; unknown values fail at load
// time so a typo in JSONC surfaces immediately.
//
// Requirements appear in two fields with opposite semantics:
//
//	Requires    — AND-combined positives. Operation enabled iff all
//	              satisfied.
//	DisabledWhen — OR-combined negatives. Operation disabled if any
//	                satisfied.
type Requirement string

const (
	RequirementEngine    Requirement = "engine"
	RequirementContainer Requirement = "container"
	RequirementImage     Requirement = "image"
	RequirementRunning   Requirement = "running"
	RequirementManifest  Requirement = "manifest"
)

// allRequirements is the canonical Requirement universe. The loader
// rejects any Requirement outside this set so the action bar's
// evaluator stays a closed switch.
var allRequirements = map[Requirement]struct{}{
	RequirementEngine:    {},
	RequirementContainer: {},
	RequirementImage:     {},
	RequirementRunning:   {},
	RequirementManifest:  {},
}

// OperationSpec is one entry in the Operations registry. The struct
// is intentionally small and mode-driven: the three optional
// sub-objects (Form / Page / Async) make the JSON schema self-checking
// because exactly one of them must be set based on Mode.
type OperationSpec struct {
	// Kind is the Operation's identifier within its scope (e.g. "rename",
	// "top", "copy"). Combined with Scope it forms a globally unique
	// "<scope>::<kind>" key.
	Kind string `json:"kind"`

	// Scope is the resource scope the Operation acts on. Repeated here
	// (in addition to the per-scope file) so a single Operation object
	// is fully self-describing — the file structure is the only reason
	// this field exists at all.
	Scope OperationScope `json:"scope"`

	// Action is the bound KeyAction identifier (matches keys.ActionXxx
	// constants). It is the bridge to the existing keyboard layer; the
	// dispatcher in `internal/tui/keyboard/operation_dispatch.go`
	// resolves `Action` → `OperationSpec` → Mode-specific handler.
	Action string `json:"action"`

	// Mode drives which of Form / Page / Async is populated and how the
	// body is rendered.
	Mode OperationMode `json:"mode"`

	// Form is required when Mode = form.
	Form *FormBody `json:"form,omitempty"`

	// Page is required when Mode = page.
	Page *PageBody `json:"page,omitempty"`

	// Async is required when Mode = async. Today only "wait" is
	// defined; future async Operations (imagePrune, volumePrune, ...)
	// add their own Body values.
	Async *AsyncBody `json:"async,omitempty"`

	// Label is the human-readable action-bar entry text.
	Label string `json:"label"`

	// Description is the secondary line shown in the action-bar entry
	// and used as the fuzzy-search target.
	Description string `json:"description"`

	// Requires is a list of positive predicates (AND-combined). The
	// Operation is enabled only when every requirement is satisfied.
	Requires []Requirement `json:"requires,omitempty"`

	// DisabledWhen is a list of negative predicates (OR-combined). If
	// any predicate matches, the Operation is disabled. Useful for
	// "disable if state is X" rules like "manifest image has no
	// history".
	DisabledWhen []Requirement `json:"disabledWhen,omitempty"`
}

// Operations is the resolved, in-memory form of the per-scope JSONC
// files. Lookups are O(1) for action-based dispatch and O(1) for
// kind-within-scope lookup.
type Operations struct {
	All      []OperationSpec
	ByKind   map[string]OperationSpec // "<scope>::<kind>" → spec
	ByAction map[string]OperationSpec // KeyAction string → spec
	ByScope  map[OperationScope][]OperationSpec
}

// LookupByKind resolves a spec by its fully-qualified "<scope>::<kind>"
// key. The `::` separator is illegal in either segment so collision is
// impossible.
func (o *Operations) LookupByKind(scope OperationScope, kind string) (OperationSpec, bool) {
	if o == nil {
		return OperationSpec{}, false
	}
	spec, ok := o.ByKind[string(scope)+operationsKindScopeSep+kind]
	return spec, ok
}

// LookupByAction resolves a spec by its KeyAction identifier.
func (o *Operations) LookupByAction(action string) (OperationSpec, bool) {
	if o == nil {
		return OperationSpec{}, false
	}
	spec, ok := o.ByAction[action]
	return spec, ok
}

// ForScope returns every Operation in the given scope, in declaration
// order. The returned slice is the slice registered with the caller;
// callers must not mutate it.
func (o *Operations) ForScope(scope OperationScope) []OperationSpec {
	if o == nil {
		return nil
	}
	return o.ByScope[scope]
}