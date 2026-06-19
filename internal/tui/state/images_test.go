package state

import (
	"testing"

	dockerclient "github.com/elizabevil/docker-tui/internal/data/docker"
)

func makeTestImages() []dockerclient.ImageSummary {
	return []dockerclient.ImageSummary{
		{ID: "sha256:aaa111", RepoTags: []string{"nginx:latest"}, Created: 1700000000, Size: 187000000},
		{ID: "sha256:bbb222", RepoTags: []string{"redis:7-alpine", "redis:latest"}, Created: 1699000000, Size: 32000000},
		{ID: "sha256:ccc333", RepoTags: []string{"postgres:16"}, Created: 1700100000, Size: 420000000},
		{ID: "sha256:ddd444", RepoTags: []string{"alpine:3.19"}, Created: 1698000000, Size: 7200000},
		{ID: "sha256:eee555", RepoTags: []string{}, Created: 1697000000, Size: 15000000},
	}
}

func TestNewImageListModel(t *testing.T) {
	m := NewImageListModel()
	if m == nil {
		t.Fatal("NewImageListModel returned nil")
	}
	if m.Len() != 0 {
		t.Errorf("expected 0 items, got %d", m.Len())
	}
	if !m.SortAsc {
		t.Error("expected SortAsc to be true by default")
	}
	if m.SortBy != ImageSortByRepo {
		t.Errorf("expected SortBy to be ImageSortByRepo, got %d", m.SortBy)
	}
}

func TestImageListModelFilteredItems(t *testing.T) {
	m := NewImageListModel()
	m.Items = makeTestImages()

	all := m.FilteredItems()
	if len(all) != 5 {
		t.Errorf("expected 5 unfiltered, got %d", len(all))
	}

	m.Filter = "nginx"
	nginxOnly := m.FilteredItems()
	if len(nginxOnly) != 1 {
		t.Errorf("expected 1 nginx match, got %d", len(nginxOnly))
	}

	m.Filter = "redis"
	redisOnly := m.FilteredItems()
	if len(redisOnly) != 1 {
		t.Errorf("expected 1 redis match, got %d", len(redisOnly))
	}

	m.Filter = "sha256"
	allMatch := m.FilteredItems()
	if len(allMatch) != 5 {
		t.Errorf("expected 5 ID matches, got %d", len(allMatch))
	}
}

func TestImageListModelSortByRepo(t *testing.T) {
	m := NewImageListModel()
	m.Items = makeTestImages()
	m.SortBy = ImageSortByRepo
	m.SortAsc = true

	sorted := m.SortedItems()
	if len(sorted) != 5 {
		t.Fatalf("expected 5 sorted items, got %d", len(sorted))
	}
	// <none>:<none> should come first
	firstTag := repoTag(sorted[0])
	if firstTag != "<none>:<none>" {
		t.Errorf("expected <none>:<none> first, got %q", firstTag)
	}
	// alpine should come before nginx
	if repoTag(sorted[2]) != "nginx:latest" {
		t.Errorf("expected nginx:latest at index 2, got %q", repoTag(sorted[2]))
	}
}

func TestImageListModelSortByID(t *testing.T) {
	m := NewImageListModel()
	m.Items = makeTestImages()
	m.SortBy = ImageSortByID
	m.SortAsc = true

	sorted := m.SortedItems()
	if len(sorted) != 5 {
		t.Fatalf("expected 5 sorted items, got %d", len(sorted))
	}
	if sorted[0].ID != "sha256:aaa111" {
		t.Errorf("expected aaa111 first, got %s", sorted[0].ID)
	}
	if sorted[4].ID != "sha256:eee555" {
		t.Errorf("expected eee555 last, got %s", sorted[4].ID)
	}
}

func TestImageListModelSortBySize(t *testing.T) {
	m := NewImageListModel()
	m.Items = makeTestImages()
	m.SortBy = ImageSortBySize
	m.SortAsc = true

	sorted := m.SortedItems()
	if sorted[0].Size != 7200000 {
		t.Errorf("expected smallest 7.2M first, got %d", sorted[0].Size)
	}
	if sorted[4].Size != 420000000 {
		t.Errorf("expected largest 420M last, got %d", sorted[4].Size)
	}
}

func TestImageListModelSortByCreated(t *testing.T) {
	m := NewImageListModel()
	m.Items = makeTestImages()
	m.SortBy = ImageSortByCreated
	m.SortAsc = true

	sorted := m.SortedItems()
	if sorted[0].Created != 1697000000 {
		t.Errorf("expected oldest first (%d), got %d", 1697000000, sorted[0].Created)
	}
}

func TestImageListModelSortDescending(t *testing.T) {
	m := NewImageListModel()
	m.Items = makeTestImages()
	m.SortBy = ImageSortBySize
	m.SortAsc = false

	sorted := m.SortedItems()
	if sorted[0].Size != 420000000 {
		t.Errorf("expected largest first, got %d", sorted[0].Size)
	}
}

func TestRepoTag(t *testing.T) {
	img := dockerclient.ImageSummary{RepoTags: []string{"nginx:latest", "nginx:stable"}}
	if repoTag(img) != "nginx:latest" {
		t.Errorf("expected 'nginx:latest', got %q", repoTag(img))
	}
	img.RepoTags = nil
	if repoTag(img) != "<none>:<none>" {
		t.Errorf("expected '<none>:<none>', got %q", repoTag(img))
	}
	img.RepoTags = []string{}
	if repoTag(img) != "<none>:<none>" {
		t.Errorf("expected '<none>:<none>', got %q", repoTag(img))
	}
}

func TestImageListModelSelected(t *testing.T) {
	m := NewImageListModel()
	if m.Selected() != nil {
		t.Error("expected nil Selected on empty")
	}
	m.Items = makeTestImages()
	m.Cursor = 0
	// Default sort by repo ascending: <none> → alpine → nginx → postgres → redis
	if m.Selected().ID != "sha256:eee555" {
		t.Errorf("expected eee555 (<none>) at cursor 0, got %s", m.Selected().ID)
	}
	m.Cursor = 99
	if m.Selected() != nil {
		t.Error("expected nil with out-of-bounds cursor")
	}
	m.Cursor = -1
	if m.Selected() != nil {
		t.Error("expected nil with negative cursor")
	}
}
