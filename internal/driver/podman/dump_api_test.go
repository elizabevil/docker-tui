package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpPodmanAPIResponses calls the live Podman REST API and writes raw
// JSON responses to testdata/{version}/{name}.json for offline inspection.
// Run with:
//
//	go test -v -run TestDumpPodmanAPIResponses ./internal/data/runtime/podman/
func TestDumpPodmanAPIResponses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live Podman API dump")
	}

	uid := os.Getuid()
	socketPaths := []string{
		fmt.Sprintf("/run/user/%d/podman/podman.sock", uid),
		"/run/user/1000/podman/podman.sock",
		"/var/run/podman/podman.sock",
	}
	var socketPath string
	for _, p := range socketPaths {
		if _, err := os.Stat(p); err == nil {
			socketPath = p
			break
		}
	}
	if socketPath == "" {
		t.Skip("No Podman socket found")
	}
	t.Logf("Using socket: %s", socketPath)
	uri := fmt.Sprintf("unix://%s", socketPath)
	client, err := NewRESTClient(RESTConfig{Endpoint: uri})
	if err != nil {
		t.Fatalf("NewRESTClient: %v", err)
	}

	ctx := context.Background()
	version, err := client.APIVersion(ctx)
	if err != nil {
		t.Fatalf("APIVersion: %v", err)
	}
	t.Logf("Podman API version: %s", version)

	outDir := filepath.Join("testdata", version)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	endpoints := []struct {
		name string
		path string
	}{
		{"libpod_version", PathSystemVersion},
		{"info", "/info"},
		{"libpod_info", "/libpod/info"},
		{"containers_json", "/libpod/containers/json"},
		{"containers_json_all", "/libpod/containers/json?all=true"},
		{"volumes_json", "/libpod/volumes/json"},
		{"networks_json", "/libpod/networks/json"},
		{"images_json", "/libpod/images/json"},
		{"images_json_all", "/libpod/images/json?all=true"},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			var raw json.RawMessage
			err := client.Get(ctx, ep.path, nil, &raw)
			if err != nil {
				t.Logf("GET %s error: %v", ep.path, err)
				writeErrorFile(t, outDir, ep.name, ep.path, err)
				return
			}
			pretty, err := json.MarshalIndent(json.RawMessage(raw), "", "  ")
			if err != nil {
				pretty = raw
			}

			filePath := filepath.Join(outDir, ep.name+".json")
			if err := os.WriteFile(filePath, pretty, 0o644); err != nil {
				t.Fatalf("write %s: %v", filePath, err)
			}
			t.Logf("wrote %s (%d bytes)", filePath, len(pretty))
		})
	}
}

func writeErrorFile(t *testing.T, dir, name, apiPath string, err error) {
	t.Helper()
	filePath := filepath.Join(dir, name+"_error.txt")
	content := fmt.Sprintf("path: %s\nerror: %v\n", apiPath, err)
	_ = os.WriteFile(filePath, []byte(content), 0o644)
}
