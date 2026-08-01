package utils

import "testing"

func TestShortID(t *testing.T) {
	for input, want := range map[string]string{
		"246a10cc414f280d":        "246a10cc414f",
		"sha256:246a10cc414f280d": "246a10cc414f",
		"short":                   "short",
	} {
		if got := ShortID(input); got != want {
			t.Errorf("ShortID(%q) = %q, want %q", input, got, want)
		}
	}
}
