package docker

import "time"

// sentinels for the Docker engine wire format that have no Go-native
// representation in this package.
const (
	// zeroTimeRFC3339 is the RFC3339 rendering of Go's zero time.Time
	// (time.Time{}.Format(time.RFC3339)). The Docker daemon emits this
	// exact string as a "never finished / never started" sentinel for
	// container state fields, so we compare against the literal rather
	// than parsing back to time.Time on every row.
	zeroTimeRFC3339 = "0001-01-01T00:00:00Z"
)

// IsZeroTime reports whether s is the Docker "no value" sentinel for
// time-typed fields (zero time.Time formatted as RFC3339, or empty).
func IsZeroTime(s string) bool {
	return s == "" || s == zeroTimeRFC3339
}

// init guards time.RFC3339 against drift: if the standard-library
// constant changes formatting in a future release, the compile-time
// reference below catches it. Storing the literal above is intentional
// (we want the exact string Docker sends, not whatever Go formats).
var _ = time.RFC3339
