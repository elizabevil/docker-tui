package keyboard

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/data/i18n"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func doContainerAction(m *state.AppModel, action string, cmdFn func(*docker.Client, string) tea.Cmd) (*state.AppModel, tea.Cmd) {
	if m.Connection.Docker == nil || m.Navigation.ActivePanel != state.PanelContainers {
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
	if action == docker.ContainerActionStart {
		m.Feedback.InfoMessage = "starting " + ctr.Name + "..."
		trace := beginAudit(m, "resource.container.start", containerTarget(m, ctr.ID), "Starting "+ctr.Name)
		return m, withContainerAudit(cmdFn(m.Connection.Docker, ctr.ID), trace)
	}
	// Destructive actions (stop/kill/restart) require confirmation
	confirmAction(m, "container-"+action, ctr.ID, fmt.Sprintf("%s container %s?", action, ctr.Name))
	m.Confirm.ConfirmAudit = beginAudit(m, "resource.container."+action, containerTarget(m, ctr.ID), fmt.Sprintf("%s container %s", action, ctr.Name))
	return m, nil
}

func doBatchContainerAction(m *state.AppModel, action string, cmdFn func(*docker.Client, string) tea.Cmd) (*state.AppModel, tea.Cmd) {
	if len(m.Selection.MarkedIDs) == 0 {
		return m, nil
	}
	m.Navigation.Mode = state.ModeConfirm
	m.Confirm.ConfirmAction = "batch-" + action
	m.Confirm.ConfirmTarget = fmt.Sprintf("%d items", len(m.Selection.MarkedIDs))
	m.Confirm.ConfirmMessage = fmt.Sprintf("Batch %s %d containers?", action, len(m.Selection.MarkedIDs))
	m.Confirm.ConfirmAudit = beginAudit(m, "resource.container."+action, audit.ContainerTarget{ID: fmt.Sprintf("batch-%d", len(m.Selection.MarkedIDs)), Name: fmt.Sprintf("%d containers", len(m.Selection.MarkedIDs))}, m.Confirm.ConfirmMessage)
	return m, nil
}

func executeBatchAction(m *state.AppModel, action string, trace audit.Trace) (*state.AppModel, tea.Cmd) {
	ids := make([]string, 0, len(m.Selection.MarkedIDs))
	for id := range m.Selection.MarkedIDs {
		ids = append(ids, id)
	}
	m.Selection.MarkedIDs = make(map[string]bool)

	var cmdFn func(*docker.Client, string) tea.Cmd
	switch action {
	case docker.ContainerActionStart:
		cmdFn = containerStartCmd
	case docker.ContainerActionStop:
		cmdFn = func(c *docker.Client, id string) tea.Cmd { return containerStopCmd(c, id) }
	case docker.ContainerActionRestart:
		cmdFn = func(c *docker.Client, id string) tea.Cmd { return containerRestartCmd(c, id) }
	case docker.ContainerActionKill:
		cmdFn = func(c *docker.Client, id string) tea.Cmd { return containerKillCmd(c, id) }
	default:
		ShowToastNow(m, fmt.Sprintf("✕ unknown batch action: %s", action))
		return m, nil
	}

	var cmds []tea.Cmd
	for _, id := range ids {
		cmds = append(cmds, withContainerAudit(cmdFn(m.Connection.Docker, id), trace))
	}
	ShowToastNow(m, fmt.Sprintf("✓ batch %s %d containers", action, len(ids)))
	return m, tea.Batch(cmds...)
}

func doContainerRemove(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Docker == nil || m.Navigation.ActivePanel != state.PanelContainers {
		return m, nil
	}
	ctr := m.Resources.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	confirmAction(m, "container-remove", ctr.ID, fmt.Sprintf("Remove container %s?", ctr.Name))
	m.Confirm.ConfirmAudit = beginAudit(m, "resource.container.delete", containerTarget(m, ctr.ID), "Remove container "+ctr.Name)
	return m, nil
}

func doLogAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Docker == nil {
		return m, nil
	}
	ctr := m.Resources.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	return m, ToLogView(m, ctr.ID)
}

func doStatsAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Docker == nil {
		return m, nil
	}
	ctr := m.Resources.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	m.Metrics.ToggleContainerStats()
	if m.Metrics.StatsActive {
		return m, FetchStats(m.Connection.Docker, ctr.ID)
	}
	return m, nil
}

func doExecAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Connection.Docker == nil {
		return m, nil
	}
	ctr := m.Resources.Containers.Selected()
	if ctr == nil {
		return m, nil
	}

	shell := m.Exec.ExecShell
	if shell == "" {
		shell = "/bin/sh"
	}
	m.Exec.ExecShell = ""
	trace := beginAudit(m, "resource.container.exec", audit.ExecTarget{ID: ctr.ID, Name: ctr.Name, Meta: audit.ExecMeta{ContainerID: ctr.ID}}, "Starting exec session in "+ctr.Name)

	cli := m.Connection.Docker.Raw()
	ctx := context.Background()

	execConfig := container.ExecOptions{
		Cmd:          []string{shell},
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}
	execCreate, err := cli.ContainerExecCreate(ctx, ctr.ID, execConfig)
	if err != nil {
		FinishAudit(m, trace, audit.ResultFailed, "Exec session creation failed", audit.Details{Error: err.Error(), Shell: shell})
		ShowToastNow(m, fmt.Sprintf("✕ exec create: %v", err))
		return m, nil
	}

	resp, err := cli.ContainerExecAttach(ctx, execCreate.ID, container.ExecAttachOptions{Tty: true})
	if err != nil {
		FinishAudit(m, trace, audit.ResultFailed, "Exec session attach failed", audit.Details{Error: err.Error(), Shell: shell})
		ShowToastNow(m, fmt.Sprintf("✕ exec attach: %v", err))
		return m, nil
	}

	if m.Viewport.Width > 0 && m.Viewport.Height > 0 {
		go cli.ContainerExecResize(context.Background(), execCreate.ID, container.ResizeOptions{
			Height: uint(m.Viewport.Height),
			Width:  uint(m.Viewport.Width),
		})
	}

	ch := make(chan string, 100)
	done := make(chan struct{})
	m.Exec.SetShell(shell)
	m.Exec.Start(execCreate.ID, resp.Conn, ch, done, trace)
	m.Navigation.Mode = state.ModeExecPassthrough

	// Reader goroutine: reads raw TTY output from exec attach and sends it on ch.
	go func() {
		buf := make([]byte, 4096)
		defer close(done)
		defer resp.Close()
		for {
			n, err := resp.Reader.Read(buf)
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
	if m.Connection.Docker == nil {
		return m, nil
	}
	ctr := m.Resources.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	rawJSON, err := m.Connection.Docker.InspectContainer(ctr.ID)
	if err != nil {
		m.Feedback.RecordError(err.Error())
		return m, nil
	}
	m.Detail.SetRaw(state.ResourceContainer, rawJSON)
	ToDetail(m, i18n.T("detail.title.container", ctr.Name, ctr.ID), "")
	return m, nil
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
	trace := beginAudit(m, "panel.runtime.switch", audit.RuntimeTarget{Name: next, Meta: audit.RuntimeMeta{Previous: current}}, fmt.Sprintf("Switching runtime from %s to %s", current, next))
	if err := m.Connection.Pool.Connect(next, 10*time.Second); err != nil {
		FinishAudit(m, trace, audit.ResultFailed, "Runtime switch failed", audit.Details{Error: err.Error()})
		ShowToastNow(m, fmt.Sprintf("switch failed: %v", err))
		return m, nil
	}
	m.Connection.Docker = m.Connection.Pool.ActiveClient()
	if m.Connection.Docker != nil {
		m.Connection.RuntimeType = string(m.Connection.Docker.RuntimeType)
		m.Connection.EngineVersion = m.Connection.Docker.EngineVersion
	}
	FinishAudit(m, trace, audit.ResultSucceeded, "Switched runtime to "+next, audit.Details{})
	cmds := FetchAll(m.Connection.Docker)
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
