package podman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
)

// PodmanFilterQuery encodes a filter map as a JSON query parameter for Podman
// REST endpoints that accept a "filters" query string.
func PodmanFilterQuery(filters map[string][]string) (url.Values, error) {
	query := make(url.Values)
	if len(filters) == 0 {
		return query, nil
	}
	encoded, err := json.Marshal(filters)
	if err != nil {
		return nil, fmt.Errorf("encode Podman prune filters: %w", err)
	}
	log.Printf("podman filter: %s", string(encoded))
	query.Set("filters", string(encoded))
	return query, nil
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
