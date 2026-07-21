package docker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ConnState int

const (
	StateDisconnected ConnState = iota
	StateConnecting
	StateConnected
	StateError
)

type HostEntry struct {
	Name    string
	Host    string
	Runtime string
	TLS     TLSConfig
}

type PoolEntry struct {
	Name    string
	Host    string
	Client  *Client
	State   ConnState
	Error   error
	Runtime string
	TLS     TLSConfig
}

type ConnectionPool struct {
	mu      sync.RWMutex
	entries map[string]*PoolEntry
	order   []string
	active  string // currently active host name
}

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
		Name:    he.Name,
		Host:    he.Host,
		Runtime: he.Runtime,
		TLS:     he.TLS,
		State:   StateDisconnected,
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

func (p *ConnectionPool) Get(name string) *PoolEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.entries[name]
}

func (p *ConnectionPool) Active() *PoolEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.entries[p.active]
}

func (p *ConnectionPool) ActiveName() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.active
}

func (p *ConnectionPool) SetActive(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.active = name
}

func (p *ConnectionPool) Connect(name string, timeout time.Duration) error {
	p.mu.Lock()
	entry, ok := p.entries[name]
	if !ok {
		p.mu.Unlock()
		return fmt.Errorf("unknown host: %s", name)
	}
	if entry.State == StateConnected && entry.Client != nil {
		// Already connected — ping to verify
		if err := entry.Client.Ping(); err != nil {
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
	runtimeType := entry.Runtime
	tlsConfig := entry.TLS
	p.mu.Unlock()

	client, err := NewClient(ClientConfig{Host: host, Timeout: timeout, Runtime: runtimeType, TLS: tlsConfig})
	p.mu.Lock()
	if err != nil {
		entry.State = StateError
		entry.Error = err
		p.mu.Unlock()
		return err
	}
	// Close old client if any
	if entry.Client != nil {
		entry.Client.Close()
	}
	entry.Client = client
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
	host, runtimeType, tlsConfig, existing := entry.Host, entry.Runtime, entry.TLS, entry.Client
	p.mu.RUnlock()

	var err error
	if existing != nil {
		err = existing.PingTimeout(timeout)
	} else {
		var client *Client
		client, err = NewClient(ClientConfig{Host: host, Timeout: timeout, Runtime: runtimeType, TLS: tlsConfig})
		if client != nil {
			_ = client.Close()
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if current := p.entries[name]; current != nil && current.Client == existing {
		current.Error = err
		if err != nil {
			current.State = StateError
		} else {
			current.State = StateConnected
		}
	}
	return err
}

func (p *ConnectionPool) ActiveClient() *Client {
	p.mu.RLock()
	defer p.mu.RUnlock()
	entry := p.entries[p.active]
	if entry == nil {
		return nil
	}
	return entry.Client
}

func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, entry := range p.entries {
		if entry.Client != nil {
			entry.Client.Close()
		}
	}
	p.entries = make(map[string]*PoolEntry)
	p.order = nil
}

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
		if e.Client != nil {
			if err := e.Client.Ping(); err != nil {
				p.mu.Lock()
				e.State = StateDisconnected
				e.Error = err
				p.mu.Unlock()
			}
		}
	}
}
