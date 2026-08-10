package utils

import "time"

// Time format constants grouped by purpose.
//
// Each constant has a single, documented use case. Pick by intent, not
// by format string similarity:
//
//   - FileNameXxx  — used inside filesystem paths. Must not contain
//                    ':' or '/' which conflict with OS path separators.
//   - JSONXxx      — used as JSON / API fields. Must round-trip through
//                    time.RFC3339 parsing and preserve timezone.
//   - TableXxx     — used in TUI table columns. Human-readable, no
//                    timezone (display only).
//   - LogXxx       — used in composed log line prefixes. Sub-second
//                    precision where the source format provides it.

const (
	// FileNameDateTime is the filename-safe datetime used in tar / archive
	// names: container-myapp-20260810-153045.tar
	FileNameDateTime = "20060102-150405"

	// FileNameDate is the filename-safe date for daily rolling log files:
	// audit-2026-08-10.jsonl
	FileNameDate = "2006-01-02"
)

const (
	// JSONDateTimeUTC is the ISO 8601 datetime in UTC (no fractional seconds,
	// explicit 'Z' suffix) for JSON / API fields where the timestamp is always
	// UTC. Example: "2026-08-10T14:30:45Z"
	JSONDateTimeUTC = "2006-01-02T15:04:05Z"
)

const (
	// TableDateTime is the human-readable datetime used in fixed-width
	// TUI columns (19 chars, sized for column "fixed: 19" configs).
	// Example: "2026-08-10 14:30:45"
	TableDateTime = "2006-01-02 15:04:05"

	// TableShortDate is the eza-l "Mon DD HH:MM" style for narrow UI rows
	// where the year is implicit (current year). Example: "Aug 10 14:30"
	TableShortDate = "Jan 02 15:04"

	// TableShortDateFull is the short style with explicit year, for files
	// outside the current year. Example: "Aug 10 2025 14:30"
	TableShortDateFull = "Jan 02 2006 15:04"
)

const (
	// LogLineTime is the time-of-day with milliseconds used as the prefix
	// in composed log lines. Example: "14:30:45.123"
	LogLineTime = "15:04:05.000"
)

// ParseTimeOrZero parses an RFC3339 (with optional fractional seconds and
// explicit timezone offset) timestamp string. Empty input or parse failure
// returns the zero time.Time so callers can rely on IsZero() to detect
// missing timestamps without an extra error return.
func ParseTimeOrZero(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
