package docker

import (
	"context"
	"fmt"
	"sync"
	"time"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
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
	Engine     runtimeapi.Engine
	State      ConnState
	Error      error
	Runtime    RuntimeType
	TLS        TLSConfig
}

// ConnectionPool manages multiple runtime connections with active selection.
type ConnectionPool struct {
	mu      sync.RWMutex
	entries map[string]*PoolEntry
	order   []string
	active  string // currently active host name
}

// NewPool creates an empty connection pool.
func NewPool() *ConnectionPool {
	return &ConnectionPool{
		entries: make(map[string]*PoolEntry),
	}
}

// Validate that HostEntry defines required host only.
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

// Connect establishes or reuses a connection to the named host. When the host
// is already connected, it pings to verify and reuses the existing client.
func (p *ConnectionPool) Connect(name string, timeout time.Duration) error {
	p.mu.Lock()
	entry, ok := p.entries[name]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("unknown host: %s", name)
	}
	if entry.State == StateConnected && entry.Engine != nil {
		// Already connected — ping to verify
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		err := entry.Engine.PingContext(ctx)
		cancel()
		if err != nil {
			entry.State = StateDisconnected
		} else {
			p.active = name
			p.mu.Unlock()
			return nil
		}
	}
	entry.State = StateConnecting
	entry.Error = nil
	host := entry.Host
	apiVersion := entry.APIVersion
	runtimeType := entry.Runtime
	tlsConfig := entry.TLS
	p.mu.Unlock()

	engine, err := NewEngine(ClientConfig{Host: host, APIVersion: apiVersion, Timeout: timeout, Runtime: runtimeType, TLS: tlsConfig})
	p.mu.Lock()
	if err != nil {
		entry.State = StateError
		entry.Error = err
		p.mu.Unlock()
		return err
	}
	// Close the previous engine if any.
	if entry.Engine != nil {
		entry.Engine.Close()
	}
	entry.Engine = engine
	entry.State = StateConnected
	entry.Error = nil
	p.active = name
	p.mu.Unlock()
	return nil
}

// Probe checks one connection without changing the active runtime.
func (p *ConnectionPool) Probe(name string, timeout time.Duration) error {
	p.mu.RLock()
	entry := p.entries[name]
	if entry == nil {
		p.mu.RUnlock()
		return fmt.Errorf("unknown host: %s", name)
	}
	host, apiVersion, runtimeType, tlsConfig, existing := entry.Host, entry.APIVersion, entry.Runtime, entry.TLS, entry.Engine
	p.mu.RUnlock()

	var err error
	if existing != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		err = existing.PingContext(ctx)
		cancel()
	} else {
		var engine runtimeapi.Engine
		engine, err = NewEngine(ClientConfig{Host: host, APIVersion: apiVersion, Timeout: timeout, Runtime: runtimeType, TLS: tlsConfig})
		if engine != nil {
			_ = engine.Close()
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if current := p.entries[name]; current != nil && current.Engine == existing {
		current.Error = err
		if err != nil {
			current.State = StateError
		} else {
			current.State = StateConnected
		}
	}
	return err
}

// ActiveEngine is the SDK-independent connection used by migrated callers.
func (p *ConnectionPool) ActiveEngine() runtimeapi.Engine {
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
			entry.Engine.Close()
		}
	}
	p.entries = make(map[string]*PoolEntry)
	p.order = nil
}

// PingLoop periodically pings all connections until the context is cancelled.
// Failed pings transition the connection to StateDisconnected so the UI can
// react to connectivity loss.
func (p *ConnectionPool) PingLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.pingAll()
		}
	}
}

func (p *ConnectionPool) pingAll() {
	p.mu.RLock()
	entries := make([]*PoolEntry, 0, len(p.entries))
	for _, e := range p.entries {
		entries = append(entries, e)
	}
	p.mu.RUnlock()
	for _, e := range entries {
		if e.Engine != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			err := e.Engine.PingContext(ctx)
			cancel()
			if err != nil {
				p.mu.Lock()
				e.State = StateDisconnected
				e.Error = err
				p.mu.Unlock()
			}
		}
	}
}
