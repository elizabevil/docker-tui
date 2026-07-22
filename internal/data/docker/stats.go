package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// NetworkStats holds per-interface network I/O counters.
type NetworkStats struct {
	RxBytes uint64 `json:"rx_bytes"`
	TxBytes uint64 `json:"tx_bytes"`
}

type statsJSON struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     uint32 `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
	Networks map[string]NetworkStats `json:"networks"`
}

// ParseStats parses a Docker stats JSON response. Returns CPU%, memUsage, memLimit, mem%, netRx, netTx.
func ParseStats(reader io.ReadCloser) (cpuPerc, memUsage, memLimit, memPerc, netRx, netTx float64) {
	var stats statsJSON
	if err := json.NewDecoder(reader).Decode(&stats); err != nil {
		return 0, 0, 0, 0, 0, 0
	}
	return computeStats(stats)
}

func (c *Client) containerStatsContext(ctx context.Context, id string) (runtimeapi.ContainerStats, error) {
	if c.RuntimeType == RuntimePodman {
		return c.containerStatsPodmanREST(ctx, id)
	}
	response, err := c.cli.ContainerStats(ctx, id, false)
	if err != nil {
		return runtimeapi.ContainerStats{}, fmt.Errorf("stats container: %w", err)
	}
	defer response.Body.Close()
	var raw statsJSON
	if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
		return runtimeapi.ContainerStats{}, fmt.Errorf("decode container stats: %w", err)
	}
	cpu, usage, limit, memory, rx, tx := computeStats(raw)
	return runtimeapi.ContainerStats{ReadAt: time.Now(), CPUPercent: cpu, MemoryUsage: usage, MemoryLimit: limit, MemoryPercent: memory, NetworkRx: rx, NetworkTx: tx}, nil
}

func computeStats(stats statsJSON) (cpuPerc, memUsage, memLimit, memPerc, netRx, netTx float64) {
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
