package runtime

import (
	"fmt"
	"math"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// HostStats holds detailed host system information.
type HostStats struct {
	CPUPercent float64
	CPUCores   int
	MemPercent float64
	MemUsed    uint64 // bytes
	MemTotal   uint64 // bytes
	DiskStr    string
}

// ReadHostStats reads host system metrics using gopsutil.
// All underlying probes are best-effort; individual failures are
// tolerated so the UI degrades gracefully instead of crashing.
func ReadHostStats() HostStats {
	cores, err := cpu.Counts(true)
	if err != nil {
		cores = 0
	}

	cpuPercents, err := cpu.Percent(500*time.Millisecond, false) //nolint:errcheck // best-effort sample.
	cpuPct := 0.0
	if err == nil && len(cpuPercents) > 0 {
		cpuPct = math.Round(cpuPercents[0]*100) / 100
	}

	memInfo, err := mem.VirtualMemory()
	memPct := 0.0
	if err == nil {
		memPct = math.Round(memInfo.UsedPercent*100) / 100
	}

	diskUsage, err := disk.Usage("/")
	diskStr := "—"
	if err == nil && diskUsage != nil {
		avail := formatBytes(diskUsage.Free)
		total := formatBytes(diskUsage.Total)
		diskStr = avail + "/" + total
	}

	used, total64 := uint64(0), uint64(0)
	if memInfo != nil {
		used = memInfo.Used
		total64 = memInfo.Total
	}
	return HostStats{
		CPUPercent: cpuPct,
		CPUCores:   cores,
		MemPercent: memPct,
		MemUsed:    used,
		MemTotal:   total64,
		DiskStr:    diskStr,
	}
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return "0B"
	}
	div := uint64(unit)
	exp := 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", float64(b)/float64(div), "KMGT"[exp])
}
