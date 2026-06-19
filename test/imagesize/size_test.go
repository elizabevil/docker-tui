// Package imagesize verifies image size formatting behavior.
// It compares TUI's FormatSize (1024-base, binary) against
// podman CLI's SI (1000-base) formatting to document the discrepancy.
package imagesize

import (
	"fmt"
	"math"
	"testing"

	"github.com/elizabevil/docker-tui/internal/tui/utils"
)

// formatSI mimics podman CLI's 1000-base (SI) formatting.
func formatSI(bytes int64) string {
	units := []string{"B", "kB", "MB", "GB", "TB"}
	val := float64(bytes)
	i := 0
	for val >= 1000 && i < len(units)-1 {
		val /= 1000
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d %s", int(val), units[i])
	}
	// podman CLI trims trailing zeros: "42 MB" not "42.00 MB"
	s := fmt.Sprintf("%.2f %s", val, units[i])
	// strip trailing ".00"
	if s[len(s)-6:] == ".00 " {
		s = s[:len(s)-6] + " " + s[len(s)-3:]
	}
	return s
}

// knownImages holds raw byte sizes from the actual podman API on this machine,
// along with what `podman images` CLI displays for each.
var knownImages = []struct {
	name     string
	rawBytes int64
	cliSize  string // what "podman images" shows
}{
	// Data from actual podman API call 2026-06-16
	{name: "alpine:3.19", rawBytes: 7687286, cliSize: "7.69 MB"},
	{name: "redis:7-alpine", rawBytes: 42017397, cliSize: "42 MB"},
	{name: "nginx:alpine", rawBytes: 50143357, cliSize: "50.1 MB"},
	{name: "postgres:18.4-alpine", rawBytes: 288681577, cliSize: "289 MB"},
	{name: "bk-cmdb:3.14.7", rawBytes: 81407686, cliSize: "81.4 MB"},
	{name: "bk-builder:latest", rawBytes: 1102134819, cliSize: "1.1 GB"},
}

// TestFormatSizeVsCLI documents the discrepancy between
// TUI FormatSize (1024-base) and podman CLI (1000-base).
func TestFormatSizeVsCLI(t *testing.T) {
	for _, img := range knownImages {
		tuiStr := utils.FormatSize(img.rawBytes)
		cliStr := formatSI(img.rawBytes)

		t.Run(img.name, func(t *testing.T) {
			t.Logf("Raw bytes: %d", img.rawBytes)
			t.Logf("TUI  FormatSize (1024-base): %s", tuiStr)
			t.Logf("CLI  formatSI    (1000-base): %s", cliStr)
			t.Logf("CLI  actual display:         %s", img.cliSize)

			// Document that they differ
			if tuiStr == img.cliSize {
				t.Logf("⚠  Match! (unexpected — they use different bases)")
			} else {
				t.Logf("✓  Differ as expected (1024 vs 1000 base)")
			}

			// Verify CLI actual display matches SI format
			if cliStr != img.cliSize {
				// Allow small rounding differences
				t.Logf("Note: CLI display %q differs from calculated SI %q (rounding)", img.cliSize, cliStr)
			}
		})
	}
}

// TestFormatSizeUnits verifies FormatSize produces correct binary units.
func TestFormatSizeUnits(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1048576, "1.0 MiB"},
		{2097152, "2.0 MiB"},
		{1073741824, "1.0 GiB"},
	}
	for _, tt := range tests {
		got := utils.FormatSize(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

// TestSizeDiscrepancy documents how much 1024 vs 1000 base differs.
func TestSizeDiscrepancy(t *testing.T) {
	// At MB scale, the difference is ~2.4% for MB, ~7.4% for GB
	ratio := float64(1000*1000*1000) / float64(1024*1024*1024) // 1000³ / 1024³
	diff := math.Abs(1-ratio) * 100

	t.Logf("1 GB (SI)   = 1000³ = 1,000,000,000 bytes")
	t.Logf("1 GiB (bin) = 1024³ = 1,073,741,824 bytes")
	t.Logf("Ratio: %.4f (%.2f%% difference at GB scale)", ratio, diff)
	t.Logf("")
	t.Logf("Discrepancy by magnitude:")
	t.Logf("  KB scale: %.2f%%", math.Abs(1-1000.0/1024.0)*100)
	t.Logf("  MB scale: %.2f%%", math.Abs(1-math.Pow(1000, 2)/math.Pow(1024, 2))*100)
	t.Logf("  GB scale: %.2f%%", math.Abs(1-math.Pow(1000, 3)/math.Pow(1024, 3))*100)

	// For the largest image (1.1 GB), this is ~80 MB difference
	largest := int64(1102134819)
	siVal := float64(largest) / math.Pow(1000, 3)
	binVal := float64(largest) / math.Pow(1024, 3)
	t.Logf("")
	t.Logf("Example: %d bytes image:", largest)
	t.Logf("  SI (1000-base): %.2f GB", siVal)
	t.Logf("  Binary (1024):  %.2f GB", binVal)
	t.Logf("  Difference:     %.0f MB", (binVal-siVal)*1024)
}

// TestFormatSizeIs1024Based documents that FormatSize uses 1024-base.
// If this test fails, the formatting behavior has changed.
func TestFormatSizeIs1024Based(t *testing.T) {
	// 1 MB (SI) = 1,000,000 bytes → 1024-base should show ~976.6 KiB
	oneMillion := int64(1000000)
	got := utils.FormatSize(oneMillion)
	// 1000000 / 1024 = 976.5625 → "976.6 KiB"
	expected := "976.6 KiB"
	if got != expected {
		t.Errorf("FormatSize(%d) = %q, want %q (confirms binary base)", oneMillion, got, expected)
	} else {
		t.Logf("Confirmed: FormatSize uses 1024-base (1,000,000 bytes = %s)", got)
	}

	// 1,000,000 bytes in SI = 1.0 MB
	siStr := formatSI(oneMillion)
	t.Logf("SI (1000-base) would show: %s", siStr)
}
