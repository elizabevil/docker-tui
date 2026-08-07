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
// mode carries its own typed body sub-object (Form / Page / Async /
// Confirm) so the JSON schema is self-documenting: a reader sees
// exactly which extra fields a given mode requires without consulting
// Go code.
//
// Mode semantics:
//
//	form    — Structured input form with multiple typed fields (Text,
//	          Int, Path, Bool, Select). The body is `FormFields` and is
//	          rendered by FormDialog. Examples: Copy / Update / Export /
//	          Commit / Rename / Remove.
//	page    — Read-only page view that fills the body panel. The body is
//	          the page renderer name (processes / portBindings / etc.).
//	          Examples: Top / Port / Diff / History.
//	async   — Long-running operation whose progress the user can cancel.
//	          The body is the async renderer name (wait / imagePrune).
//	confirm — Interactive modal the user must dismiss before the action
//	          fires. Two flavours live here today: yes/no (VolumePrune /
//	          NetworkPrune open ModeConfirm) and text-input (VolumeCreate /
//	          NetworkCreate open ModeResourceCreate). Both share the
//	          "user must confirm before backend action runs" semantic —
//	          `confirm` is read as "confirm-before-action", not strictly
//	          "yes/no". The body is the confirm handler's Kind label
//	          (volumePrune / volumeCreate / etc.) which the dispatcher
//	          routes to the right openResourceCreate / confirmResourcePrune
//	          call.
//
// The mode↔body sub-object contract is enforced at load time in
// validateSpec: each mode requires its own sub-object and rejects the
// other three.
type OperationMode string

const (
	OperationModeForm    OperationMode = "form"
	OperationModePage    OperationMode = "page"
	OperationModeAsync   OperationMode = "async"
	OperationModeConfirm OperationMode = "confirm"
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
// switches on this string.
type AsyncBody struct {
	Body string `json:"body"`
}

// ConfirmBody is the per-mode body for mode=confirm. Its `Kind` value
// names the confirm dispatcher handler (e.g. "volumePrune",
// "networkCreate"); the dispatcher routes it to openResourceCreate or
// confirmResourcePrune with the right runtimeapi.ResourceType.
//
// Two flavours share this body type today:
//
//	yes/no  — Kind="volumePrune" / "networkPrune" open ModeConfirm with
//	          a Yes/No dialog (handler: confirmResourcePrune).
//	input   — Kind="volumeCreate" / "networkCreate" open
//	          ModeResourceCreate with a text input (handler:
//	          openResourceCreate).
//
// Both are modal-before-action flows; the Kind label disambiguates
// which handler the dispatcher calls.
type ConfirmBody struct {
	Kind string `json:"kind"`
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
	// RequirementVolume and RequirementNetwork gate the volume and
	// network resource operations. The action bar's
	// evaluateEnabled checks them against the selected resource in
	// the matching panel; the dispatcher does NOT branch on them.
	RequirementVolume   Requirement = "volume"
	RequirementNetwork  Requirement = "network"
	RequirementRunning  Requirement = "running"
	RequirementManifest Requirement = "manifest"
)

// allRequirements is the canonical Requirement universe. The loader
// rejects any Requirement outside this set so the action bar's
// evaluator stays a closed switch.
var allRequirements = map[Requirement]struct{}{
	RequirementEngine:    {},
	RequirementContainer: {},
	RequirementImage:     {},
	RequirementVolume:    {},
	RequirementNetwork:   {},
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

	// Confirm is required when Mode = confirm. Routes to
	// openResourceCreate (Create-style) or confirmResourcePrune
	// (Yes/No-style) handlers via dispatchConfirm.
	Confirm *ConfirmBody `json:"confirm,omitempty"`

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

	// InActionBar controls whether this Operation appears in the action
	// bar (the `;`-triggered overlay). Default true; loader fills the
	// default for JSONC files that omit the field.
	//
	// Set to false when the Operation is already self-discoverable
	// through a direct shortcut and the description adds nothing:
	// simple inputs the user can trigger with one keystroke without
	// needing to read first. Examples today: volumeCreate / networkCreate
	// (both bound to KeyC) — typing `;` to read "Create a new empty
	// volume" before pressing KeyC is friction, not help.
	//
	// Set to true (or omit) when the Operation is the primary trigger,
	// carries no direct shortcut, or is destructive enough that
	// previewing the description is valuable. Examples: all 9 container
	// ops, imageHistory, imagePrune, volumeRemove / networkRemove,
	// volumePrune / networkPrune — the description "Remove a volume" /
	// "Remove all unused volumes" protects against accidental triggers.
	//
	// Constraint C04 documents the full rationale for what should and
	// should not be in the action bar.
	InActionBar bool `json:"inActionBar"`

	// InActionBarWasSet is the unmarshal sentinel for InActionBar:
	// the loader sets it to true when the JSONC key was present and
	// false when the field was omitted. Used to distinguish
	// "explicitly false" from "omitted (defaults to true)". Not
	// serialised; the `json:"-"` tag keeps it out of any future dump.
	InActionBarWasSet bool `json:"-"`
}

// ActionWindow groups the background + border colours of the window chrome
// used by any Operation whose body fills a window (Form / Page / Async).
type ActionWindow struct {
	Background ColorRef `json:"background" yaml:"background"`
	Border     ColorRef `json:"border" yaml:"border"`
}

// ActionFormInput groups the foreground + background colours of the
// editable form input fields inside the window. Only consumed when the
// Operation Mode is Form.
type ActionFormInput struct {
	Foreground ColorRef `json:"foreground" yaml:"foreground"`
	Background ColorRef `json:"background" yaml:"background"`
}

// ActionButton groups the foreground + background colours of an action
// button (Confirm or Cancel) inside the window. Only consumed when the
// Operation Mode is Form or Async.
type ActionButton struct {
	Foreground ColorRef `json:"foreground" yaml:"foreground"`
	Background ColorRef `json:"background" yaml:"background"`
}

// ActionScopeStyles is the per-scope visual contract: one entry per
// resource scope (Container / Image / ...). The four slots are reused
// across all scopes so each Operation Mode (Form / Page / Async) can
// pull from the same shape regardless of scope.
type ActionScopeStyles struct {
	Window    ActionWindow    `json:"window" yaml:"window"`
	FormInput ActionFormInput `json:"formInput" yaml:"formInput"`
	Confirm   ActionButton    `json:"confirm" yaml:"confirm"`
	Cancel    ActionButton    `json:"cancel" yaml:"cancel"`
}

// ActionStyles is the visual contract for any Operation, grouped by
// resource Scope. It is embedded in the Theme struct under
// `theme.action.*` so themes can paint each scope's window chrome.
//
// Adding a new scope (e.g. Volume / Network in R06-09) requires:
//  1. Add `OperationScope<Scope>` constant in operations.go
//  2. Add the matching field here
//  3. Add the JSONC section in every theme
//  4. Add the matching StyleAction<Scope>* constants + applyActionScopeTo
//     wiring in styles_load.go
type ActionStyles struct {
	Container ActionScopeStyles `json:"container" yaml:"container"`
	Image     ActionScopeStyles `json:"image" yaml:"image"`
	// Volume + Network are the R06-09 expansions. They follow the
	// same 4-slot contract as Container / Image (window / formInput /
	// confirm / cancel) and default to the same neutral colours; themes
	// can differentiate by overriding action.volume or action.network.
	Volume  ActionScopeStyles `json:"volume" yaml:"volume"`
	Network ActionScopeStyles `json:"network" yaml:"network"`
}

// defaultActionScopeStyles returns the fallback contract for a single
// Action scope. Identical for container / image today; a future theme
// divergence is a single constant edit away.
func defaultActionScopeStyles() ActionScopeStyles {
	return ActionScopeStyles{
		Window: ActionWindow{
			Background: TokenRef(ColorTokenBackgroundSubtle),
			Border:     TokenRef(ColorTokenPrimary),
		},
		FormInput: ActionFormInput{
			Foreground: TokenRef(ColorTokenForeground),
			Background: TokenRef(ColorTokenBackgroundDeep),
		},
		Confirm: ActionButton{
			Foreground: TokenRef(ColorTokenSuccess),
			Background: TokenRef(ColorTokenTransparent),
		},
		Cancel: ActionButton{
			Foreground: TokenRef(ColorTokenForegroundMuted),
			Background: TokenRef(ColorTokenTransparent),
		},
	}
}

// ActionWindowPatch / ActionFormInputPatch / ActionButtonPatch are the
// user-override shapes for the matching fields.
type ActionWindowPatch struct {
	Background *ColorRef `json:"background,omitempty" yaml:"background,omitempty"`
	Border     *ColorRef `json:"border,omitempty" yaml:"border,omitempty"`
}

type ActionFormInputPatch struct {
	Foreground *ColorRef `json:"foreground,omitempty" yaml:"foreground,omitempty"`
	Background *ColorRef `json:"background,omitempty" yaml:"background,omitempty"`
}

type ActionButtonPatch struct {
	Foreground *ColorRef `json:"foreground,omitempty" yaml:"foreground,omitempty"`
	Background *ColorRef `json:"background,omitempty" yaml:"background,omitempty"`
}

type ActionScopeStylesPatch struct {
	Window    *ActionWindowPatch    `json:"window,omitempty" yaml:"window,omitempty"`
	FormInput *ActionFormInputPatch `json:"formInput,omitempty" yaml:"formInput,omitempty"`
	Confirm   *ActionButtonPatch    `json:"confirm,omitempty" yaml:"confirm,omitempty"`
	Cancel    *ActionButtonPatch    `json:"cancel,omitempty" yaml:"cancel,omitempty"`
}

type ActionStylesPatch struct {
	Container *ActionScopeStylesPatch `json:"container,omitempty" yaml:"container,omitempty"`
	Image     *ActionScopeStylesPatch `json:"image,omitempty" yaml:"image,omitempty"`
	Volume    *ActionScopeStylesPatch `json:"volume,omitempty" yaml:"volume,omitempty"`
	Network   *ActionScopeStylesPatch `json:"network,omitempty" yaml:"network,omitempty"`
}

// applyActionScopePatch overwrites the matching fields on target with
// non-nil entries from p.
func applyActionScopePatch(target *ActionScopeStyles, p *ActionScopeStylesPatch) {
	if p.Window != nil {
		assign(&target.Window.Background, p.Window.Background)
		assign(&target.Window.Border, p.Window.Border)
	}
	if p.FormInput != nil {
		assign(&target.FormInput.Foreground, p.FormInput.Foreground)
		assign(&target.FormInput.Background, p.FormInput.Background)
	}
	if p.Confirm != nil {
		assign(&target.Confirm.Foreground, p.Confirm.Foreground)
		assign(&target.Confirm.Background, p.Confirm.Background)
	}
	if p.Cancel != nil {
		assign(&target.Cancel.Foreground, p.Cancel.Foreground)
		assign(&target.Cancel.Background, p.Cancel.Background)
	}
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
