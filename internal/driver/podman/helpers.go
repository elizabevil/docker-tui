package podman

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// PodmanFilterQuery is a pass-through seam for filter maps destined for
// Podman's "filters" query parameter. The driver layer (e.g.
// containers_rest.go:ListContainers) does the actual JSON marshaling
// when setting the query param; pre-encoding here would double-encode
// and produce a malformed {"filters":["{\"label\":[...]}"]} that Podman
// rejects ("filters is an invalid filter [2]").
func PodmanFilterQuery(filters map[string][]string) (map[string][]string, error) {
	return filters, nil
}

// PodmanReportError extracts an error from a Podman prune report's raw JSON
// Error field. Returns nil when the field is null, empty, or an empty string.
func PodmanReportError(raw json.RawMessage) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) || bytes.Equal(trimmed, []byte(`""`)) {
		return nil
	}
	var message string
	if json.Unmarshal(trimmed, &message) == nil && message != "" {
		return fmt.Errorf("%s", message)
	}
	return fmt.Errorf("resource prune failed")
}
