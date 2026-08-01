package detail

import (
	"fmt"
	"strings"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func TestRenderViewClampsImageDetailOverscroll(t *testing.T) {
	environment := make([]string, 30)
	for i := range environment {
		environment[i] = fmt.Sprintf("KEY_%02d=value", i)
	}
	app := &state.AppModel{}
	app.Detail.OpenImage("image", "Image Detail", &runtimeapi.ImageDetail{
		ID: "image",
		Runtime: runtimeapi.ImageRuntimeConfig{
			Environment: environment,
		},
	})
	app.Detail.DetailOffset = 1000

	RenderView(app, 8, 80)
	bottom := app.Detail.DetailOffset
	if bottom <= 0 || bottom >= 1000 {
		t.Fatalf("render did not clamp overscroll: %d", bottom)
	}

	app.Detail.Scroll(-1)
	RenderView(app, 8, 80)
	if app.Detail.DetailOffset != bottom-1 {
		t.Fatalf("up scroll remained stuck: got %d, want %d", app.Detail.DetailOffset, bottom-1)
	}
}

func TestRenderSectionsFooterUsesActualVisibleCount(t *testing.T) {
	app := &state.AppModel{}
	rendered := renderDocument(app, flattenSections([]detailSection{{
		Title: "Basic",
		Lines: []string{"ID: image"},
	}}), 8, 80)

	if plain := component.StripANSI(rendered); !strings.Contains(plain, "1-2/2") {
		t.Fatalf("footer does not use actual visible count: %q", plain)
	}
}

func TestRenderImageRegistryWrapsWithoutTruncation(t *testing.T) {
	const registry = "docker-bkrepo.internal.example.com:5000"
	app := &state.AppModel{}
	app.Detail.OpenImage("image", "Image Detail", &runtimeapi.ImageDetail{ID: "image", Registry: registry})

	plain := component.StripANSI(RenderView(app, 12, 28))
	joined := strings.ReplaceAll(strings.ReplaceAll(plain, "\n", ""), " ", "")
	if !strings.Contains(joined, "Registry:"+registry) {
		t.Fatalf("registry was truncated while wrapping: %q", plain)
	}
	if strings.Contains(plain, "docker-bkrepo...") {
		t.Fatalf("registry contains truncation marker: %q", plain)
	}
}

func TestRenderViewReusesDocumentWhileScrolling(t *testing.T) {
	environment := make([]string, 50)
	for i := range environment {
		environment[i] = fmt.Sprintf("KEY_%02d=value", i)
	}
	app := &state.AppModel{}
	app.Detail.OpenImage("image", "Image Detail", &runtimeapi.ImageDetail{
		ID: "image",
		Runtime: runtimeapi.ImageRuntimeConfig{
			Environment: environment,
		},
	})

	RenderView(app, 8, 80)
	document := app.Detail.Documents[state.DetailSourceSection]
	if len(document.Lines) == 0 {
		t.Fatal("detail document was not cached")
	}
	first := &document.Lines[0]
	app.Detail.Scroll(1)
	RenderView(app, 8, 80)
	document = app.Detail.Documents[state.DetailSourceSection]
	if first != &document.Lines[0] {
		t.Fatal("scroll rebuilt the detail document")
	}
}

func TestSourceDocumentIsCachedAndNeverBlank(t *testing.T) {
	app := &state.AppModel{}
	app.Detail.OpenImage("image", "Image Detail", &runtimeapi.ImageDetail{ID: "image", Name: "demo"})
	app.Detail.CycleSource()

	RenderView(app, 8, 80)
	yamlDocument := app.Detail.Documents[state.DetailSourceYAML]
	var sourceLines []string
	for _, line := range yamlDocument.Lines {
		sourceLines = append(sourceLines, line.Left)
	}
	if source := strings.Join(sourceLines, "\n"); !strings.Contains(source, "ID: image") {
		t.Fatalf("yaml source missing image data: %q", source)
	}
	first := &yamlDocument.Lines[0]
	app.Detail.Scroll(1)
	RenderView(app, 8, 80)
	yamlDocument = app.Detail.Documents[state.DetailSourceYAML]
	if first != &yamlDocument.Lines[0] {
		t.Fatal("scroll rebuilt the source document")
	}

	app.Detail.CycleSource()
	RenderView(app, 8, 80)
	app.Detail.CycleSource()
	app.Detail.CycleSource()
	RenderView(app, 8, 80)
	yamlDocument = app.Detail.Documents[state.DetailSourceYAML]
	if first != &yamlDocument.Lines[0] {
		t.Fatal("source cycle rebuilt the cached YAML document")
	}

	app.Detail.DetailSourceType = state.DetailSourceJSON
	app.Detail.DetailRawJSON = nil
	app.Detail.Documents = nil
	rendered := RenderView(app, 8, 80)
	if plain := component.StripANSI(rendered); !strings.Contains(plain, "unavailable") {
		t.Fatalf("empty source did not render an explicit placeholder: %q", plain)
	}
}

func BenchmarkRenderCachedContainerDetail(b *testing.B) {
	environment := make([]string, 500)
	labels := make(map[string]string, 500)
	for i := range environment {
		environment[i] = fmt.Sprintf("KEY_%03d=value", i)
		labels[fmt.Sprintf("label.%03d", i)] = "value"
	}
	app := &state.AppModel{}
	app.Detail.Open("Container Detail", "")
	app.Detail.SetContainerDetail(&runtimeapi.ContainerDetail{
		ID: "container",
		Config: runtimeapi.ContainerConfig{
			Environment: environment,
			Labels:      labels,
		},
	})
	RenderView(app, 30, 80)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Detail.DetailOffset = i % 900
		RenderView(app, 30, 80)
	}
}
