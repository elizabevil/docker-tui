package state

import (
	"testing"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
)

func makeTestContainers() []dockerclient.ContainerSummary {
	return []dockerclient.ContainerSummary{
		{ID: "abc123def456", Name: "web-app", Image: "nginx:latest", State: "running", Status: "Up 2 hours"},
		{ID: "def456abc789", Name: "redis-cache", Image: "redis:7", State: "running", Status: "Up 1 hour"},
		{ID: "ghi789jkl012", Name: "postgres-db", Image: "postgres:16", State: "running", Status: "Up 3 hours"},
		{ID: "jkl012mno345", Name: "stopped-app", Image: "alpine:3", State: "exited", Status: "Exited (0) 5 days ago"},
		{ID: "mno345pqr678", Name: "worker-1", Image: "python:3.12", State: "running", Status: "Up 30 minutes"},
	}
}

func TestNewContainerListModel(t *testing.T) {
	m := NewContainerListModel()
	if m == nil {
		t.Fatal("NewContainerListModel returned nil")
	}
	if m.Len() != 0 {
		t.Errorf("expected 0 items, got %d", m.Len())
	}
	if m.Cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.Cursor)
	}
}

func TestContainerListModelSelected(t *testing.T) {
	m := NewContainerListModel()
	if m.Selected() != nil {
		t.Error("expected nil Selected on empty model")
	}
	m.Items = makeTestContainers()
	m.Cursor = 0
	sel := m.Selected()
	if sel == nil {
		t.Fatal("expected non-nil Selected at cursor 0")
	}
	if sel.Name != "web-app" {
		t.Errorf("expected 'web-app', got %q", sel.Name)
	}
	m.Cursor = 10
	if m.Selected() != nil {
		t.Error("expected nil Selected with out-of-bounds cursor")
	}
}

func TestContainerListModelLen(t *testing.T) {
	m := NewContainerListModel()
	if m.Len() != 0 {
		t.Error("expected Len 0 for empty model")
	}
	m.Items = makeTestContainers()
	if m.Len() != 5 {
		t.Errorf("expected Len 5, got %d", m.Len())
	}
}

func TestContainerListModelFilter(t *testing.T) {
	m := NewContainerListModel()
	m.Items = makeTestContainers()

	all := m.FilteredItems()
	if len(all) != 5 {
		t.Errorf("expected 5 unfiltered items, got %d", len(all))
	}

	m.Filter = "nginx"
	nginxOnly := m.FilteredItems()
	if len(nginxOnly) != 1 {
		t.Errorf("expected 1 nginx match, got %d", len(nginxOnly))
	}
	if nginxOnly[0].Name != "web-app" {
		t.Errorf("expected 'web-app', got %q", nginxOnly[0].Name)
	}

	m.Filter = "postgres"
	pgOnly := m.FilteredItems()
	if len(pgOnly) != 1 {
		t.Errorf("expected 1 postgres match, got %d", len(pgOnly))
	}

	m.Filter = "hours"
	hoursOnly := m.FilteredItems()
	if len(hoursOnly) != 2 {
		t.Errorf("expected 2 status matches for 'hours', got %d", len(hoursOnly))
	}
}

func TestContainerListModelFilterEmpty(t *testing.T) {
	m := NewContainerListModel()
	m.Items = makeTestContainers()
	m.Filter = "nonexistent"
	filtered := m.FilteredItems()
	if len(filtered) != 0 {
		t.Errorf("expected 0 matches for nonexistent filter, got %d", len(filtered))
	}
	m.Filter = ""
	filtered = m.FilteredItems()
	if len(filtered) != 5 {
		t.Errorf("expected all items for empty filter, got %d", len(filtered))
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s, substr string
		expected  bool
	}{
		{"hello", "ll", true},
		{"hello", "he", true},
		{"hello", "lo", true},
		{"hello", "hello", true},
		{"hello", "world", false},
		{"hello", "", true},
		{"", "x", false},
		{"short", "longer", false},
	}
	for _, tc := range tests {
		result := contains(tc.s, tc.substr)
		if result != tc.expected {
			t.Errorf("contains(%q, %q) = %v, want %v", tc.s, tc.substr, result, tc.expected)
		}
	}
}

func TestContainerListModelViewOffset(t *testing.T) {
	m := NewContainerListModel()
	if m.ViewOffset != 0 {
		t.Errorf("expected initial ViewOffset 0, got %d", m.ViewOffset)
	}
}

func TestContainerListModelStatsMap(t *testing.T) {
	m := NewContainerListModel()
	if m.Stats == nil {
		t.Fatal("expected Stats map to be initialized")
	}
	if len(m.Stats) != 0 {
		t.Errorf("expected empty Stats map, got %d entries", len(m.Stats))
	}
	m.Stats["abc123"] = ContainerStats{CPU: 10.5, MemPerc: 25.0}
	if v, ok := m.Stats["abc123"]; !ok || v.CPU != 10.5 {
		t.Error("failed to store/retrieve container stats")
	}
}
