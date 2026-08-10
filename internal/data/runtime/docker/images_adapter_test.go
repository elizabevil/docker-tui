package docker

import (
	"encoding/json"
	"testing"

	"github.com/docker/docker/api/types/image"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

func TestListImagesSetsRegistryFromRepoTag(t *testing.T) {
	raw := image.Summary{
		ID:       "sha256:img",
		RepoTags: []string{"h536b8/canway_d/blueking/cmdb_adminserver:v3.14.8-alpha3-cw.1"},
	}
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	var decoded image.Summary
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	s := runtimeapi.ImageSummary{
		ID:       decoded.ID,
		RepoTags: decoded.RepoTags,
	}
	s.Registry, _, _ = SplitImageRef(s.RepoTags)
	if s.Registry == "" {
		t.Fatalf("docker adapter must populate Registry, got %q", s.Registry)
	}
	if s.Registry != "docker.io" && s.Registry != "h536b8" {
		t.Fatalf("docker adapter produced unexpected registry %q", s.Registry)
	}
}
