package runtime

import (
	"context"
	"sync"
	"time"
)

// InspectFields is the subset of container inspect data that drives
// CoLocated group derivation. The CoLocated resolver in each adapter
// fetches these via an InspectFn and caches results with a 5-second TTL
// (per R08-15 Q2). Inspect-derived fields on ContainerSummary
// (NetworkMode / PidMode / IpcMode / CoLocatedGroupID / CoLocatedGroupSrc)
// are populated from this struct.
//
// The shared cache is thread-safe and process-global per adapter; each
// adapter's FillCoLocated writes through its own cache. The events-driven
// invalidation path is wired by the adapter (per R08-15 Q2).
type InspectFields struct {
	NetworkMode string
	PidMode     string
	IpcMode     string
}

// InspectFn fetches InspectFields for a single container. Adapters
// supply their own implementation backed by the appropriate transport
// (Docker SDK / Podman REST).
type InspectFn func(ctx context.Context, containerID string) (InspectFields, error)

// CoLocatedCache memoizes InspectFn results with a configurable TTL.
// Reads are safe for concurrent use. Events-driven invalidation is
// the adapter's responsibility (R08-15 Q2).
type CoLocatedCache struct {
	ttl     time.Duration
	mu      sync.Mutex
	entries map[string]cacheEntry
}

type cacheEntry struct {
	fields  InspectFields
	expires time.Time
}

// NewCoLocatedCache builds a cache with the given TTL. The recommended
// default (5s) is set at the call site in each adapter.
func NewCoLocatedCache(ttl time.Duration) *CoLocatedCache {
	return &CoLocatedCache{
		ttl:     ttl,
		entries: make(map[string]cacheEntry),
	}
}

// Get returns cached inspect fields, or invokes fn to populate. fn errors
// are propagated without caching (caller can retry).
func (c *CoLocatedCache) Get(ctx context.Context, id string, fn InspectFn) (InspectFields, error) {
	c.mu.Lock()
	if e, ok := c.entries[id]; ok && time.Now().Before(e.expires) {
		c.mu.Unlock()
		return e.fields, nil
	}
	c.mu.Unlock()

	fields, err := fn(ctx, id)
	if err != nil {
		return InspectFields{}, err
	}

	c.mu.Lock()
	c.entries[id] = cacheEntry{fields: fields, expires: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return fields, nil
}

// Invalidate clears all entries. Used by events handlers when a container
// state change should bust the cache.
func (c *CoLocatedCache) Invalidate() {
	c.mu.Lock()
	c.entries = make(map[string]cacheEntry)
	c.mu.Unlock()
}
