package runtime

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/elizabevil/docker-tui/internal/utils"
)

// ConnState tracks the lifecycle phase of a pooled connection.
type ConnState int

// Connection states managed by the pool.
const (
	StateDisconnected ConnState = iota
	StateConnecting
	StateConnected
	StateError
)

// PoolEntry holds the state and client for a single pooled connection.
type PoolEntry struct {
	Name       string
	Host       string
	APIVersion string
	Engine     Engine
	State      ConnState
	Error      error
	Runtime    RuntimeType
	TLS        utils.TLSConfig
	Latency    time.Duration
	ProbedAt   time.Time
}

// EngineFactory creates an Engine from a ClientConfig. It is injected into
// the ConnectionPool so the pool can spin up engines without depending on a
// concrete docker/podman package (which would create an import cycle with
// the runtime package).
type EngineFactory func(ClientConfig) (Engine, error)

// ConnectionPool manages multiple runtime connections with active selection.
// The factory is required: it knows how to materialize an Engine for a given
// ClientConfig and is typically wired up by the runtimeinit package.
type ConnectionPool struct {
	factory EngineFactory
	mu      sync.RWMutex
	entries map[string]*PoolEntry
	order   []string
	active  string // currently active host name
}

// NewPool creates an empty connection pool bound to the given engine factory.
// A nil factory is allowed for tests that only exercise metadata operations
// (AddHost/Get/KnownHostNames) and never trigger Connect or RefreshAll.
func NewPool(factory EngineFactory) *ConnectionPool {
	return &ConnectionPool{
		factory: factory,
		entries: make(map[string]*PoolEntry),
	}
}

// SetEngineFactory replaces the engine factory. Useful for tests that swap
// adapters after construction; production code wires the factory once in
// NewPool.
func (p *ConnectionPool) SetEngineFactory(factory EngineFactory) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.factory = factory
}

// newEngineWithFactory creates an engine using a factory captured outside the
// lock. Connect and refreshOne copy the factory before dropping the lock so
// concurrent SetEngineFactory calls don't race with in-flight connections.
func (p *ConnectionPool) newEngineWithFactory(factory EngineFactory, cfg ClientConfig) (Engine, error) {
	if factory == nil {
		return nil, fmt.Errorf("connection pool has no engine factory: pass one to NewPool")
	}
	e, err := factory(cfg)
	return sanitizeEngine(e, err), err
}

// sanitizeEngine collapses a typed-nil Engine interface back to a true nil
// interface. This guards against the classic Go gotcha where a concrete
// adapter returns (nilPtr, err): wrapping the nil pointer in an interface
// makes `e == nil` false and a subsequent method call would panic. Adapters
// should return literal nil on error; this is a safety net for any that
// forget.
func sanitizeEngine(e Engine, err error) Engine {
	if e == nil {
		return nil
	}
	v := reflect.ValueOf(e)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	return e
}

// engineIsUsable reports whether engine holds a real implementation that is
// safe to call methods on. It catches both nil interfaces and typed-nil
// pointers.
func engineIsUsable(engine Engine) bool {
	if engine == nil {
		return false
	}
	v := reflect.ValueOf(engine)
	return !(v.Kind() == reflect.Ptr && v.IsNil())
}

// AddHost registers a connection candidate by its name.
func (p *ConnectionPool) AddHost(he HostEntry) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.entries[he.Name]; ok {
		return
	}
	p.entries[he.Name] = &PoolEntry{
		Name:       he.Name,
		Host:       he.Host,
		APIVersion: he.APIVersion,
		Runtime:    he.Runtime,
		TLS:        he.TLS,
		State:      StateDisconnected,
	}
	p.order = append(p.order, he.Name)
}

// KnownHostNames returns all host names.
func (p *ConnectionPool) KnownHostNames() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	names := append([]string(nil), p.order...)
	return names
}

// Get returns the pool entry for the given host name, or nil if unknown.
func (p *ConnectionPool) Get(name string) *PoolEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.entries[name]
}

// Active returns the pool entry for the currently active connection.
func (p *ConnectionPool) Active() *PoolEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.entries[p.active]
}

// ActiveName returns the name of the currently active connection.
func (p *ConnectionPool) ActiveName() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.active
}

// SetActive marks the named entry as the active connection.
func (p *ConnectionPool) SetActive(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.active = name
}

// Connect establishes or reuses a connection to the named host, measures its
// baseline latency and sets it as the active connection.
func (p *ConnectionPool) Connect(name string, timeout time.Duration) error {
	p.mu.Lock()
	entry, ok := p.entries[name]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("unknown host: %s", name)
	}

	host := entry.Host
	apiVersion := entry.APIVersion
	runtimeType := entry.Runtime
	tlsConfig := entry.TLS

	var (
		engine  Engine
		err     error
		latency time.Duration
	)

	if entry.State == StateConnected && engineIsUsable(entry.Engine) {
		engine = entry.Engine
		latency, err = pingEngine(engine, timeout)
		if err != nil {
			entry.State = StateDisconnected
		}
	}

	if engine == nil {
		entry.State = StateConnecting
		entry.Error = nil
		factory := p.factory
		p.mu.Unlock()

		start := time.Now()
		engine, err = p.newEngineWithFactory(factory, ClientConfig{
			Host:       host,
			APIVersion: apiVersion,
			Timeout:    timeout,
			Runtime:    runtimeType,
			TLS:        tlsConfig,
		})
		latency = time.Since(start)
		p.mu.Lock()
	}

	if err != nil {
		entry.State = StateError
		entry.Error = err
		entry.Latency = 0
		entry.ProbedAt = time.Now()
		p.mu.Unlock()
		return err
	}

	if entry.Engine != nil && entry.Engine != engine {
		_ = entry.Engine.Close() //nolint:errcheck // closing a stale engine; failure is non-fatal during reconnect.
	}
	entry.Engine = engine
	entry.State = StateConnected
	entry.Error = nil
	entry.Latency = latency
	entry.ProbedAt = time.Now()
	p.active = name
	p.mu.Unlock()
	return nil
}

// RefreshAll iterates through every known host and updates its latency/state
// without changing the active connection. Hosts without an engine get a
// transient engine created to probe availability, then closed.
func (p *ConnectionPool) RefreshAll(timeout time.Duration) []RefreshResult {
	p.mu.RLock()
	snapshot := make([]*PoolEntry, 0, len(p.entries))
	for _, name := range p.order {
		if e, ok := p.entries[name]; ok {
			snapshot = append(snapshot, e)
		}
	}
	p.mu.RUnlock()

	results := make([]RefreshResult, 0, len(snapshot))
	for _, e := range snapshot {
		results = append(results, p.refreshOne(e, timeout))
	}
	return results
}

// refreshOne probes a single entry and updates its latency/state in place.
// It avoids touching the active connection pointer; callers that want to
// connect use Connect instead.
func (p *ConnectionPool) refreshOne(entry *PoolEntry, timeout time.Duration) RefreshResult {
	name := entry.Name
	var (
		engine    Engine
		err       error
		latency   time.Duration
		transient bool
	)

	p.mu.RLock()
	existing := entry.Engine
	p.mu.RUnlock()

	if engineIsUsable(existing) {
		engine = existing
		latency, err = pingEngine(engine, timeout)
	} else {
		transient = true
		p.mu.RLock()
		factory := p.factory
		p.mu.RUnlock()
		start := time.Now()
		engine, err = p.newEngineWithFactory(factory, ClientConfig{
			Host:       entry.Host,
			APIVersion: entry.APIVersion,
			Timeout:    timeout,
			Runtime:    entry.Runtime,
			TLS:        entry.TLS,
		})
		// Adapter factories like docker.NewClient already ping during
		// construction, but Podman only allocates the struct — no HTTP
		// traffic. Ping here so the recorded latency reflects the actual
		// round-trip instead of just struct allocation time, and a failed
		// socket surfaces as StateError instead of pretending connected.
		if err == nil && engineIsUsable(engine) {
			latency, err = pingEngine(engine, timeout)
		} else {
			latency = time.Since(start)
		}
		if engineIsUsable(engine) {
			_ = engine.Close() //nolint:errcheck // closing a probe engine; failure is non-fatal.
		}
	}

	p.mu.Lock()
	current := p.entries[name]
	if current == nil {
		p.mu.Unlock()
		return RefreshResult{Name: name, Error: err, Latency: latency, Transient: transient}
	}
	current.Error = err
	current.Latency = latency
	current.ProbedAt = time.Now()
	if err != nil {
		current.State = StateError
	} else {
		current.State = StateConnected
	}
	p.mu.Unlock()

	return RefreshResult{Name: name, Error: err, Latency: latency, Transient: transient}
}

// pingEngine measures the round-trip latency of an existing engine via ping.
func pingEngine(engine Engine, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err := engine.PingContext(ctx)
	cancel()
	return time.Since(start), err
}

// ActiveEngine is the SDK-independent connection used by migrated callers.
func (p *ConnectionPool) ActiveEngine() Engine {
	p.mu.RLock()
	defer p.mu.RUnlock()
	entry := p.entries[p.active]
	if entry == nil {
		return nil
	}
	return entry.Engine
}

// Close shuts down all pooled connections and clears the pool.
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, entry := range p.entries {
		if entry.Engine != nil {
			_ = entry.Engine.Close() //nolint:errcheck // best-effort shutdown during pool reset.
		}
	}
	p.entries = make(map[string]*PoolEntry)
	p.order = nil
}

// PingLoop periodically pings all connected engines until the context is
// cancelled. Failed pings transition the connection to StateError so the UI
// can react to connectivity loss.
func (p *ConnectionPool) PingLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.pingConnected()
		}
	}
}

// pingConnected updates latency for entries that already hold an engine.
// Entries without an engine are skipped — RefreshAll handles those.
func (p *ConnectionPool) pingConnected() {
	p.mu.RLock()
	entries := make([]*PoolEntry, 0, len(p.entries))
	for _, e := range p.entries {
		if e.Engine != nil {
			entries = append(entries, e)
		}
	}
	p.mu.RUnlock()

	for _, e := range entries {
		latency, err := pingEngine(e.Engine, 3*time.Second)
		p.mu.Lock()
		if err != nil {
			e.State = StateError
			e.Error = err
			e.Latency = 0
		} else {
			e.Latency = latency
		}
		e.ProbedAt = time.Now()
		p.mu.Unlock()
	}
}

// RefreshResult is the per-entry outcome returned by ConnectionPool.RefreshAll.
type RefreshResult struct {
	Name      string
	Error     error
	Latency   time.Duration
	Transient bool // true if the engine was created and closed just for the probe
}
