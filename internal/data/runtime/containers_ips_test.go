package runtime

import (
	"sort"
	"testing"
)

func TestOrderedIPsBridgeFirst(t *testing.T) {
	s := &ContainerSummary{
		IPs: []string{"10.0.0.2", "172.17.0.2"},
		Networks: map[string]string{
			"bridge": "172.17.0.2",
			"myNet":  "10.0.0.2",
		},
	}
	got := s.OrderedIPs()
	want := []string{"172.17.0.2", "10.0.0.2"}
	if !equalSlice(got, want) {
		t.Fatalf("OrderedIPs = %v, want %v", got, want)
	}
}

func TestOrderedIPsDropsEmpty(t *testing.T) {
	s := &ContainerSummary{
		IPs:      []string{"", "10.0.0.2", ""},
		Networks: map[string]string{},
	}
	got := s.OrderedIPs()
	want := []string{"10.0.0.2"}
	if !equalSlice(got, want) {
		t.Fatalf("OrderedIPs = %v, want %v", got, want)
	}
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

var _ = sort.Strings
