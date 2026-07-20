package docker

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
func ReadHostStats() HostStats {
	cores, _ := cpu.Counts(true)

	// CPU: 500ms sample interval for accurate reading
	cpuPercents, _ := cpu.Percent(500*time.Millisecond, false)
	cpuPct := 0.0
	if len(cpuPercents) > 0 {
		cpuPct = math.Round(cpuPercents[0]*100) / 100
	}

	// Memory
	memInfo, _ := mem.VirtualMemory()
	memPct := math.Round(memInfo.UsedPercent*100) / 100

	// Disk
	diskUsage, _ := disk.Usage("/")
	diskStr := "—"
	if diskUsage != nil {
		avail := formatBytes(diskUsage.Free)
		total := formatBytes(diskUsage.Total)
		diskStr = avail + "/" + total
	}

	return HostStats{
		CPUPercent: cpuPct,
		CPUCores:   cores,
		MemPercent: memPct,
		MemUsed:    memInfo.Used,
		MemTotal:   memInfo.Total,
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
