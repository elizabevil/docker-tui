package state

import (
	"testing"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func TestMetricsStateApplyHostAndToggle(t *testing.T) {
	var s MetricsState
	s.ToggleContainerStats()
	if !s.StatsActive {
		t.Fatal("container stats were not enabled")
	}
	want := dockerclient.HostStats{CPUPercent: 12.5, CPUCores: 8, MemPercent: 50, MemUsed: 10, MemTotal: 20, DiskStr: "1G/2G"}
	s.ApplyHost(want)
	if s.HostCPU != want.CPUPercent || s.HostCPUCores != want.CPUCores || s.HostMem != want.MemPercent || s.HostMemUsed != want.MemUsed || s.HostMemTotal != want.MemTotal || s.HostDisk != want.DiskStr {
		t.Fatalf("host metrics = %#v", s)
	}
}
