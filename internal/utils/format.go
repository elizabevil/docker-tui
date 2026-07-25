package utils

import (
	"fmt"
	"strings"
	"time"
)

// ShortID truncates an ID string to its first 12 characters.
func ShortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

type SizeFormat = string

const (
	SizeFormatSi SizeFormat = "si"
)

// sizeFormatSI controls whether FormatSize uses SI (1000-base, matching podman/docker CLI)
// or binary (1024-base, default). Set via SetSizeFormat at startup.
var sizeFormatSI bool

// SetSizeFormat controls the base used by FormatSize.
//
//	false — 1024-base binary (default): 1 MB = 1048576 bytes
//	true  — 1000-base SI:              1 MB = 1000000 bytes (matches podman/docker CLI)
func SetSizeFormat(si SizeFormat) { sizeFormatSI = SizeFormatSi == si }

// FormatSize formats a byte count as a human-readable string.
// Default uses 1024-base (KiB/MiB/GiB); SetSizeFormat(true) switches to 1000-base (kB/MB/GB).
func FormatSize(bytes int64) string {
	var unit int64 = 1024
	prefixes := []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	if sizeFormatSI {
		unit = 1000
		prefixes = []string{"kB", "MB", "GB", "TB", "PB", "EB"}
	}
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := unit, 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	if exp >= len(prefixes) {
		exp = len(prefixes) - 1
	}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), prefixes[exp])
}

// FormatBytes formats a float64 byte count (e.g., network bytes).
func FormatBytes(b float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for b >= 1024 && i < len(units)-1 {
		b /= 1024
		i++
	}
	return fmt.Sprintf("%.1f%s", b, units[i])
}

// ShortImage extracts the image name from a full image reference (e.g. "docker.io/library/nginx:latest" → "nginx:latest").
func ShortImage(ref string) string {
	idx := strings.LastIndexByte(ref, '/')
	if idx >= 0 && idx < len(ref)-1 {
		return ref[idx+1:]
	}
	return ref
}

// FormatPercent formats a float64 as a percentage string like "25%".
func FormatPercent(v float64) string {
	return fmt.Sprintf("%.0f%%", v)
}

// FormatCreated formats a Unix timestamp as elapsed time from now.
// Examples: "1y 23d 12h 30m 15s", "5h 3m 10s", "30s".
func FormatCreated(unix int64) string {
	return time.Unix(unix, 0).Format(time.DateTime)
}

// Truncate cuts a string to maxLen characters, appending "..." if truncated.
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(r[:maxLen])
	}
	return string(r[:maxLen-3]) + "..."
}
