package dto

// StatsJSON is the Podman/Docker container stats JSON response. All
// nested shapes are named so callers don't depend on anonymous structs.
type StatsJSON struct {
	CPUStats    StatsCPU                `json:"cpu_stats"`
	PreCPUStats StatsCPU                `json:"precpu_stats"`
	MemoryStats StatsMemory             `json:"memory_stats"`
	Networks    map[string]StatsNetwork `json:"networks"`
}

// StatsCPU captures the CPU usage portion of a container stats payload.
type StatsCPU struct {
	CPUUsage       StatsCPUUsage `json:"cpu_usage"`
	SystemCPUUsage uint64        `json:"system_cpu_usage"`
	OnlineCPUs     uint32        `json:"online_cpus"`
}

// StatsCPUUsage holds per-CPU usage counters.
type StatsCPUUsage struct {
	TotalUsage uint64 `json:"total_usage"`
}

// StatsMemory captures the memory usage portion of a container stats payload.
type StatsMemory struct {
	Usage uint64 `json:"usage"`
	Limit uint64 `json:"limit"`
}

// StatsNetwork holds per-interface network I/O counters for container stats.
type StatsNetwork struct {
	RxBytes uint64 `json:"rx_bytes"`
	TxBytes uint64 `json:"tx_bytes"`
}
