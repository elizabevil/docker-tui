package docker

import (
	"testing"

	podmanapi "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

func TestEngineForClientSelectsRuntimeAdapter(t *testing.T) {
	rest, err := podmanapi.NewRESTClient(podmanapi.RESTConfig{Endpoint: "http://podman.test"})
	if err != nil {
		t.Fatalf("NewRESTClient() error = %v", err)
	}

	tests := []struct {
		name   string
		client *Client
		check  func(any) bool
	}{
		{name: "docker", client: &Client{RuntimeType: RuntimeDocker}, check: func(value any) bool { _, ok := value.(*dockerEngine); return ok }},
		{name: "podman", client: &Client{RuntimeType: RuntimePodman, podmanREST: rest}, check: func(value any) bool { _, ok := value.(*podmanapi.Engine); return ok }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if engine := engineForClient(test.client); !test.check(engine) {
				t.Fatalf("engine type = %T", engine)
			}
		})
	}
}
