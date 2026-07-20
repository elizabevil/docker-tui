package keyboard

import (
	"bufio"

	"github.com/elizabevil/docker-tui/internal/data/docker"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	tea "charm.land/bubbletea/v2"
)

func containerStartCmd(client *docker.Client, id string) tea.Cmd {
	return func() tea.Msg {
		err := client.ContainerStart(id)
		return state.ContainerActioned{Action: state.ActionStarted, ID: id, Success: err == nil, Error: err}
	}
}

func containerStopCmd(client *docker.Client, id string) tea.Cmd {
	return func() tea.Msg {
		err := client.ContainerStop(id)
		return state.ContainerActioned{Action: state.ActionStopped, ID: id, Success: err == nil, Error: err}
	}
}

func containerRestartCmd(client *docker.Client, id string) tea.Cmd {
	return func() tea.Msg {
		err := client.ContainerRestart(id)
		return state.ContainerActioned{Action: state.ActionRestarted, ID: id, Success: err == nil, Error: err}
	}
}

func containerKillCmd(client *docker.Client, id string) tea.Cmd {
	return func() tea.Msg {
		err := client.ContainerKill(id)
		return state.ContainerActioned{Action: state.ActionKilled, ID: id, Success: err == nil, Error: err}
	}
}

func containerRemoveCmd(client *docker.Client, id string, force bool) tea.Cmd {
	return func() tea.Msg {
		err := client.ContainerRemove(id, force)
		return state.ContainerActioned{Action: state.ActionRemoved, ID: id, Success: err == nil, Error: err}
	}
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

func FetchStats(client *docker.Client, containerID string) tea.Cmd {
	return func() tea.Msg {
		reader, err := client.ContainerStats(containerID)
		if err != nil {
			return state.StatsReceived{ContainerID: containerID, Error: err}
		}
		defer reader.Close()

		cpu, memUsage, memLimit, memPerc, netRx, netTx := docker.ParseStats(reader)
		return state.StatsReceived{
			ContainerID: containerID,
			CPU:         cpu,
			MemUsage:    memUsage,
			MemLimit:    memLimit,
			MemPerc:     memPerc,
			NetRx:       netRx,
			NetTx:       netTx,
		}
	}
}
