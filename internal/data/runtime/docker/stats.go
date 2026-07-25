package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// NetworkStats holds per-interface network I/O counters.
type NetworkStats struct {
	RxBytes uint64 `json:"rx_bytes"`
	TxBytes uint64 `json:"tx_bytes"`
}

// StatsCPUUsage is the CPU usage portion of a Docker stats sample.
type StatsCPUUsage struct {
	TotalUsage uint64 `json:"total_usage"`
}

// StatsCPU is the CPU stats portion of a Docker stats sample.
type StatsCPU struct {
	CPUUsage       StatsCPUUsage `json:"cpu_usage"`
	SystemCPUUsage uint64        `json:"system_cpu_usage"`
	OnlineCPUs     uint32        `json:"online_cpus"`
}

// StatsMemory is the memory stats portion of a Docker stats sample.
type StatsMemory struct {
	Usage uint64 `json:"usage"`
	Limit uint64 `json:"limit"`
}

// StatsResponse is the JSON shape returned by Docker's /containers/{id}/stats.
// TASK-022 Phase D: previously nested anonymous structs; renamed so each
// sub-shape can be referenced and tested directly.
type StatsResponse struct {
	CPUStats    StatsCPU            `json:"cpu_stats"`
	PreCPUStats StatsCPU            `json:"precpu_stats"`
	MemoryStats StatsMemory         `json:"memory_stats"`
	Networks    map[string]NetworkStats `json:"networks"`
}

func (c *Client) containerStatsContext(ctx context.Context, id string) (runtimeapi.ContainerStats, error) {
	response, err := c.cli.ContainerStats(ctx, id, false)
	if err != nil {
		return runtimeapi.ContainerStats{}, fmt.Errorf("stats container: %w", err)
	}
	defer response.Body.Close()
	var raw StatsResponse
	if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
		return runtimeapi.ContainerStats{}, fmt.Errorf("decode container stats: %w", err)
	}
	cpu, usage, limit, memory, rx, tx := computeStats(raw)
	return runtimeapi.ContainerStats{ReadAt: time.Now(), CPUPercent: cpu, MemoryUsage: usage, MemoryLimit: limit, MemoryPercent: memory, NetworkRx: rx, NetworkTx: tx}, nil
}

func computeStats(stats StatsResponse) (cpuPerc, memUsage, memLimit, memPerc, netRx, netTx float64) {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage)
	if systemDelta > 0 && cpuDelta > 0 {
		cpuPerc = (cpuDelta / systemDelta) * float64(stats.CPUStats.OnlineCPUs) * 100
	}

	memUsage = float64(stats.MemoryStats.Usage)
	memLimit = float64(stats.MemoryStats.Limit)
	if memLimit > 0 {
		memPerc = (memUsage / memLimit) * 100
	}

	for _, n := range stats.Networks {
		netRx += float64(n.RxBytes)
		netTx += float64(n.TxBytes)
	}

	return
}
