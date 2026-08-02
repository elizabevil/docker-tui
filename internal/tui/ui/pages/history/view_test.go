package history

import (
	"fmt"
	"strings"
	"testing"

	"github.com/elizabevil/docker-tui/internal/data/config"
	dockerclient "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
)

func sampleState() *state.AppModel {
	m := state.NewAppModel(config.DefaultConfig(), nil, "test")
	m.History.Open("sha256:abc", "nginx:latest")
	return m
}

func sampleLayers() []dockerclient.ImageHistoryLayer {
	return []dockerclient.ImageHistoryLayer{
		{ID: "layer-a", Created: 1700000000, CreatedBy: "RUN apt-get install vim", Size: 1024, Comment: ""},
		{ID: "layer-b", Created: 1700001000, CreatedBy: "COPY . /app", Size: 2048, Comment: "build"},
		{ID: "layer-c", Created: 1700002000, CreatedBy: `CMD ["/bin/sh"]`, Size: 0, Comment: ""},
	}
}

func TestRenderViewNoImageID(t *testing.T) {
	m := state.NewAppModel(config.DefaultConfig(), nil, "test")
	got := RenderView(m, 30, 120)
	if !strings.Contains(got, "Select an image") {
		t.Errorf("missing 'Select an image' message in: %q", got)
	}
}

func TestRenderViewLoading(t *testing.T) {
	m := sampleState()
	// History.Loading is true after Open()
	got := RenderView(m, 30, 120)
	if !strings.Contains(got, "Loading history") {
		t.Errorf("missing loading message in: %q", got)
	}
}

func TestRenderViewError(t *testing.T) {
	m := sampleState()
	m.History.Layers = nil
	m.History.Loading = false
	m.History.Error = "socket hung up"
	got := RenderView(m, 30, 120)
	if !strings.Contains(got, "socket hung up") {
		t.Errorf("missing error message in: %q", got)
	}
}

func TestRenderViewManifest(t *testing.T) {
	m := sampleState()
	m.History.Loading = false
	m.History.Layers = nil
	m.History.Source = dockerclient.ImageHistoryManifest
	got := RenderView(m, 30, 120)
	if !strings.Contains(got, "manifest list") {
		t.Errorf("missing manifest notice in: %q", got)
	}
}

func TestFilterLayersEmptyFilter(t *testing.T) {
	got := FilterLayers(sampleLayers(), "")
	if len(got) != 3 {
		t.Errorf("empty filter = %d, want 3", len(got))
	}
}

func TestFilterLayersCreatedBy(t *testing.T) {
	got := FilterLayers(sampleLayers(), "apt")
	if len(got) != 1 {
		t.Fatalf("apt filter = %d layers, want 1", len(got))
	}
	if got[0].ID != "layer-a" {
		t.Errorf("matched = %s, want layer-a", got[0].ID)
	}
}

func TestFilterLayersComment(t *testing.T) {
	got := FilterLayers(sampleLayers(), "build")
	if len(got) != 1 {
		t.Fatalf("build filter = %d layers, want 1", len(got))
	}
	if got[0].ID != "layer-b" {
		t.Errorf("matched = %s, want layer-b", got[0].ID)
	}
}

func TestFilterLayersCaseInsensitive(t *testing.T) {
	got := FilterLayers(sampleLayers(), "APT")
	if len(got) != 1 {
		t.Errorf("uppercase filter = %d, want 1", len(got))
	}
}

func TestLayerRowFormats(t *testing.T) {
	row := layerRow(sampleLayers()[0])
	if len(row) != 5 {
		t.Fatalf("len = %d, want 5", len(row))
	}
	if row[0] != "layer-a" {
		t.Errorf("id = %q", row[0])
	}
	if !strings.HasPrefix(row[3], "RUN apt-get") {
		t.Errorf("created_by = %q, want 'RUN apt-get' prefix", row[3])
	}
	// Selected highlight is rendered by RenderTable from TableData.Selected,
	// not by layerRow, so row[3] is the raw CreatedBy text.
	if strings.HasPrefix(row[3], "▶ ") {
		t.Errorf("created_by should not have manual arrow prefix; got %q", row[3])
	}
}

func TestRenderViewTableOnlyVisibleRows(t *testing.T) {
	m := sampleState()
	m.History.Loading = false
	m.History.Layers = sampleLayers()
	m.History.Cursor = 0
	m.History.ViewOffset = 0
	// BodyHeight = 30 - 4 = 26 visible rows
	got := RenderView(m, 30, 120)
	// We render at most 26 rows; with 3 layers we should see all 3
	// plus footer mention of 1-3/3.
	if !strings.Contains(got, "1-3/3") {
		t.Errorf("footer should show 1-3/3, got: %q", got[len(got)-100:])
	}
}

func TestRenderViewClampsViewOffset(t *testing.T) {
	m := sampleState()
	m.History.Loading = false
	m.History.Layers = sampleLayers()
	// ViewOffset past end should clamp
	m.History.ViewOffset = 100
	m.History.Cursor = 0
	_ = RenderView(m, 30, 120) // should not panic
}

func BenchmarkRenderCachedHistory(b *testing.B) {
	m := sampleState()
	m.History.Loading = false
	m.History.Layers = make([]dockerclient.ImageHistoryLayer, 100)
	for i := range m.History.Layers {
		m.History.Layers[i] = dockerclient.ImageHistoryLayer{ID: fmt.Sprintf("sha256:%064d", i), Created: 1700000000 + int64(i), CreatedBy: "RUN apt-get install package", Size: int64(i * 1024)}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.History.Cursor = i % len(m.History.Layers)
		m.History.EnsureVisible(20)
		_ = RenderView(m, 24, 160)
	}
}
