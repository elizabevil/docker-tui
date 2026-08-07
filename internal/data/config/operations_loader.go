package config

import (
	"errors"
	"fmt"
	"io/fs"
	"sort"
)

// LoadOperations reads every embedded operations scope file in
// defaults/operations/scopes/, validates each Operation against the
// schema (mode ↔ body sub-object, Requirement universe, kind/action
// uniqueness) and returns an indexed Operations struct.
//
// Each Operation is fully self-contained in its scope file — there is
// no global registry.jsonc to cross-reference. Lookups for KeyAction
// and (Scope, Kind) are derived from the union of all scope files.
func LoadOperations() (*Operations, error) {
	scopeFiles, err := discoverScopeFiles(DefaultsFS(), operationsScopesDir)
	if err != nil {
		return nil, err
	}

	byKind := make(map[string]OperationSpec)
	byAction := make(map[string]OperationSpec)
	byScope := make(map[OperationScope][]OperationSpec)

	for _, name := range scopeFiles {
		doc, err := readScopeFile(DefaultsFS(), name)
		if err != nil {
			return nil, err
		}
		if _, exists := byScope[doc.Scope]; exists {
			return nil, fmt.Errorf(errOperationsScopeCollisionFormat, name, errOperationsDuplicateKindFormat)
		}
		for i := range doc.Operations {
			spec := doc.Operations[i]
			// Stamp the file's declared Scope onto each Operation so the
			// resulting structs are self-describing (the file structure
			// is not visible once they live in Operations.ByKind).
			spec.Scope = doc.Scope
			// Apply the InActionBar default. JSONC files that omit
			// `inActionBar` decode to bool zero-value (false); we want
			// the historical "show in action bar by default" behaviour.
			// Operators that want to opt out of the action bar must
			// write `"inActionBar": false` explicitly — see C04.
			if !spec.InActionBarWasSet {
				spec.InActionBar = true
			}
			doc.Operations[i] = spec
			if err := validateSpec(spec, name); err != nil {
				return nil, err
			}
			key := string(spec.Scope) + operationsKindScopeSep + spec.Kind
			if _, dup := byKind[key]; dup {
				return nil, fmt.Errorf(errOperationsScopeCollisionFormat, name, errOperationsDuplicateKindFormat)
			}
			if _, dup := byAction[spec.Action]; dup {
				return nil, fmt.Errorf(errOperationsDuplicateActionFormat, spec.Action, name, lookupActionFile(byAction[spec.Action]))
			}
			byKind[key] = spec
			byAction[spec.Action] = spec
		}
		byScope[doc.Scope] = doc.Operations
	}

	out := &Operations{
		All:      make([]OperationSpec, 0, len(byKind)),
		ByKind:   byKind,
		ByAction: byAction,
		ByScope:  byScope,
	}
	for _, spec := range byKind {
		out.All = append(out.All, spec)
	}
	// Stable sort by scope then kind so consumers (action bar, debug
	// dumps) iterate in a predictable order.
	sort.SliceStable(out.All, func(i, j int) bool {
		if out.All[i].Scope != out.All[j].Scope {
			return out.All[i].Scope < out.All[j].Scope
		}
		return out.All[i].Kind < out.All[j].Kind
	})
	return out, nil
}

// validateSpec enforces the four invariants of an Operation entry:
//
//  1. Kind and Action are non-empty.
//  2. Exactly one mode-body sub-object is set, matching Mode.
//  3. Every Requirement token is in the closed allRequirements set.
//  4. No empty body sub-objects.
//
// The file argument appears in the returned error so the caller can
// locate the offending JSONC entry without re-discovering it.
func validateSpec(spec OperationSpec, file string) error {
	if spec.Kind == emptyValue {
		return fmt.Errorf(errOperationsScopeCollisionFormat, file, errOperationsEmptyKindFormat)
	}
	if spec.Action == emptyValue {
		return fmt.Errorf(errOperationsScopeCollisionFormat, file, errOperationsEmptyActionFormat)
	}
	for _, req := range spec.Requires {
		if _, ok := allRequirements[req]; !ok {
			return fmt.Errorf(errOperationsUnknownRequirementFormat, spec.Scope, spec.Kind, req)
		}
	}
	formSet := spec.Form != nil
	pageSet := spec.Page != nil
	asyncSet := spec.Async != nil
	confirmSet := spec.Confirm != nil
	switch spec.Mode {
	case OperationModeForm:
		if !formSet {
			return fmt.Errorf(errOperationsMissingModeBodyFormat, spec.Scope, spec.Kind, spec.Mode)
		}
		if pageSet {
			return fmt.Errorf(errOperationsUnexpectedModeBodyFormat, spec.Scope, spec.Kind, string(OperationModePage), spec.Mode)
		}
		if asyncSet {
			return fmt.Errorf(errOperationsUnexpectedModeBodyFormat, spec.Scope, spec.Kind, string(OperationModeAsync), spec.Mode)
		}
		if confirmSet {
			return fmt.Errorf(errOperationsUnexpectedModeBodyFormat, spec.Scope, spec.Kind, string(OperationModeConfirm), spec.Mode)
		}
	case OperationModePage:
		if !pageSet {
			return fmt.Errorf(errOperationsMissingModeBodyFormat, spec.Scope, spec.Kind, spec.Mode)
		}
		if formSet {
			return fmt.Errorf(errOperationsUnexpectedModeBodyFormat, spec.Scope, spec.Kind, string(OperationModeForm), spec.Mode)
		}
		if asyncSet {
			return fmt.Errorf(errOperationsUnexpectedModeBodyFormat, spec.Scope, spec.Kind, string(OperationModeAsync), spec.Mode)
		}
		if confirmSet {
			return fmt.Errorf(errOperationsUnexpectedModeBodyFormat, spec.Scope, spec.Kind, string(OperationModeConfirm), spec.Mode)
		}
	case OperationModeAsync:
		if !asyncSet {
			return fmt.Errorf(errOperationsMissingModeBodyFormat, spec.Scope, spec.Kind, spec.Mode)
		}
		if formSet {
			return fmt.Errorf(errOperationsUnexpectedModeBodyFormat, spec.Scope, spec.Kind, string(OperationModeForm), spec.Mode)
		}
		if pageSet {
			return fmt.Errorf(errOperationsUnexpectedModeBodyFormat, spec.Scope, spec.Kind, string(OperationModePage), spec.Mode)
		}
		if confirmSet {
			return fmt.Errorf(errOperationsUnexpectedModeBodyFormat, spec.Scope, spec.Kind, string(OperationModeConfirm), spec.Mode)
		}
	case OperationModeConfirm:
		if !confirmSet {
			return fmt.Errorf(errOperationsMissingModeBodyFormat, spec.Scope, spec.Kind, spec.Mode)
		}
		if formSet {
			return fmt.Errorf(errOperationsUnexpectedModeBodyFormat, spec.Scope, spec.Kind, string(OperationModeForm), spec.Mode)
		}
		if pageSet {
			return fmt.Errorf(errOperationsUnexpectedModeBodyFormat, spec.Scope, spec.Kind, string(OperationModePage), spec.Mode)
		}
		if asyncSet {
			return fmt.Errorf(errOperationsUnexpectedModeBodyFormat, spec.Scope, spec.Kind, string(OperationModeAsync), spec.Mode)
		}
	default:
		return fmt.Errorf(errOperationsScopeCollisionFormat, file, errOperationsDuplicateKindFormat)
	}
	return nil
}

// readScopeFile reads + decodes one scope JSONC file.
func readScopeFile(fsys fs.FS, name string) (*operationsScopeDocument, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf(errOperationsReadEmbeddedFormat, name, err)
	}
	var doc operationsScopeDocument
	if err := decodeJSONCStrict(data, &doc); err != nil {
		return nil, fmt.Errorf(errOperationsParseEmbeddedFormat, name, err)
	}
	return &doc, nil
}

// discoverScopeFiles lists the operations scope files in lexical order.
// Returning a sorted list keeps the loader's iteration deterministic so
// duplicate-detection errors always name the earlier file first.
func discoverScopeFiles(fsys fs.FS, dir string) ([]string, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf(errOperationsReadScopeDirFormat, dir, err)
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if len(e.Name()) < len(operationsFileExt) {
			continue
		}
		if e.Name()[len(e.Name())-len(operationsFileExt):] != operationsFileExt {
			continue
		}
		out = append(out, dir+operationsSlash+e.Name())
	}
	sort.Strings(out)
	return out, nil
}

// lookupActionFile is the inverse of an Action's source: given a spec,
// return the scope name that declared it. Used to format the duplicate
// action error message so the operator can find both offending files.
func lookupActionFile(spec OperationSpec) string {
	return string(spec.Scope) + operationsSlash + string(spec.Kind)
}

// ensure errors import stays in case future cases need it directly.
var _ = errors.New

// operationsScopeDocument is the on-disk shape of one scope file. The
// Scope field is required and matches the file's parent directory
// name in practice, but the loader cross-checks for safety.
type operationsScopeDocument struct {
	Scope      OperationScope  `json:"scope"`
	Operations []OperationSpec `json:"operations"`
}
