package keys

import "github.com/elizabevil/docker-tui/internal/data/config"

// KeyMapping maps key sequences to actions.
type KeyMapping map[string]KeyAction

// DefaultKeyMapping returns the default key bindings.
func DefaultKeyMapping() KeyMapping {
	bindings := CompileBindings(config.KeymapConfig{})
	mapping := make(KeyMapping, len(bindings.ByKey))
	for key, actions := range bindings.ByKey {
		if len(actions) > 0 {
			mapping[key] = actions[0]
		}
	}
	return mapping
}
