package runtime

import (
	"strings"
	"time"
)

// ComposeLogOptions configures multi-container log aggregation
// (R08-06 §F1-F7). Fetched by FetchComposeLogsBatch.
type ComposeLogOptions struct {
	// Since is the RFC3339 since filter forwarded to the engine.
	Since string
	// Tail is the line-count tail forwarded to the engine.
	Tail string
	// Timestamps toggles the [HH:MM:SS.mmm] time-of-day prefix in
	// the rendered output. Internally the engine is always queried
	// with timestamps=true so the time-sorted aggregation has data
	// to sort on; this flag only controls whether the parsed time
	// surfaces in the visible line.
	Timestamps bool
	// Merged collapses all replicas of a service under a single
	// [service] tag instead of the default [service.N] suffix
	// per replica (R08-06 §F3).
	Merged bool
}

// ComposeLogLine is the record produced during aggregation (R08-06
// §F2). It carries the metadata needed for time-sorted display:
// which service/replica produced it, when, and the raw container
// output with the engine-prepended timestamp stripped.
type ComposeLogLine struct {
	Service     string
	Replica     int       // 1-based replica index within the service
	ContainerID string
	Timestamp   time.Time // zero value if engine did not provide one
	Line        string    // raw log line, no prefix
}

// LogTimeFormat is the time-of-day format used in the
// [HH:MM:SS.mmm] prefix (R08-06 §F4).
const LogTimeFormat = "15:04:05.000"

// ParseLogLineTimestamp extracts the engine-prepended RFC3339Nano
// timestamp from a log line. Returns (timestamp, content). If no
// timestamp is present (or unparseable), returns (zero, line) so
// the caller can emit a non-timestamped line and let the sort
// collector push it to the tail of the merged stream.
func ParseLogLineTimestamp(line string) (time.Time, string) {
	spaceIdx := strings.IndexByte(line, ' ')
	if spaceIdx == -1 || spaceIdx > 40 { // RFC3339Nano ≤ 35 chars; allow slack.
		return time.Time{}, line
	}
	ts, err := time.Parse(time.RFC3339Nano, line[:spaceIdx])
	if err != nil {
		return time.Time{}, line
	}
	return ts, line[spaceIdx+1:]
}
