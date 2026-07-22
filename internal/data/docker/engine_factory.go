package docker

import (
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	podmanapi "github.com/elizabevil/docker-tui/internal/data/runtime/podman"
)

// dockerEngine is the Docker adapter exposed through the runtime boundary.
type dockerEngine struct{ *Client }

var _ runtimeapi.Engine = (*dockerEngine)(nil)

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
		if rest := client.podmanREST; rest != nil {
			if e, err := podmanapi.NewEngine(podmanapi.EngineConfig{
				REST: rest,
				Host: client.Host,
				TLS:  client.TLS.Enabled,
			}); err == nil {
				return e
			}
		}
	}
	return &dockerEngine{Client: client}
}
