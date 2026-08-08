package podman

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/elizabevil/docker-tui/internal/driver/podman/dto"
)

// TestAdvancedContainerPaths verifies the TASK-019 URL helpers return the
// expected Libpod REST paths so a typo doesn't silently route an advanced
// operation to the wrong endpoint.
func TestAdvancedContainerPaths(t *testing.T) {
	id := "abc123"
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"update", ContainerUpdatePath(id), "/libpod/containers/abc123/update"},
		{"wait", ContainerWaitPath(id), "/libpod/containers/abc123/wait"},
		{"changes", ContainerChangesPath(id), "/libpod/containers/abc123/changes"},
		{"export", ContainerExportPath(id), "/libpod/containers/abc123/export"},
		{"commit", ContainerCommitPath(), "/libpod/commit"},
		{"archive", ContainerArchivePath(id), "/libpod/containers/abc123/archive"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("path = %q, want %q", tc.got, tc.want)
			}
		})
	}
}

func TestContainerCommitUsesCollectionEndpointAndQuery(t *testing.T) {
	client, err := NewRESTClient(RESTConfig{
		Endpoint:   "http://podman.test",
		APIVersion: "5.4.2",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodPost || request.URL.Path != "/v5.4.2/libpod/commit" {
				t.Fatalf("request = %s %s", request.Method, request.URL.Path)
			}
			query := request.URL.Query()
			for key, want := range map[string]string{
				"container": "abc123", "repo": "localhost/demo", "tag": "v1",
				"author": "dtui", "comment": "snapshot", "pause": "true",
			} {
				if got := query.Get(key); got != want {
					t.Fatalf("query %s = %q, want %q", key, got, want)
				}
			}
			return response(http.StatusCreated, `{"Id":"sha256:new"}`), nil
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ContainerCommit(context.Background(), "abc123", dto.ContainerCommitOptions{
		Repository: "localhost/demo", Tag: "v1", Author: "dtui", Comment: "snapshot", Pause: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "sha256:new" {
		t.Fatalf("commit ID = %q", result.ID)
	}
}

// TestContainerUpdateSwaggerCompliance verifies the update request is
// shaped per ContainerUpdateLibpod: query carries only restartPolicy /
// restartRetries, resource limits ride in the UpdateEntities body, and
// the response parses the containerUpdateResponse {ID} schema.
func TestContainerUpdateSwaggerCompliance(t *testing.T) {
	client, err := NewRESTClient(RESTConfig{
		Endpoint:   "http://podman.test",
		APIVersion: "5.4.2",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodPost || request.URL.Path != "/v5.4.2/libpod/containers/abc123/update" {
				t.Fatalf("request = %s %s", request.Method, request.URL.Path)
			}
			query := request.URL.Query()
			for key, want := range map[string]string{
				"restartPolicy": "on-failure", "restartRetries": "3",
			} {
				if got := query.Get(key); got != want {
					t.Fatalf("query %s = %q, want %q", key, got, want)
				}
			}
			for _, banned := range []string{"memory", "cpus", "restartMaxRetries"} {
				if query.Has(banned) {
					t.Fatalf("query must not contain %q (swagger ContainerUpdateLibpod)", banned)
				}
			}
			var body dto.UpdateEntities
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Memory == nil || body.Memory.Limit != 512*1024*1024 {
				t.Fatalf("body.Memory = %+v", body.Memory)
			}
			if body.CPU == nil || body.CPU.Quota != 200000 || body.CPU.Period != 100000 {
				t.Fatalf("body.CPU = %+v", body.CPU)
			}
			return response(http.StatusCreated, `{"ID":"abc123"}`), nil
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ContainerUpdate(context.Background(), "abc123", dto.ContainerUpdateOptions{
		Memory: 512 * 1024 * 1024, NanoCPUs: 2_000_000_000,
		RestartPolicy: "on-failure", RestartMaxRetries: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "abc123" {
		t.Fatalf("update ID = %q", result.ID)
	}
}

// TestAdvancedContainerPathEscapesID verifies that IDs containing URL-
// unsafe characters are properly percent-encoded.
func TestAdvancedContainerPathEscapesID(t *testing.T) {
	id := "container/with spaces"
	got := ContainerUpdatePath(id)
	want := "/libpod/containers/container%2Fwith%20spaces/update"
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}
