package component

import (
	"testing"
)

func TestMatchFilter(t *testing.T) {
	rows := []string{"nginx-web", "redis-cache", "postgres-db", "nginx-proxy"}

	// Match "nginx" → 2 results
	got := MatchFilter(rows, "nginx")
	if len(got) != 2 || got[0] != 0 || got[1] != 3 {
		t.Errorf("MatchFilter(nginx) = %v, want [0 3]", got)
	}

	// Match "redis" → 1 result
	got = MatchFilter(rows, "redis")
	if len(got) != 1 || got[0] != 1 {
		t.Errorf("MatchFilter(redis) = %v, want [1]", got)
	}

	// Match "" → all rows
	got = MatchFilter(rows, "")
	if len(got) != 4 {
		t.Errorf("MatchFilter('') = %v, want 4 results", got)
	}

	// Match "nonexistent" → 0 results
	got = MatchFilter(rows, "nonexistent")
	if len(got) != 0 {
		t.Errorf("MatchFilter(nonexistent) = %v, want empty", got)
	}

	// Case insensitive
	got = MatchFilter(rows, "NGINX")
	if len(got) != 2 {
		t.Errorf("MatchFilter(NGINX) = %v, want 2 results", got)
	}
}

func TestAutocompleteHint(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "compose/images/..."},
		{"com", "compose"},
		{"ima", "images"},
		{"con", "containers"},
		{"vol", "volumes"},
		{"net", "networks"},
		{"log", "logs"},
		{"hel", "help"},
		{"xyz", "xyz"},
	}
	for _, tt := range tests {
		got := AutocompleteHint(tt.input, 0)
		if got != tt.want {
			t.Errorf("AutocompleteHint(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
