package keyboard

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/keys"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func doContainerAction(m *state.AppModel, action string, cmdFn func(runtimeapi.Engine, string) tea.Cmd) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelContainers {
		return m, nil
	}
	if len(m.Selection.MarkedIDs) > 0 {
		return doBatchContainerAction(m, action, cmdFn)
	}
	ctr := m.Resources.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	// Starting a container is non-destructive, so execute it immediately.
	if action == string(runtimeapi.ActionStart) {
		m.Feedback.InfoMessage = "starting " + ctr.Name + "..."
		trace := beginAudit(m, "resource.container.start", containerTarget(m, ctr.ID), "Starting "+ctr.Name)
		return m, withContainerAudit(cmdFn(m.Connection.Engine, ctr.ID), trace)
	}
	// Destructive actions (stop/kill/restart) require confirmation
	confirmAction(m, "container-"+action, ctr.ID, fmt.Sprintf("%s container %s?", action, ctr.Name))
	m.Confirm.ConfirmAudit = beginAudit(m, "resource.container."+action, containerTarget(m, ctr.ID),
		fmt.Sprintf("%s container %s", action, ctr.Name))
	return m, nil
}

func doBatchContainerAction(m *state.AppModel, action string, cmdFn func(runtimeapi.Engine, string) tea.Cmd) (*state.AppModel, tea.Cmd) {
	if len(m.Selection.MarkedIDs) == 0 {
		return m, nil
	}
	count := len(m.Selection.MarkedIDs)
	target := fmt.Sprintf("%d items", count)
	message := fmt.Sprintf("Batch %s %d containers?", action, count)
	trace := beginAudit(m, "resource.container."+action,
		audit.ContainerTarget{
			ID:   fmt.Sprintf("batch-%d", count),
			Name: fmt.Sprintf("%d containers", count),
		}, message)
	m.Confirm.Open("batch-"+action, target, message, trace)
	if action == string(runtimeapi.ActionStop) {
		m.Confirm.Options = []state.ChoiceOption{
			{ID: keys.ShowOptionCancel, Label: "Cancel"},
			{ID: keys.ShowOptionForce, Label: "Force", Description: "stop immediately"},
		}
	}
	m.Navigation.Mode = state.ModeConfirm
	return m, nil
}

func executeBatchAction(m *state.AppModel, action string, trace audit.Trace) (*state.AppModel, tea.Cmd) {
	ids := make([]string, 0, len(m.Selection.MarkedIDs))
	for id := range m.Selection.MarkedIDs {
		ids = append(ids, id)
	}
	m.Selection.MarkedIDs = make(map[string]bool)
	if len(ids) == 0 {
		return m, nil
	}

	var cmdFn func(runtimeapi.Engine, string) tea.Cmd
	switch action {
	case string(runtimeapi.ActionStart):
		cmdFn = containerStartCmd
	case string(runtimeapi.ActionStop):
		cmdFn = func(c runtimeapi.Engine, id string) tea.Cmd { return containerStopCmd(c, id) }
	case string(runtimeapi.ActionRestart):
		cmdFn = func(c runtimeapi.Engine, id string) tea.Cmd { return containerRestartCmd(c, id) }
	case string(runtimeapi.ActionKill):
		cmdFn = func(c runtimeapi.Engine, id string) tea.Cmd { return containerKillCmd(c, id) }
	default:
		ShowToastNow(m, fmt.Sprintf("✕ unknown batch action: %s", action))
		return m, nil
	}

	engine := m.Connection.Engine
	return m, func() tea.Msg {
		result := state.BatchActioned{
			Scope:    "container.batch." + action,
			Resource: state.ResourceContainer,
			Total:    len(ids),
			Audit:    trace,
		}
		var failures []error
		for _, id := range ids {
			msg := cmdFn(engine, id)()
			switch v := msg.(type) {
			case state.ContainerActioned:
				if v.Success {
					result.Success++
				} else {
					result.Failed++
					result.FailedIDs = append(result.FailedIDs, id)
					if v.Error != nil {
						failures = append(failures, v.Error)
					}
				}
			default:
				// Unknown message — count as failure so the operator sees it.
				result.Failed++
				result.FailedIDs = append(result.FailedIDs, id)
			}
		}
		result.Error = errors.Join(failures...)
		return result
	}
}

func doLogAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	ctr := m.Resources.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	return m, ToLogView(m, ctr.ID)
}

func doStatsAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	ctr := m.Resources.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	m.Metrics.ToggleContainerStats()
	if m.Metrics.StatsActive {
		return m, FetchStats(m.Connection.Engine.Containers(), ctr.ID)
	}
	return m, nil
}

func doPauseAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	if len(m.Selection.MarkedIDs) > 0 {
		return doBatchPauseAction(m)
	}
	ctr := m.Resources.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	unpause := ctr.State == state.ContainerStatePaused
	if ctr.State != state.ContainerStateRunning && !unpause {
		ShowToastWarn(m, i18n.T("container.pause.unavailable"))
		return m, nil
	}
	action := string(runtimeapi.ActionPause)
	if unpause {
		action = string(runtimeapi.ActionUnpause)
	}
	trace := beginAudit(m, "resource.container."+action, containerTarget(m, ctr.ID), action+" "+ctr.Name)
	return m, withContainerAudit(containerPauseCmd(m.Connection.Engine, ctr.ID, unpause), trace)
}

func doBatchPauseAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	type target struct {
		id      string
		unpause bool
	}
	marked := m.Selection.MarkedIDs
	targets := make([]target, 0, len(marked))
	skipped := 0
	for _, ctr := range m.Resources.Containers.Items {
		if !marked[ctr.ID] {
			continue
		}
		switch ctr.State {
		case state.ContainerStateRunning:
			targets = append(targets, target{id: ctr.ID})
		case state.ContainerStatePaused:
			targets = append(targets, target{id: ctr.ID, unpause: true})
		default:
			skipped++
		}
	}
	m.Selection.ClearMarks()
	trace := beginAudit(m, "resource.container.pause_toggle",
		audit.ContainerTarget{ID: "batch", Name: fmt.Sprintf("%d containers", len(marked))},
		"Toggle pause for selected containers")
	client := m.Connection.Engine
	return m, func() tea.Msg {
		result := state.ContainerBatchActioned{Action: "pause", Skipped: skipped, Audit: trace}
		var failures []error
		for _, item := range targets {
			action := runtimeapi.ActionPause
			if item.unpause {
				action = runtimeapi.ActionUnpause
			}
			_, err := client.Actions().Execute(context.Background(), containerRef(item.id), action, runtimeapi.ActionOptions{})
			if err != nil {
				result.Failed++
				failures = append(failures, err)
			} else {
				result.Success++
			}
		}
		result.Error = errors.Join(failures...)
		return result
	}
}

func openRenameDialog(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	ctr := m.Resources.Containers.Selected()
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelContainers || ctr == nil {
		return m, nil
	}
	m.Dialog.Open(state.DialogSpec{
		Kind:  state.DialogContainerRename,
		Title: i18n.T("container.rename.title"),
		Body:  ctr.ID,
		Input: ctr.Name,
	})
	m.Navigation.Mode = state.ModeRename
	return m, nil
}

func handleRenameDialogKey(key string, m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if key == keys.KeyEsc {
		clearDialogState(m)
		return m, nil
	}
	if key != keys.KeyEnter {
		editQueryInput(key, &m.Dialog.Input)
		return m, nil
	}
	name := strings.TrimSpace(strings.TrimPrefix(m.Dialog.Input.Text, "/"))
	if name == "" || strings.ContainsAny(name, " /\\") {
		ShowToastWarn(m, i18n.T("container.rename.invalid"))
		return m, nil
	}
	id := m.Dialog.Body
	trace := beginAudit(m, "resource.container.rename", containerTarget(m, id), "Rename container to "+name)
	clearDialogState(m)
	return m, withContainerAudit(containerRenameCmd(m.Connection.Engine, id, name), trace)
}

func openTopView(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	ctr := m.Resources.Containers.Selected()
	if m.Connection.Engine == nil || m.Navigation.ActivePanel != state.PanelContainers || ctr == nil {
		return m, nil
	}
	if ctr.State != state.ContainerStateRunning {
		ShowToastWarn(m, i18n.T("container.top.unavailable"))
		return m, nil
	}
	m.Processes.Open(ctr.ID, ctr.Name)
	m.Navigation.Mode = state.ModeTop
	return m, fetchContainerProcesses(m.Connection.Engine.Containers(), ctr.ID)
}

func openPortDetail(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	ctr := m.Resources.Containers.Selected()
	if m.Navigation.ActivePanel != state.PanelContainers || ctr == nil {
		return m, nil
	}
	lines := make([]string, 0, len(ctr.PortBindings))
	for _, binding := range ctr.PortBindings {
		containerPort := fmt.Sprintf("%d/%s", binding.ContainerPort, binding.Protocol)
		if binding.HostPort == 0 {
			lines = append(lines, containerPort)
		} else {
			host := binding.HostIP
			if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
				host = "[" + host + "]"
			}
			lines = append(lines, fmt.Sprintf("%s -> %s:%d", containerPort, host, binding.HostPort))
		}
	}
	if len(lines) == 0 {
		lines = append(lines, i18n.T("container.ports.empty"))
	}
	m.Detail.Open(i18n.T("container.ports.title", ctr.Name), strings.Join(lines, "\n"))
	m.Navigation.Mode = state.ModeDetail
	return m, nil
}

func doExecAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	shell := m.Exec.ExecShell
	if shell == "" {
		return doExecCandidates(m, autoShellCandidates())
	}
	return doExecCandidates(m, []string{shell})
}

func doAutoExecAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	return doExecCandidates(m, autoShellCandidates())
}

func autoShellCandidates() []string {
	return []string{"/bin/sh", "/bin/bash", "/bin/ash", "cmd.exe", "powershell.exe", "pwsh.exe"}
}

func doExecCandidates(m *state.AppModel, candidates []string) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	ctr := m.Resources.Containers.Selected()
	if ctr == nil {
		return m, nil
	}

	m.Exec.ExecShell = ""
	trace := beginAudit(m, "resource.container.exec",
		audit.ExecTarget{ID: ctr.ID, Name: ctr.Name, Meta: audit.ExecMeta{ContainerID: ctr.ID}},
		"Starting exec session in "+ctr.Name)

	ctx := context.Background()
	var session runtimeapi.ExecSession
	var err error
	selectedShell := ""
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		session, err = m.Connection.Engine.Exec().Open(ctx, ctr.ID, runtimeapi.ExecOptions{
			Command:      []string{candidate},
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			TTY:          true,
		})
		if err == nil {
			selectedShell = candidate
			break
		}
	}
	if session == nil {
		if err == nil {
			err = fmt.Errorf("no shell candidates available")
		}
		FinishAudit(m, trace, audit.ResultFailed, "Exec session failed", audit.Details{Error: err.Error(), Shell: strings.Join(candidates, ", ")})
		ShowToastNow(m, fmt.Sprintf("exec: %v", err))
		fallback := "/bin/sh"
		if len(candidates) > 0 && candidates[0] != "" {
			fallback = candidates[0]
		}
		m.Dialog.Open(state.DialogSpec{Kind: state.DialogExec, Input: fallback})
		m.Navigation.Mode = state.ModeExec
		return m, nil
	}
	shell := selectedShell

	if m.Viewport.Width > 0 && m.Viewport.Height > 0 {
		go func() {
			_ = session.Resize(context.Background(), runtimeapi.TerminalSize{Height: uint(m.Viewport.Height), Width: uint(m.Viewport.Width)})
		}() //nolint:errcheck // fire-and-forget resize on session start.
	}

	ch := make(chan string, 100)
	done := make(chan struct{})
	m.Exec.SetShell(shell)
	m.Exec.Start(session.ID(), session, ch, done, trace)
	m.Navigation.Mode = state.ModeExecPassthrough

	// Reader goroutine: reads raw TTY output from exec attach and sends it on ch.
	go func() {
		buf := make([]byte, 4096)
		defer close(done)
		defer func() { _ = session.Close() }() //nolint:errcheck // exec reader exits; close best-effort.
		for {
			n, err := session.Read(buf)
			if n > 0 {
				ch <- string(buf[:n])
			}
			if err != nil {
				close(ch)
				return
			}
		}
	}()

	// Return a cmd that reads the first chunk from the output channel.
	return m, func() tea.Msg {
		data, ok := <-ch
		if !ok {
			return state.ExecDone{}
		}
		return state.ExecOutput{Data: data}
	}
}

func doInspectAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Engine == nil {
		return m, nil
	}
	ctr := m.Resources.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	title := i18n.T("detail.title.container", ctr.Name, ctr.ID)
	ToDetail(m, title, "")
	return m, containerInspectCmd(m.Connection.Engine, ctr.ID, title)
}

func containerInspectCmd(client runtimeapi.Engine, id, title string) tea.Cmd {
	return func() tea.Msg {
		detail, err := client.Containers().Inspect(context.Background(), id)
		return state.ContainerDetailLoaded{
			ContainerID: id,
			Title:       title,
			Detail:      detail,
			Error:       err,
		}
	}
}

func doSwitchRuntime(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Pool == nil {
		ShowToastNow(m, "no runtime pool")
		return m, nil
	}
	names := m.Connection.Pool.KnownHostNames()
	if len(names) < 2 {
		ShowToastNow(m, fmt.Sprintf("only one runtime (%s)", m.Connection.Pool.ActiveName()))
		return m, nil
	}
	current := m.Connection.Pool.ActiveName()
	next := names[0]
	for i, n := range names {
		if n == current && i+1 < len(names) {
			next = names[i+1]
			break
		}
	}
	if current == next {
		next = names[0]
	}
	trace := beginAudit(m, "panel.runtime.switch",
		audit.RuntimeTarget{Name: next, Meta: audit.RuntimeMeta{Previous: current}},
		fmt.Sprintf("Switching runtime from %s to %s", current, next))
	if err := m.Connection.Pool.Connect(next, 10*time.Second); err != nil {
		safeError := i18n.ConnectionFailureMessage(string(runtimeapi.ClassifyConnectionError(err).Kind))
		FinishAudit(m, trace, audit.ResultFailed, "Runtime switch failed", audit.Details{Error: safeError})
		ShowToastNow(m, fmt.Sprintf("switch failed: %s", safeError))
		return m, nil
	}
	m.Connection.ConnectedTo(next, m.Connection.Pool.ActiveEngine())
	FinishAudit(m, trace, audit.ResultSucceeded, "Switched runtime to "+next, audit.Details{})
	cmds := FetchAll(m.Connection.Engine)
	cmds = append(cmds, ShowKeyHint(m, fmt.Sprintf("F2: Switched to %s (%s)", next, m.Connection.RuntimeType)))
	return m, tea.Batch(cmds...)
}

func doContainerSort(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Navigation.ActivePanel != state.PanelContainers {
		return m, nil
	}
	// Cycle: Name → ID → CPU → Mem → State → Created → Name…
	if m.Resources.Containers.SortBy >= state.ContainerSortByCreated {
		m.Resources.Containers.SortBy = state.ContainerSortByName
		m.Resources.Containers.SortAsc = !m.Resources.Containers.SortAsc
	} else {
		m.Resources.Containers.SortBy++
	}
	m.Resources.Containers.Cursor = 0
	m.Resources.Containers.ViewOffset = 0
	labels := map[state.ContainerSortColumn]string{
		state.ContainerSortByName:    "name",
		state.ContainerSortByID:      "id",
		state.ContainerSortByCPU:     "cpu",
		state.ContainerSortByMem:     "mem",
		state.ContainerSortByState:   "state",
		state.ContainerSortByCreated: "created",
	}
	m.Feedback.InfoMessage = fmt.Sprintf("Sort by %s (%v)", labels[m.Resources.Containers.SortBy], m.Resources.Containers.SortAsc)
	return m, nil
}
