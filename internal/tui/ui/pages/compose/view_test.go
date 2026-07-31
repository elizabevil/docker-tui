package compose

import (
	"strings"
	"testing"
)

func TestBuildProjectDetailOmitsShortcutSection(t *testing.T) {
	got := BuildProjectDetail(composeProj{
		name:    "proj-a",
		total:   2,
		running: 1,
		svcs: map[string]composeSvc{
			"api": {count: 1, image: "nginx:latest", running: 1, total: 1},
		},
	})
	if strings.Contains(got, "快捷操作") {
		t.Fatalf("detail text still contains shortcut section: %q", got)
	}
	if !strings.Contains(got, "Project: proj-a") {
		t.Fatalf("project summary missing: %q", got)
	}
}
