package component

import (
	_ "embed"
	"strings"

	"github.com/elizabevil/docker-tui/internal/tui/keys"
)

//go:embed filter.jsonc
var filterDefaultData []byte

func init() {
	loader := ConfigLoader[styleConfig]{
		RawData:   filterDefaultData,
		Fallback:  styleConfig{},
		Normalize: nil,
	}
	cfg := loader.Load()
	RegisterComponentStyles("filter", cfg.Styles)
}

// FilterConfig holds debounce and display settings for the search/command filter.
type FilterConfig struct {
	DebounceMs  int    // debounce delay in ms (default 1000)
	MinChars    int    // minimum chars before filtering (default 1)
	Placeholder string // placeholder text for search mode
	CommandHint string // hint text for command mode
}

func (c *FilterConfig) normalize() {
	if c.DebounceMs <= 0 {
		c.DebounceMs = 1000
	}
	if c.MinChars <= 0 {
		c.MinChars = 1
	}
	if c.Placeholder == "" {
		c.Placeholder = "Search..."
	}
	if c.CommandHint == "" {
		c.CommandHint = "compose/images/..."
	}
}

// DefaultFilterConfig returns sensible defaults.
func DefaultFilterConfig() FilterConfig {
	loader := ConfigLoader[FilterConfig]{
		RawData: filterDefaultData,
		Fallback: FilterConfig{
			DebounceMs:  1000,
			MinChars:    1,
			Placeholder: "Search...",
			CommandHint: "compose/images/...",
		},
		Normalize: func(c *FilterConfig) {
			c.normalize()
		},
	}
	return loader.Load()
}

// AutocompleteHint returns the best matching autocomplete suggestion.
func AutocompleteHint(input string, _ int) string {
	if input == "" {
		return "compose/images/..."
	}
	lower := strings.ToLower(input)
	for _, cmd := range keys.Commands() {
		if strings.HasPrefix(cmd, lower) {
			return cmd
		}
	}
	return input
}

// AutocompleteSuffix returns the untyped suffix of the best matching command.
// For input "ima" it returns "ges" (completing to "images").
// For empty input it returns the default hint "compose/images/...".
// Returns "" when no completion is available (exact match or no match).
func AutocompleteSuffix(input string) string {
	if input == "" {
		return "compose/images/..."
	}
	lower := strings.ToLower(input)
	for _, cmd := range keys.Commands() {
		if strings.HasPrefix(cmd, lower) && len(cmd) > len(lower) {
			return cmd[len(lower):]
		}
	}
	return ""
}

// MatchFilter returns indices of rows that match the filter text (case-insensitive).
func MatchFilter(rows []string, filter string) []int {
	if filter == "" {
		// return all rows
		result := make([]int, len(rows))
		for i := range rows {
			result[i] = i
		}
		return result
	}
	lower := strings.ToLower(filter)
	var result []int
	for i, row := range rows {
		if strings.Contains(strings.ToLower(row), lower) {
			result = append(result, i)
		}
	}
	return result
}
