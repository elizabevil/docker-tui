package podman

import "github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"

// ComputeStats calculates CPU%, memory usage/limit/percent, and network
// I/O from a raw stats JSON response.
func ComputeStats(stats dto.StatsJSON) (cpuPerc, memUsage, memLimit, memPerc, netRx, netTx float64) {
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
