package docker

import runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"

// dockerEngine is the Docker adapter exposed through the runtime boundary.
type dockerEngine struct{ *Client }

// podmanEngine is the Podman adapter exposed through the runtime boundary.
// Transport selection and wire mapping remain private to the data package.
type podmanEngine struct{ *Client }

var _ runtimeapi.Engine = (*dockerEngine)(nil)
var _ runtimeapi.Engine = (*podmanEngine)(nil)

// NewEngine creates the runtime adapter selected by the validated connection
// configuration. Callers outside the data layer should use this factory rather
// than constructing a concrete Client.
func NewEngine(config ClientConfig) (runtimeapi.Engine, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}
	return engineForClient(client), nil
}

func engineForClient(client *Client) runtimeapi.Engine {
	if client.RuntimeType == RuntimePodman {
		return &podmanEngine{Client: client}
	}
	return &dockerEngine{Client: client}
}

func (e *podmanEngine) Containers() runtimeapi.ContainerService {
	return podmanContainerService{client: e.Client}
}

func (e *podmanEngine) Events() runtimeapi.EventService {
	return podmanEventService{client: e.Client}
}

func (e *podmanEngine) Exec() runtimeapi.ExecService {
	return podmanExecService{client: e.Client}
}
