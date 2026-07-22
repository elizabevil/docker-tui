package docker

import "testing"

func TestEngineForClientSelectsRuntimeAdapter(t *testing.T) {
	tests := []struct {
		name   string
		client *Client
		check  func(any) bool
	}{
		{name: "docker", client: &Client{RuntimeType: RuntimeDocker}, check: func(value any) bool { _, ok := value.(*dockerEngine); return ok }},
		{name: "podman", client: &Client{RuntimeType: RuntimePodman}, check: func(value any) bool { _, ok := value.(*podmanEngine); return ok }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if engine := engineForClient(test.client); !test.check(engine) {
				t.Fatalf("engine type = %T", engine)
			}
		})
	}
}
