package keyboard

import (
	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

// dispatchOperation resolves an action string to a registered Operation
// (R06-03) and routes to the Mode-specific body handler. It returns
// (m, nil, false) when the action is not an Operation — the caller
// falls through to the legacy handleAction switch.
//
// The Mode ↔ body contract is symmetric:
//
//	Mode=form   →  spec.Form.Kind   ("containerCopy", "imageSave", ...)
//	Mode=page   →  spec.Page.Body   ("processes", "historyLayers", ...)
//	Mode=async  →  spec.Async.Body  ("wait", "imagePrune", ...)
//
// Each arm is a closed switch on the body's value, so a typo in
// JSONC surfaces as `handled=false` at dispatch time rather than a
// silent no-op.
func dispatchOperation(action string, m *state.AppModel) (*state.AppModel, tea.Cmd, bool) {
	ops, err := loadOperations()
	if err != nil {
		return m, nil, false
	}
	spec, ok := ops.LookupByAction(action)
	if !ok {
		return m, nil, false
	}
	switch spec.Mode {
	case config.OperationModeForm:
		return dispatchForm(spec, m)
	case config.OperationModePage:
		return dispatchPage(spec, m)
	case config.OperationModeAsync:
		return dispatchAsync(spec, m)
	default:
		return m, nil, false
	}
}

// dispatchForm opens the Form whose state.FormKind matches
// spec.Form.Kind. The FormKind strings here are the JSONC identifiers
// (no Go package prefix); they map 1:1 onto state.FormKind constants.
func dispatchForm(spec config.OperationSpec, m *state.AppModel) (*state.AppModel, tea.Cmd, bool) {
	if spec.Form == nil {
		return m, nil, false
	}
	switch spec.Form.Kind {
	case "containerRename":
		m2, c := openRenameDialog(m)
		return m2, c, true
	case "containerCopy":
		m2, c := openContainerCopyForm(m)
		return m2, c, true
	case "containerUpdate":
		m2, c := openContainerUpdateForm(m)
		return m2, c, true
	case "containerExport":
		m2, c := openContainerExportForm(m)
		return m2, c, true
	case "containerCommit":
		m2, c := openContainerCommitForm(m)
		return m2, c, true
	default:
		return m, nil, false
	}
}

// dispatchPage opens the page whose renderer name matches
// spec.Page.Body. The Body strings here are the JSONC identifiers and
// map onto the page-specific openXxxView functions.
func dispatchPage(spec config.OperationSpec, m *state.AppModel) (*state.AppModel, tea.Cmd, bool) {
	if spec.Page == nil {
		return m, nil, false
	}
	switch spec.Page.Body {
	case "processes":
		m2, c := openTopView(m)
		return m2, c, true
	case "portBindings":
		m2, c := openPortDetail(m)
		return m2, c, true
	case "diffChanges":
		m2, c := doContainerDiff(m)
		return m2, c, true
	case "historyLayers":
		m2, c := openHistoryPage(m)
		return m2, c, true
	default:
		return m, nil, false
	}
}

// dispatchAsync starts the long-running Operation whose async renderer
// matches spec.Async.Body. Today "wait" and "imagePrune" are wired;
// future entries (volumePrune, networkPrune, ...) add their own
// cases here.
func dispatchAsync(spec config.OperationSpec, m *state.AppModel) (*state.AppModel, tea.Cmd, bool) {
	if spec.Async == nil {
		return m, nil, false
	}
	switch spec.Async.Body {
	case "wait":
		m2, c := doContainerWait(m)
		return m2, c, true
	case "imagePrune":
		m2, c := doImagePrune(m)
		return m2, c, true
	default:
		return m, nil, false
	}
}

// loadOperations is a thin wrapper over config.CachedLoadOperations
// so the keyboard dispatcher shares the action bar's cache without
// importing it (and the resulting cycle).
func loadOperations() (*config.Operations, error) {
	return config.CachedLoadOperations()
}