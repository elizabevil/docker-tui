package podman

import (
	"testing"
	"time"
)

// TestContainerActionQuerySwaggerCompliance locks the query parameters to
// what docs/api/podman-swagger.yaml defines for each endpoint:
//
//	stop   — timeout, ignore (no signal)
//	restart — t, timeout
//	kill   — signal
//	rename — name
//	remove — depend, force, ignore, timeout, v
//
// Unknown parameters are rejected by the Podman API server, so a query
// that drifts from the swagger definition surfaces as a 400 on the wire.
func TestContainerActionQuerySwaggerCompliance(t *testing.T) {
	cases := []struct {
		name    string
		action  string
		options ContainerActionOptions
		allowed []string
		absent  []string
	}{
		{
			name:   "stop",
			action: ActionStop,
			options: ContainerActionOptions{
				Timeout: 5 * time.Second,
				Signal:  "SIGKILL",
			},
			allowed: []string{"timeout"},
			absent:  []string{"signal"},
		},
		{
			name:   "restart",
			action: ActionRestart,
			options: ContainerActionOptions{
				Timeout: 10 * time.Second,
				Signal:  "SIGKILL",
			},
			allowed: []string{"timeout"},
			absent:  []string{"signal"},
		},
		{
			name:   "kill",
			action: ActionKill,
			options: ContainerActionOptions{
				Signal: "SIGTERM",
			},
			allowed: []string{"signal"},
			absent:  []string{"timeout"},
		},
		{
			name:   "rename",
			action: ActionRename,
			options: ContainerActionOptions{
				Name: "new-name",
			},
			allowed: []string{"name"},
			absent:  []string{"timeout", "signal"},
		},
		{
			name:   "remove",
			action: ActionRemove,
			options: ContainerActionOptions{
				Force:         true,
				RemoveVolumes: true,
			},
			allowed: []string{"force", "v"},
			absent:  []string{"timeout", "signal"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			query := tc.options.Query(tc.action)
			for _, key := range tc.allowed {
				if query.Get(key) == "" {
					t.Errorf("Query(%s) missing allowed parameter %q (got %v)", tc.action, key, query)
				}
			}
			for _, key := range tc.absent {
				if query.Get(key) != "" {
					t.Errorf("Query(%s) contains parameter %q not defined by swagger (got %v)", tc.action, key, query)
				}
			}
		})
	}
}
