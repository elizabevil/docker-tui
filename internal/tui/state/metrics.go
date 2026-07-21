package state

import dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"

type MetricsState struct {
	StatsActive  bool
	HostCPU      float64
	HostMem      float64
	HostDisk     string
	HostCPUCores int
	HostMemUsed  uint64
	HostMemTotal uint64
}

func (s *MetricsState) ToggleContainerStats() { s.StatsActive = !s.StatsActive }

func (s *MetricsState) ApplyHost(stats dockerclient.HostStats) {
	s.HostCPU = stats.CPUPercent
	s.HostMem = stats.MemPercent
	s.HostDisk = stats.DiskStr
	s.HostCPUCores = stats.CPUCores
	s.HostMemUsed = stats.MemUsed
	s.HostMemTotal = stats.MemTotal
}
