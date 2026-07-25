package docker

import (
	"encoding/json"
	"testing"
)

// TestStatsResponseRoundTrip verifies the renamed StatsResponse / StatsCPU
// / StatsCPUUsage / StatsMemory / NetworkStats types correctly decode and
// re-encode the Docker /containers/{id}/stats payload. TASK-022 Phase D
// renamed these from anonymous nested structs.
func TestStatsResponseRoundTrip(t *testing.T) {
	raw := `{
		"cpu_stats": {
			"cpu_usage": {"total_usage": 100},
			"system_cpu_usage": 1000,
			"online_cpus": 4
		},
		"precpu_stats": {
			"cpu_usage": {"total_usage": 50},
			"system_cpu_usage": 800
		},
		"memory_stats": {"usage": 2048, "limit": 4096},
		"networks": {"eth0": {"rx_bytes": 10, "tx_bytes": 20}}
	}`
	var s StatsResponse
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if s.CPUStats.CPUUsage.TotalUsage != 100 {
		t.Errorf("TotalUsage = %d", s.CPUStats.CPUUsage.TotalUsage)
	}
	if s.CPUStats.SystemCPUUsage != 1000 {
		t.Errorf("SystemCPUUsage = %d", s.CPUStats.SystemCPUUsage)
	}
	if s.CPUStats.OnlineCPUs != 4 {
		t.Errorf("OnlineCPUs = %d", s.CPUStats.OnlineCPUs)
	}
	if s.PreCPUStats.CPUUsage.TotalUsage != 50 {
		t.Errorf("PreCPU TotalUsage = %d", s.PreCPUStats.CPUUsage.TotalUsage)
	}
	if s.MemoryStats.Usage != 2048 || s.MemoryStats.Limit != 4096 {
		t.Errorf("Memory = %+v", s.MemoryStats)
	}
	if n := s.Networks["eth0"]; n.RxBytes != 10 || n.TxBytes != 20 {
		t.Errorf("Networks[eth0] = %+v", n)
	}

	// Encode back and check we get the same shape.
	out, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var again StatsResponse
	if err := json.Unmarshal(out, &again); err != nil {
		t.Fatalf("roundtrip unmarshal: %v", err)
	}
	if again.CPUStats != s.CPUStats {
		t.Errorf("CPUStats roundtrip differs")
	}
	if again.PreCPUStats != s.PreCPUStats {
		t.Errorf("PreCPUStats roundtrip differs")
	}
	if again.MemoryStats != s.MemoryStats {
		t.Errorf("MemoryStats roundtrip differs")
	}
	if len(again.Networks) != len(s.Networks) {
		t.Errorf("Networks length differs: %d vs %d", len(again.Networks), len(s.Networks))
	}
	for k, v := range s.Networks {
		if again.Networks[k] != v {
			t.Errorf("network[%s] differs", k)
		}
	}
}

// TestComputeStatsNoSystemDelta confirms that computeStats returns the
// zero CPU percent when the system delta is non-positive (avoiding the
// classic NaN produced by dividing by zero).
func TestComputeStatsNoSystemDelta(t *testing.T) {
	stats := StatsResponse{
		CPUStats:    StatsCPU{CPUUsage: StatsCPUUsage{TotalUsage: 100}, SystemCPUUsage: 100, OnlineCPUs: 4},
		PreCPUStats: StatsCPU{CPUUsage: StatsCPUUsage{TotalUsage: 100}, SystemCPUUsage: 100},
		MemoryStats: StatsMemory{Usage: 1024, Limit: 2048},
	}
	cpu, usage, limit, perc, rx, tx := computeStats(stats)
	if cpu != 0 {
		t.Errorf("expected cpu=0 with zero delta, got %v", cpu)
	}
	if usage != 1024 || limit != 2048 || perc != 50 {
		t.Errorf("memory math: usage=%v limit=%v perc=%v", usage, limit, perc)
	}
	if rx != 0 || tx != 0 {
		t.Errorf("net rx/tx = (%v, %v), want 0", rx, tx)
	}
}

// TestComputeStatsAccumulatesNetworks confirms per-interface byte counters
// sum across all networks.
func TestComputeStatsAccumulatesNetworks(t *testing.T) {
	stats := StatsResponse{
		Networks: map[string]NetworkStats{
			"eth0": {RxBytes: 100, TxBytes: 50},
			"eth1": {RxBytes: 200, TxBytes: 75},
		},
	}
	_, _, _, _, rx, tx := computeStats(stats)
	if rx != 300 || tx != 125 {
		t.Errorf("net totals = (%v, %v), want (300, 125)", rx, tx)
	}
}
