package keyboard

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/elizabevil/docker-tui/internal/data/audit"
	"github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func doContainerAction(m *state.AppModel, action string, cmdFn func(*docker.Client, string) tea.Cmd) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil || m.ActivePanel != state.PanelContainers {
		return m, nil
	}
	if len(m.MarkedIDs) > 0 {
		return doBatchContainerAction(m, action, cmdFn)
	}
	ctr := m.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	// "start" is non-destructive — execute immediately
	if action == "start" {
		m.InfoMessage = "starting " + ctr.Name + "..."
		trace := beginAudit(m, "resource.container.start", containerTarget(m, ctr.ID), "Starting "+ctr.Name)
		return m, withContainerAudit(cmdFn(m.Docker, ctr.ID), trace)
	}
	// Destructive actions (stop/kill/restart) require confirmation
	confirmAction(m, "container-"+action, ctr.ID, fmt.Sprintf("%s container %s?", action, ctr.Name))
	m.ConfirmAudit = beginAudit(m, "resource.container."+action, containerTarget(m, ctr.ID), fmt.Sprintf("%s container %s", action, ctr.Name))
	return m, nil
}

func doBatchContainerAction(m *state.AppModel, action string, cmdFn func(*docker.Client, string) tea.Cmd) (*state.AppModel, tea.Cmd) {
	if len(m.MarkedIDs) == 0 {
		return m, nil
	}
	m.Mode = state.ModeConfirm
	m.ConfirmAction = "batch-" + action
	m.ConfirmTarget = fmt.Sprintf("%d items", len(m.MarkedIDs))
	m.ConfirmMessage = fmt.Sprintf("Batch %s %d containers?", action, len(m.MarkedIDs))
	m.ConfirmAudit = beginAudit(m, "resource.container."+action, audit.ContainerTarget{ID: fmt.Sprintf("batch-%d", len(m.MarkedIDs)), Name: fmt.Sprintf("%d containers", len(m.MarkedIDs))}, m.ConfirmMessage)
	return m, nil
}

func executeBatchAction(m *state.AppModel, action string, trace audit.Trace) (*state.AppModel, tea.Cmd) {
	ids := make([]string, 0, len(m.MarkedIDs))
	for id := range m.MarkedIDs {
		ids = append(ids, id)
	}
	m.MarkedIDs = make(map[string]bool)

	var cmdFn func(*docker.Client, string) tea.Cmd
	switch action {
	case "start":
		cmdFn = containerStartCmd
	case "stop":
		cmdFn = func(c *docker.Client, id string) tea.Cmd { return containerStopCmd(c, id) }
	case "restart":
		cmdFn = func(c *docker.Client, id string) tea.Cmd { return containerRestartCmd(c, id) }
	case "kill":
		cmdFn = func(c *docker.Client, id string) tea.Cmd { return containerKillCmd(c, id) }
	default:
		ShowToastNow(m, fmt.Sprintf("✕ unknown batch action: %s", action))
		return m, nil
	}

	var cmds []tea.Cmd
	for _, id := range ids {
		cmds = append(cmds, withContainerAudit(cmdFn(m.Docker, id), trace))
	}
	ShowToastNow(m, fmt.Sprintf("✓ batch %s %d containers", action, len(ids)))
	return m, tea.Batch(cmds...)
}

func doContainerRemove(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil || m.ActivePanel != state.PanelContainers {
		return m, nil
	}
	ctr := m.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	confirmAction(m, "container-remove", ctr.ID, fmt.Sprintf("Remove container %s?", ctr.Name))
	m.ConfirmAudit = beginAudit(m, "resource.container.delete", containerTarget(m, ctr.ID), "Remove container "+ctr.Name)
	return m, nil
}

func doLogAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
		return m, nil
	}
	ctr := m.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	return m, ToLogView(m, ctr.ID)
}

func doStatsAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
		return m, nil
	}
	ctr := m.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	m.StatsActive = !m.StatsActive
	if m.StatsActive {
		return m, FetchStats(m.Docker, ctr.ID)
	}
	return m, nil
}

func doExecAction(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Docker == nil {
		return m, nil
	}
	ctr := m.Containers.Selected()
	if ctr == nil {
		return m, nil
	}

	shell := m.ExecShell
	if shell == "" {
		shell = "/bin/sh"
	}
	m.ExecShell = ""
	trace := beginAudit(m, "resource.container.exec", audit.ExecTarget{ID: ctr.ID, Name: ctr.Name, Meta: audit.ExecMeta{ContainerID: ctr.ID}}, "Starting exec session in "+ctr.Name)

	cli := m.Docker.Raw()
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

	m.ExecID = execCreate.ID
	m.ExecAudit = trace
	m.ExecShell = shell
	m.ExecConn = resp.Conn
	m.Mode = state.ModeExecPassthrough

	if m.Width > 0 && m.Height > 0 {
		go cli.ContainerExecResize(context.Background(), execCreate.ID, container.ResizeOptions{
			Height: uint(m.Height),
			Width:  uint(m.Width),
		})
	}

	ch := make(chan string, 100)
	done := make(chan struct{})
	m.ExecCh = ch
	m.ExecDone = done

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
	if m.Docker == nil {
		return m, nil
	}
	ctr := m.Containers.Selected()
	if ctr == nil {
		return m, nil
	}
	info, err := m.Docker.InspectContainer(ctr.ID)
	if err != nil {
		m.ErrorMessage = err.Error()
		m.ErrorCount++
		return m, nil
	}
	ToDetail(m, "Container Detail: "+ctr.Name+" ("+ctr.ID+")", info)
	return m, nil
}

func doSwitchRuntime(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.Pool == nil {
		ShowToastNow(m, "no runtime pool")
		return m, nil
	}
	names := m.Pool.KnownHostNames()
	if len(names) < 2 {
		ShowToastNow(m, fmt.Sprintf("only one runtime (%s)", m.Pool.ActiveName()))
		return m, nil
	}
	current := m.Pool.ActiveName()
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
	if err := m.Pool.Connect(next, 10*time.Second); err != nil {
		FinishAudit(m, trace, audit.ResultFailed, "Runtime switch failed", audit.Details{Error: err.Error()})
		ShowToastNow(m, fmt.Sprintf("switch failed: %v", err))
		return m, nil
	}
	m.Docker = m.Pool.ActiveClient()
	if m.Docker != nil {
		m.RuntimeType = string(m.Docker.RuntimeType)
		m.EngineVersion = m.Docker.EngineVersion
	}
	FinishAudit(m, trace, audit.ResultSucceeded, "Switched runtime to "+next, audit.Details{})
	cmds := FetchAll(m.Docker)
	cmds = append(cmds, ShowKeyHint(m, fmt.Sprintf("F2: Switched to %s (%s)", next, m.RuntimeType)))
	return m, tea.Batch(cmds...)
}

func doContainerSort(m *state.AppModel) (*state.AppModel, tea.Cmd) {
	if m.ActivePanel != state.PanelContainers {
		return m, nil
	}
	// Cycle: Name → ID → CPU → Mem → State → Created → Name…
	if m.Containers.SortBy >= state.ContainerSortByCreated {
		m.Containers.SortBy = state.ContainerSortByName
		m.Containers.SortAsc = !m.Containers.SortAsc
	} else {
		m.Containers.SortBy++
	}
	m.Containers.Cursor = 0
	m.Containers.ViewOffset = 0
	labels := map[state.ContainerSortColumn]string{
		state.ContainerSortByName:    "name",
		state.ContainerSortByID:      "id",
		state.ContainerSortByCPU:     "cpu",
		state.ContainerSortByMem:     "mem",
		state.ContainerSortByState:   "state",
		state.ContainerSortByCreated: "created",
	}
	m.InfoMessage = fmt.Sprintf("Sort by %s (%v)", labels[m.Containers.SortBy], m.Containers.SortAsc)
	return m, nil
}
