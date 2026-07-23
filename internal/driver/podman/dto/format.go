package dto

import "time"

// FormatPodmanTime formats a Podman time value as RFC3339, returning "" for
// the zero time. Available in both CGO and non-CGO builds.
func FormatPodmanTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}
