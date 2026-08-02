package podman

import (
	"testing"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

func TestMapPodmanHistoryEmpty(t *testing.T) {
	if got := mapPodmanHistory(nil); len(got) != 0 {
		t.Errorf("nil = %d, want 0", len(got))
	}
	if got := mapPodmanHistory([]dto.LayerHistoryEntry{}); len(got) != 0 {
		t.Errorf("empty = %d, want 0", len(got))
	}
}

func TestMapPodmanHistoryFieldPassthrough(t *testing.T) {
	raw := []dto.LayerHistoryEntry{
		{ID: "abc", Created: 1700000000, CreatedBy: "RUN apt-get install vim", Size: 12345, Comment: "first layer"},
		{ID: "def", Created: 1700001000, CreatedBy: `CMD ["/bin/sh"]`, Size: 0},
	}
	got := mapPodmanHistory(raw)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != "abc" || got[0].Created != 1700000000 || got[0].Size != 12345 || got[0].Comment != "first layer" {
		t.Errorf("got[0] = %+v", got[0])
	}
	if got[0].CreatedBy != "RUN apt-get install vim" {
		t.Errorf("got[0].CreatedBy = %q", got[0].CreatedBy)
	}
	if got[1].ID != "def" || got[1].Size != 0 {
		t.Errorf("got[1] = %+v", got[1])
	}
}
