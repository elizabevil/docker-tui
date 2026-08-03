package keyboard

import (
	"fmt"
	"time"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func HandleKeyPress(msg tea.KeyPressMsg, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	rawKey := msg.String()
	key := keys.Normalize(rawKey)

	// Esc clears persistent error feedback; any non-Esc key clears the
	// transient toast/info line. Both side effects are global preamble
	// before mode-specific dispatch.
	if key != keys.KeyEsc {
		m.Navigation.EscPending = false
		m.Feedback.InfoMessage = ""
	}
	if key == keys.KeyEsc && m.Feedback.ErrorMessage != "" {
		m.Feedback.ClearError()
	}

	// Events is a global read-only page. Resolve it before page-owned modes
	// such as History and Help consume otherwise-unhandled keys. Input and
	// confirmation modes remain modal and are intentionally excluded.
	if m.Navigation.Mode != state.ModeEvents && canOpenEventsFrom(m.Navigation.Mode) {
		if action, known := resolveAction(key, m); known && action == keys.ActionEvents {
			return handleAction(action, m, []tea.Cmd{RecordKeyStroke(m, key, KeyStrokeActionLabel(key))})
		}
	}

	// Mode-keyed dispatch first: dialogs, inputs, exec passthrough and
	// focused detail overlays all consume the key before the global
	// action table gets a chance.
	if mm, cmd, handled := dispatchByMode(rawKey, key, m); handled {
		return mm, cmd
	}

	// Nested views (image -> containers, compose -> services) own
	// additional cursor semantics and are resolved before the main
	// action table.
	if mm, cmd := handleNestedViewKeys(key, m); mm != nil || cmd != nil {
		return mm, cmd
	}

	// Q is intentionally unassigned in normal mode, including for users
	// whose older config still binds it to Quit. Input modes receive it as
	// text because their dispatch runs before this compatibility guard.
	if key == keys.KeyQ && m.Navigation.Mode == state.ModeNormal {
		return m, nil
	}

	// Global action table. Fall through to panel-level fallbacks and
	// finally to the small set of fixed-key shortcuts.
	var cmds []tea.Cmd
	if action, known := resolveAction(key, m); known {
		cmds = append(cmds, RecordKeyStroke(m, key, KeyStrokeActionLabel(key)))
		return handleAction(action, m, cmds)
	}
	if mm, cmd := handlePanelFallbacks(key, m); mm != nil || cmd != nil {
		return mm, cmd
	}
	return handleShortcuts(key, m, cmds)
}

func canOpenEventsFrom(mode state.AppMode) bool {
	switch mode {
	case state.ModeNormal, state.ModeDetail, state.ModeLogView, state.ModeHelp,
		state.ModeTop, state.ModeAuditDetail, state.ModeHistory:
		return true
	default:
		return false
	}
}

// dispatchByMode routes the key to the handler that owns the active
// AppMode. Returns handled=false when no mode matched so the caller can
// continue to the global action table.
func dispatchByMode(rawKey, key string, m *state.AppModel) (*state.AppModel, tea.Cmd, bool) {
	switch m.Navigation.Mode {
	case state.ModeHelp:
		if action, known := resolveAction(key, m); known && (action == keys.ActionHelp || action == keys.ActionBack) {
			BackFromHelp(m)
		}
		return m, nil, true
	case state.ModeExecPassthrough:
		return handleExecPassthrough(key, m), nil, true
	case state.ModeFilter:
		m, cmd := handleFilterInput(normalizeInputKey(rawKey), m)
		return m, cmd, true
	case state.ModeSearch:
		return handleSearchInput(normalizeInputKey(rawKey), m), nil, true
	case state.ModeImagePull:
		return handleImagePullInput(normalizeInputKey(rawKey), m), nil, true
	case state.ModeImageWorkflow:
		m, cmd := handleImageWorkflowInput(normalizeInputKey(rawKey), m)
		return m, cmd, true
	case state.ModeImageTransfer:
		if key == keys.KeyEsc {
			m, cmd := cancelImageTransfer(m)
			return m, cmd, true
		}
		return m, nil, true
	case state.ModeCommand:
		m, cmd := handleCommandInput(normalizeInputKey(rawKey), m)
		return m, cmd, true
	case state.ModeRuntimeSelect:
		m, cmd := handleRuntimeSelectorKey(key, m)
		return m, cmd, true
	case state.ModeRename:
		m, cmd := handleRenameDialogKey(normalizeInputKey(rawKey), m)
		return m, cmd, true
	case state.ModeResourceCreate:
		m, cmd := handleResourceCreateKey(normalizeInputKey(rawKey), m)
		return m, cmd, true
	case state.ModeTop:
		m, cmd := handleTopKey(key, m)
		return m, cmd, true
	case state.ModeAuditDetail:
		m, cmd := handleAuditDetailKey(key, m)
		return m, cmd, true
	case state.ModeExec:
		m, cmd := handleExecDialogKeys(key, m)
		return m, cmd, true
	case state.ModeConfirm:
		m, cmd := handleConfirmKeys(key, m)
		return m, cmd, true
	case state.ModeActionBar:
		m, cmd := handleActionBarKeys(rawKey, m)
		return m, cmd, true
	case state.ModeHistory:
		m, cmd := handleHistoryKeys(rawKey, m)
		return m, cmd, true
	case state.ModeEvents:
		m, cmd := handleEventPanelKeys(rawKey, m)
		return m, cmd, true
	case state.ModeContainerForm:
		m, cmd := handleContainerFormKey(normalizeInputKey(rawKey), m)
		return m, cmd, true
	case state.ModeMark:
		// Mark mode handles both arrow-keyed navigation and Space toggle.
		m, cmd := handleMarkMode(key, m)
		return m, cmd, true
	}

	// Detail and log overlay modes share a key surface but route
	// through helper-specific state. They are not exhaustive — keys
	// they do not handle fall through to the action table.
	if m.Navigation.Mode == state.ModeDetail {
		if handled, cmd := handleDetailKeys(key, m); handled {
			return m, cmd, true
		}
	}
	if m.Navigation.Mode == state.ModeLogView {
		if handleLogKeys(key, m) {
			return m, nil, true
		}
	}

	// Confirmation / prompt dialogs that share the main view.
	if m.Dialog.Kind.IsSelection() {
		return handleSelectionDialog(key, m), nil, true
	}
	return nil, nil, false
}

// handleSelectionDialog consumes keys when a yes/no prompt is active.
func handleSelectionDialog(key string, m *state.AppModel) *state.AppModel {
	switch key {
	case keys.KeyTab:
		m.Dialog.MoveFocus(1, 2)
	case keys.KeyEnter:
		if m.Dialog.Focus == 0 {
			ShowToastNow(m, "✓ "+m.Dialog.Action)
		}
		clearDialogState(m)
	case keys.KeyEsc, keys.KeyN:
		clearDialogState(m)
	}
	return m
}

// handleExecPassthrough forwards keystrokes into the active exec session
// and tears the session down on Esc.
func handleExecPassthrough(key string, m *state.AppModel) *state.AppModel {
	if key == keys.KeyEsc {
		FinishAudit(m, m.Exec.ExecAudit, audit.ResultCancelled, "Exec session closed by user", audit.Details{Shell: m.Exec.ExecShell})
		if m.Exec.ExecConn != nil {
			_ = m.Exec.ExecConn.Close() //nolint:errcheck // teardown of exec session; close best-effort.
		}
		m.Exec.Reset()
		m.Navigation.Mode = state.ModeNormal
		return m
	}
	if m.Exec.ExecConn != nil {
		_, _ = m.Exec.ExecConn.Write(mapKeyToTerm(key)) //nolint:errcheck // passthrough keystroke; failure is non-fatal.
	}
	return m
}

// handleNestedViewKeys routes the key to a panel-internal cursor handler
// when a nested view (image -> containers, compose -> services) is open.
func handleNestedViewKeys(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Navigation.ActivePanel == state.PanelImages && m.Resources.Images.ContainersViewID != "" {
		if mm, cmd := handleImagePanelKeys(key, m); mm != nil || cmd != nil {
			return mm, cmd
		}
	}
	if m.Navigation.ActivePanel == state.PanelCompose && m.Compose.ComposeContainerViewID != "" {
		if mm, cmd := handleComposePanelKeys(key, m); mm != nil || cmd != nil {
			return mm, cmd
		}
	}
	return nil, nil
}

// handlePanelFallbacks covers keys the global action table does not own
// but a particular panel does (image / compose / audit rows).
func handlePanelFallbacks(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if mm, cmd := handleImagePanelKeys(key, m); mm != nil || cmd != nil {
		return mm, cmd
	}
	if mm, cmd := handleComposePanelKeys(key, m); mm != nil || cmd != nil {
		return mm, cmd
	}
	if m.Navigation.ActivePanel == state.PanelAudit {
		if mm, cmd := handleAuditPanelKey(key, m); mm != nil || cmd != nil {
			return mm, cmd
		}
	}
	return nil, nil
}

// handleShortcuts covers the small set of fixed-key bindings that do
// not appear in the action table: connection info, sort controls, and
// Space-mark toggle.
func handleShortcuts(key string, m *state.AppModel, cmds []tea.Cmd) (*state.AppModel, tea.Cmd) {
	if keys.IsSpace(key) {
		if m.Navigation.Mode == state.ModeMark {
			mm, cmd := doToggleMark(m)
			return mm, batchWith(cmds, RecordKeyStroke(m, key, keys.ActionLabelToggle), cmd)
		}
		mm, cmd := enterMarkMode(m)
		return mm, batchWith(cmds, RecordKeyStroke(m, key, keys.ActionLabelMark), cmd)
	}
	switch key {
	case keys.KeyO:
		if m.Navigation.Mode == state.ModeMark {
			return m, tea.Batch(cmds...)
		}
		if mm, cmd, ok := toggleSortColumn(key, m, cmds); ok {
			return mm, cmd
		}
	case keys.KeyCtrlO:
		if m.Navigation.Mode == state.ModeMark {
			return m, tea.Batch(cmds...)
		}
		if mm, cmd, ok := toggleSortDirection(key, m, cmds); ok {
			return mm, cmd
		}
	}
	return m, nil
}

// batchWith batches the accumulated cmds with any additional commands.
// When no extras are provided, it returns the existing batch as-is.
func batchWith(cmds []tea.Cmd, extras ...tea.Cmd) tea.Cmd {
	all := make([]tea.Cmd, 0, len(cmds)+len(extras))
	all = append(all, cmds...)
	all = append(all, extras...)
	return tea.Batch(all...)
}

// toggleSortColumn cycles the sort column for the panel under focus.
// Returns ok=false when the active panel has no sortable list.
func toggleSortColumn(key string, m *state.AppModel, cmds []tea.Cmd) (*state.AppModel, tea.Cmd, bool) {
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		mm, cmd := doContainerSort(m)
		return mm, batchWith(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort), cmd), true
	case state.PanelImages:
		mm, cmd := doImageSort(m)
		return mm, batchWith(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort), cmd), true
	case state.PanelNetworks:
		mm, cmd := doNetworkSort(m)
		return mm, batchWith(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort), cmd), true
	}
	return nil, nil, false
}

// toggleSortDirection flips the ascending/descending flag for the
// panel under focus without changing the sort column. Returns ok=false
// when the active panel has no sortable list.
func toggleSortDirection(key string, m *state.AppModel, cmds []tea.Cmd) (*state.AppModel, tea.Cmd, bool) {
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		m.Resources.Containers.SortAsc = !m.Resources.Containers.SortAsc
		m.Resources.Containers.Cursor = 0
		m.Resources.Containers.ViewOffset = 0
		cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort))
		return m, tea.Batch(cmds...), true
	case state.PanelImages:
		m.Resources.Images.SortAsc = !m.Resources.Images.SortAsc
		m.Resources.Images.Cursor = 0
		m.Resources.Images.ViewOffset = 0
		cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort))
		return m, tea.Batch(cmds...), true
	case state.PanelNetworks:
		m.Resources.Networks.SortAsc = !m.Resources.Networks.SortAsc
		m.Resources.Networks.Cursor = 0
		m.Resources.Networks.ViewOffset = 0
		cmds = append(cmds, RecordKeyStroke(m, key, keys.ActionLabelSort))
		return m, tea.Batch(cmds...), true
	}
	return nil, nil, false
}

func normalizeInputKey(key string) string {
	if len([]rune(key)) == 1 {
		return key
	}
	normalized := keys.Normalize(key)
	if keys.IsSpace(normalized) {
		return keys.KeySpace
	}
	return normalized
}

func resolveAction(key string, m *state.AppModel) (keys.KeyAction, bool) {
	if m == nil || m.Dependencies.Config == nil {
		return "", false
	}
	resolver := keys.NewResolver(keys.CompileBindings(m.Dependencies.Config.Keymap))
	return resolver.Resolve(key, keyContext(m))
}

func keyContext(m *state.AppModel) keys.Context {
	view := "containers"
	switch m.Navigation.ActivePanel {
	case state.PanelImages:
		if m.Resources.Images.ContainersViewID != "" {
			view = "image-containers"
		} else {
			view = "images"
		}
	case state.PanelVolumes:
		view = "volumes"
	case state.PanelNetworks:
		view = "networks"
	case state.PanelCompose:
		if m.Compose.ComposeContainerViewID != "" {
			view = "compose-containers"
		} else {
			view = "compose"
		}
	case state.PanelHelp:
		view = "help"
	case state.PanelAudit:
		view = "audit"
	}
	return keys.Context{App: "app", Surface: keySurface(m.Navigation.Mode), View: view, Mode: keyMode(m.Navigation.Mode)}
}

func keySurface(mode state.AppMode) string {
	switch mode {
	case state.ModeFilter, state.ModeSearch, state.ModeImagePull, state.ModeImageWorkflow, state.ModeCommand:
		return "input"
	case state.ModeConfirm, state.ModeExport, state.ModeDebug, state.ModeExec, state.ModeExecShell, state.ModeRename, state.ModeResourceCreate, state.ModeImageTransfer, state.ModeAuditDetail, state.ModeContainerForm:
		return "dialog"
	default:
		return "main"
	}
}

func keyMode(mode state.AppMode) string {
	switch mode {
	case state.ModeFilter:
		return "filter"
	case state.ModeSearch:
		return "search"
	case state.ModeImagePull:
		return "image-pull"
	case state.ModeCommand:
		return "command"
	case state.ModeLogView:
		return "logs"
	case state.ModeDetail:
		return "detail"
	case state.ModeMark:
		return "mark"
	case state.ModeTop:
		return "top"
	case state.ModeRename:
		return "rename"
	case state.ModeResourceCreate:
		return "resource-create"
	case state.ModeAuditDetail:
		return "audit-detail"
	case state.ModeEvents:
		return "events"
	case state.ModeContainerForm:
		return "container-form"
	default:
		return "normal"
	}
}

func handleFilterInput(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	m.Navigation.FilterInput.Clamp()

	switch key {
	case keys.KeyEnter:
		ApplyFilter(m)
		focusFirstFilteredItem(m)
		m.Navigation.Mode = state.ModeNormal
		m.Navigation.ClearFilterExit()
	case keys.KeyEsc:
		if m.Navigation.FilterExitPending {
			BackFromFilter(m)
			return m, nil
		}
		token := m.Navigation.BeginFilterExit()
		ShowToastWarn(m, "Press Esc again within 5s to clear filter and exit")
		return m, tea.Tick(5*time.Second, func(time.Time) tea.Msg {
			return state.FilterExitTimeout{Token: token}
		})
	default:
		if handled, changed := editQueryInput(key, &m.Navigation.FilterInput); handled && changed {
			ApplyFilter(m)
		}
	}
	return m, nil
}

func handleSearchInput(key string, m *state.AppModel) *state.AppModel {
	m.Navigation.SearchInput.Clamp()
	switch key {
	case keys.KeyEnter:
		matches := m.Log.ApplySearch(m.Navigation.SearchInput.Text)
		m.Navigation.Mode = state.ModeLogView
		if matches == 0 && m.Log.LogSearchText != "" {
			ShowToastWarn(m, "No log matches for: "+m.Log.LogSearchText)
		} else if matches > 0 {
			ShowToastNow(m, fmt.Sprintf("Log match 1/%d", matches))
		}
	case keys.KeyEsc:
		m.Navigation.Mode = state.ModeLogView
		m.Navigation.SearchInput.Set(m.Log.LogSearchText)
	default:
		editQueryInput(key, &m.Navigation.SearchInput)
	}
	return m
}

func focusFirstFilteredItem(m *state.AppModel) {
	if m == nil {
		return
	}
	switch m.Navigation.ActivePanel {
	case state.PanelContainers:
		if m.Resources.Containers != nil {
			m.Resources.Containers.Cursor = 0
			m.Resources.Containers.ViewOffset = 0
		}
	case state.PanelImages:
		if m.Resources.Images != nil {
			m.Resources.Images.Cursor = 0
			m.Resources.Images.ViewOffset = 0
		}
	case state.PanelVolumes:
		if m.Resources.Volumes != nil {
			m.Resources.Volumes.Cursor = 0
			m.Resources.Volumes.ViewOffset = 0
		}
	case state.PanelNetworks:
		if m.Resources.Networks != nil {
			m.Resources.Networks.Cursor = 0
			m.Resources.Networks.ViewOffset = 0
		}
	case state.PanelCompose:
		if m.Compose.ComposeFocus == 1 {
			m.Compose.ComposeServiceCursor = 0
		} else {
			m.Compose.ComposeCursor = 0
			m.Compose.ComposeServiceCursor = 0
		}
	}
}
