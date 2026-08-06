package keyboard

import (
	"bufio"
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func containerStartCmd(client runtimeapi.Engine, id string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionStart, runtimeapi.ActionOptions{})
		return state.ContainerActioned{Action: state.ActionStarted, ID: id, Success: err == nil, Error: err}
	}
}

func containerStopCmd(client runtimeapi.Engine, id string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionStop, runtimeapi.ActionOptions{})
		return state.ContainerActioned{Action: state.ActionStopped, ID: id, Success: err == nil, Error: err}
	}
}

func containerRestartCmd(client runtimeapi.Engine, id string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionRestart, runtimeapi.ActionOptions{})
		return state.ContainerActioned{Action: state.ActionRestarted, ID: id, Success: err == nil, Error: err}
	}
}

func containerKillCmd(client runtimeapi.Engine, id string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionKill, runtimeapi.ActionOptions{})
		return state.ContainerActioned{Action: state.ActionKilled, ID: id, Success: err == nil, Error: err}
	}
}

func containerPauseCmd(client runtimeapi.Engine, id string, unpause bool) tea.Cmd {
	return func() tea.Msg {
		action := state.ActionPaused
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionPause, runtimeapi.ActionOptions{})
		if unpause {
			action = state.ActionUnpaused
			_, err = client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionUnpause, runtimeapi.ActionOptions{})
		}
		return state.ContainerActioned{Action: action, ID: id, Success: err == nil, Error: err}
	}
}

func containerRenameCmd(client runtimeapi.Engine, id, name string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionRename,
			runtimeapi.ActionOptions{Lifecycle: runtimeapi.LifecycleOptions{Name: name}})
		return state.ContainerActioned{Action: state.ActionRenamed, ID: id, Success: err == nil, Error: err}
	}
}

func fetchContainerProcesses(service runtimeapi.ContainerService, id string) tea.Cmd {
	return func() tea.Msg {
		processes, err := service.Top(context.Background(), id)
		return state.ContainerProcessesLoaded{ContainerID: id, Processes: processes, Error: err}
	}
}

// FetchContainerProcesses exposes the Top refresh command to the update
// ticker while keeping the runtime request construction in the keyboard layer.
func FetchContainerProcesses(service runtimeapi.ContainerService, id string) tea.Cmd {
	return fetchContainerProcesses(service, id)
}

func containerRemoveCmd(client runtimeapi.Engine, id string, opts runtimeapi.LifecycleOptions) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionRemove,
			runtimeapi.ActionOptions{Lifecycle: opts})
		return state.ContainerActioned{Action: state.ActionRemoved, ID: id, Success: err == nil, Error: err}
	}
}

func containerRef(id string) runtimeapi.ResourceRef {
	return runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}
}

func FetchLogBatch(client runtimeapi.Engine, containerID, since, tail string, ts bool) tea.Cmd {
	return func() tea.Msg {
		reader, err := client.Containers().Logs(context.Background(), containerID, runtimeapi.ContainerLogOptions{Since: since, Tail: tail, Timestamps: ts})
		if err != nil {
			return state.LogStreamError{ContainerID: containerID, Error: err}
		}
		defer func() { _ = reader.Close() }() //nolint:errcheck // log reader fully scanned before close.

		var lines []string
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) > 0 {
				lines = append(lines, line)
			}
		}
		return state.LogBatchReceived{ContainerID: containerID, Lines: lines}
	}
}

func FetchStats(service runtimeapi.ContainerService, containerID string) tea.Cmd {
	return func() tea.Msg {
		stats, err := service.Stats(context.Background(), containerID)
		if err != nil {
			return state.StatsReceived{ContainerID: containerID, Error: err}
		}
		return state.StatsReceived{
			ContainerID: containerID,
			CPU:         stats.CPUPercent,
			MemUsage:    stats.MemoryUsage,
			MemLimit:    stats.MemoryLimit,
			MemPerc:     stats.MemoryPercent,
			NetRx:       stats.NetworkRx,
			NetTx:       stats.NetworkTx,
		}
	}
}
