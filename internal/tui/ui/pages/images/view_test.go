package images

import (
	"fmt"
	"strings"
	"testing"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"
	"github.com/elizabevil/docker-tui/internal/tui/ui/component"
)

func TestRenderListKeepsImageCursorInsideVisibleRows(t *testing.T) {
	images := make([]runtimeapi.ImageSummary, 30)
	for i := range images {
		images[i] = runtimeapi.ImageSummary{
			ID:       fmt.Sprintf("sha256:%064d", i),
			RepoTags: []string{fmt.Sprintf("example/image-%02d:latest", i)},
		}
	}
	model := state.NewImageListModel()
	model.Items = images
	model.Cursor = 20

	rendered := RenderList(model, state.NewContainerListModel(), 120, 12, nil, false)
	rowHeight := component.CalcTableRowHeight(12, true)
	wantOffset := model.Cursor - rowHeight + 1
	if model.ViewOffset != wantOffset {
		t.Fatalf("ViewOffset = %d, want %d (row height %d)", model.ViewOffset, wantOffset, rowHeight)
	}
	if model.Cursor < model.ViewOffset || model.Cursor >= model.ViewOffset+rowHeight {
		t.Fatalf("cursor %d outside viewport [%d,%d)", model.Cursor, model.ViewOffset, model.ViewOffset+rowHeight)
	}
	if plain := component.StripANSI(rendered); !strings.Contains(plain, "image-20") {
		t.Fatalf("selected image is outside rendered table: %q", plain)
	}
	if lines := strings.Count(rendered, "\n") + 1; lines > 12 {
		t.Fatalf("rendered table uses %d lines, panel only has 12: %q", lines, component.StripANSI(rendered))
	}
}

func TestRenderListScrollsBackUpToImageCursor(t *testing.T) {
	model := state.NewImageListModel()
	model.Items = make([]runtimeapi.ImageSummary, 30)
	for i := range model.Items {
		model.Items[i] = runtimeapi.ImageSummary{
			ID:       fmt.Sprintf("sha256:%064d", i),
			RepoTags: []string{fmt.Sprintf("example/image-%02d:latest", i)},
		}
	}
	model.Cursor = 4
	model.ViewOffset = 18

	RenderList(model, state.NewContainerListModel(), 120, 12, nil, false)
	if model.ViewOffset != model.Cursor {
		t.Fatalf("upward viewport did not follow cursor: offset=%d cursor=%d", model.ViewOffset, model.Cursor)
	}
}

func TestRenderListUsesStableExcelLikeColumnTracks(t *testing.T) {
	model := state.NewImageListModel()
	model.Items = []runtimeapi.ImageSummary{{
		ID:       "6b1b147de1234567890",
		RepoTags: []string{"h536b8/canway_d/blueking/cmdb_adminserver:v3.14.8-alpha3-cw.1"},
		Registry: "docker-bkrepo.example.com",
		Arch:     "amd64",
		Created:  1785395907,
		Size:     2696 * 1000 * 100,
	}}

	const pageWidth = 200
	rendered := component.StripANSI(RenderList(model, state.NewContainerListModel(), pageWidth, 12, nil, false))
	for _, line := range strings.Split(rendered, "\n") {
		if !strings.Contains(line, "cmdb_adminserver") || !strings.Contains(line, "6b1b147de") {
			continue
		}
		if got := component.VisibleLen(line); got != pageWidth {
			t.Fatalf("image row width = %d, want %d: %q", got, pageWidth, line)
		}
		for _, value := range []string{"docker-bkrepo.example.com", "v3.14.8-alpha3-cw.1", "6b1b147de1234567890"} {
			if !strings.Contains(line, value) {
				t.Fatalf("wide viewport still truncates %q: %q", value, line)
			}
		}
		if trailing := component.VisibleLen(line) - component.VisibleLen(strings.TrimRight(line, " ")); trailing > 1 {
			t.Fatalf("wide table leaves %d cells unused at the right edge: %q", trailing, line)
		}
		return
	}
	t.Fatalf("image row was not rendered: %q", rendered)
}
