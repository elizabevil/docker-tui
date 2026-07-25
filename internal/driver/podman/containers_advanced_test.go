package podman

import (
	"testing"
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
		{"commit", ContainerCommitPath(id), "/libpod/containers/abc123/commit"},
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
