package keyboard

import (
	"bufio"
	"context"

	"github.com/elizabevil/docker-tui/internal/data/docker"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func containerStartCmd(client *docker.Client, id string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionStart, runtimeapi.ActionOptions{})
		return state.ContainerActioned{Action: state.ActionStarted, ID: id, Success: err == nil, Error: err}
	}
}

func containerStopCmd(client *docker.Client, id string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionStop, runtimeapi.ActionOptions{})
		return state.ContainerActioned{Action: state.ActionStopped, ID: id, Success: err == nil, Error: err}
	}
}

func containerRestartCmd(client *docker.Client, id string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionRestart, runtimeapi.ActionOptions{})
		return state.ContainerActioned{Action: state.ActionRestarted, ID: id, Success: err == nil, Error: err}
	}
}

func containerKillCmd(client *docker.Client, id string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionKill, runtimeapi.ActionOptions{})
		return state.ContainerActioned{Action: state.ActionKilled, ID: id, Success: err == nil, Error: err}
	}
}

func containerPauseCmd(client *docker.Client, id string, unpause bool) tea.Cmd {
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

func containerRenameCmd(client *docker.Client, id, name string) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionRename, runtimeapi.ActionOptions{Name: name})
		return state.ContainerActioned{Action: state.ActionRenamed, ID: id, Success: err == nil, Error: err}
	}
}

func fetchContainerProcesses(service runtimeapi.ContainerService, id string) tea.Cmd {
	return func() tea.Msg {
		processes, err := service.Top(context.Background(), id)
		return state.ContainerProcessesLoaded{ContainerID: id, Processes: processes, Error: err}
	}
}

func containerRemoveCmd(client *docker.Client, id string, force bool) tea.Cmd {
	return func() tea.Msg {
		_, err := client.Actions().Execute(context.Background(), containerRef(id), runtimeapi.ActionRemove, runtimeapi.ActionOptions{Force: force})
		return state.ContainerActioned{Action: state.ActionRemoved, ID: id, Success: err == nil, Error: err}
	}
}

func containerRef(id string) runtimeapi.ResourceRef {
	return runtimeapi.ResourceRef{Type: runtimeapi.ResourceContainer, ID: id}
}

func FetchLogBatch(client *docker.Client, containerID, since, tail string, ts bool) tea.Cmd {
	return func() tea.Msg {
		reader, err := client.ContainerLogs(containerID, since, tail, ts)
		if err != nil {
			return state.LogStreamError{ContainerID: containerID, Error: err}
		}
		defer reader.Close()

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
