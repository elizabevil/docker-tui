package docker

import (
	"testing"
	"time"
)

func TestMapPodmanContainerSummariesExpandsPortRange(t *testing.T) {
	created := time.Unix(100, 0)
	result := mapPodmanContainerSummaries([]podmanContainerSummary{{
		ID: "1234567890123456", Names: []string{"api"}, Created: created,
		Labels: map[string]string{"com.docker.compose.project": "demo"},
		Ports:  []podmanPort{{ContainerPort: 8080, HostPort: 9080, Range: 2, Protocol: "tcp", HostIP: "127.0.0.1"}},
	}})
	if len(result) != 1 || len(result[0].PortBindings) != 2 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result[0].ID != "1234567890123456" || result[0].PortBindings[1].ContainerPort != 8081 || result[0].PortBindings[1].HostPort != 9081 {
		t.Fatalf("unexpected mapping: %#v", result[0])
	}
	if result[0].ComposeProject != "demo" {
		t.Fatalf("compose project = %q", result[0].ComposeProject)
	}
}
