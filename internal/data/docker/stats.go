package docker

import (
	"encoding/json"
	"io"
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
