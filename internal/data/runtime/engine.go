package runtime

import "context"

// Engine is the lifecycle boundary implemented by each runtime adapter.
// Resource services will be added as their existing callers are migrated.
type Engine interface {
	Identity() Identity
	Capabilities() CapabilitySet
	Ping(context.Context) error
	Close() error
}
