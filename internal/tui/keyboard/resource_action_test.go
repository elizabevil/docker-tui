package keyboard

import "testing"

func TestValidResourceName(t *testing.T) {
	for _, name := range []string{"cache", "app-data_1", "team.network"} {
		if !validResourceName(name) {
			t.Errorf("validResourceName(%q) = false", name)
		}
	}
	for _, name := range []string{"", "with space", "team/cache"} {
		if validResourceName(name) {
			t.Errorf("validResourceName(%q) = true", name)
		}
	}
}
