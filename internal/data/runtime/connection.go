package runtime

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// RuntimeType identifies the container engine backend (Docker or Podman).
type RuntimeType string

const (
	// RuntimeDocker selects the Docker Engine backend.
	RuntimeDocker RuntimeType = "docker"
	// RuntimePodman selects the Podman backend.
	RuntimePodman RuntimeType = "podman"

	// DefaultDockerSocket is the well-known Docker daemon socket path.
	DefaultDockerSocket = "/var/run/docker.sock"
)

// Default podman socket used when PodmanUserEndpoint is unavailable.
const defaultPodmanSocket = "/var/run/podman/podman.sock"

// ClientConfig holds the parameters needed to create a runtime client.
type ClientConfig struct {
	Host       string
	APIVersion string
	TLS        utils.TLSConfig
	Timeout    time.Duration
	Runtime    RuntimeType
}

// ConnectionSpec is the normalized runtime connection model used by the
// startup flow, connection pool and runtime selector.
type ConnectionSpec struct {
	Name       string
	Host       string
	APIVersion string
	Runtime    RuntimeType
	TLS        utils.TLSConfig
}

// HostEntry preserves the legacy name used by the connection pool API.
// It aliases ConnectionSpec so existing call sites continue to compile while
// the codebase moves to the normalized model.
type HostEntry = ConnectionSpec

// ConnectionOption toggles optional behaviours of BuildConnections.
type ConnectionOption func(*connectionBuilder)

// connectionBuilder collects the inputs to BuildConnections.
type connectionBuilder struct {
	hostOverride string
	usePodman    bool
	podmanUserFn func(uid int) string
	getuidFn     func() int
	defaultName  string
}

// WithHostOverride forces a single connection from a CLI flag.
func WithHostOverride(host string, usePodman bool) ConnectionOption {
	return func(b *connectionBuilder) { b.hostOverride = host; b.usePodman = usePodman }
}

// WithPodmanResolver overrides the user-endpoint resolver (used in tests).
func WithPodmanResolver(fn func(uid int) string) ConnectionOption {
	return func(b *connectionBuilder) { b.podmanUserFn = fn }
}

// WithUIDSource overrides os.Getuid (used in tests).
func WithUIDSource(fn func() int) ConnectionOption {
	return func(b *connectionBuilder) { b.getuidFn = fn }
}

// WithDefaultName overrides the initial connection name from config.
func WithDefaultName(name string) ConnectionOption {
	return func(b *connectionBuilder) { b.defaultName = name }
}

// BuildConnections assembles the connection catalogue in priority order:
// local docker → local podman → user-defined runtime.Connections. Deduplicates
// by ConnectionKey so a config entry that matches the local docker socket
// doesn't appear twice. Returns the ordered specs and the initial connection
// name (resolved against alias mappings).
func BuildConnections(cfg *config.RuntimeConfig, opts ...ConnectionOption) ([]ConnectionSpec, string) {
	b := &connectionBuilder{
		podmanUserFn: defaultPodmanUserEndpoint,
		getuidFn:     os.Getuid,
		defaultName:  cfg.Default,
	}
	for _, opt := range opts {
		opt(b)
	}

	if b.hostOverride != "" {
		driver := RuntimeDocker
		if b.usePodman {
			driver = RuntimePodman
		}
		return []ConnectionSpec{{Name: "cli", Host: b.hostOverride, Runtime: driver}}, "cli"
	}

	connections := make([]ConnectionSpec, 0, len(cfg.Connections)+2)
	seen := make(map[string]int, len(cfg.Connections)+2)
	aliases := make(map[string]string, len(cfg.Connections)+2)
	add := func(entry ConnectionSpec, replace bool) {
		key := entry.Key()
		if index, exists := seen[key]; exists {
			if replace {
				aliases[connections[index].Name] = entry.Name
				connections[index] = entry
			}
			return
		}
		seen[key] = len(connections)
		connections = append(connections, entry)
	}
	// Local runtimes are always candidates; unavailable sockets fail visibly
	// in the selector rather than being hidden.
	add(ConnectionSpec{Name: "local-docker", Host: utils.SocketURI(DefaultDockerSocket), Runtime: RuntimeDocker}, false)
	add(ConnectionSpec{Name: "local-podman", Host: b.podmanUserFn(b.getuidFn()), Runtime: RuntimePodman}, false)
	for _, conn := range cfg.Connections {
		add(FromRuntimeConnection(conn), true)
	}

	initial := b.defaultName
	if b.usePodman {
		initial = "local-podman"
	}
	if canonical, exists := aliases[initial]; exists {
		initial = canonical
	}
	return connections, initial
}

// defaultPodmanUserEndpoint is the production resolver for the per-user podman
// socket. It is package-level so tests can override it via WithPodmanResolver.
func defaultPodmanUserEndpoint(uid int) string {
	if uid <= 0 {
		return utils.SocketURI(defaultPodmanSocket)
	}
	return utils.SocketURI(fmt.Sprintf("/run/user/%d/podman/podman.sock", uid))
}

// FromRuntimeConnection converts the config model into the normalized runtime model.
func FromRuntimeConnection(conn config.RuntimeConnection) ConnectionSpec {
	return ConnectionSpec{
		Name:       conn.Name,
		Host:       conn.Endpoint,
		APIVersion: conn.APIVersion,
		Runtime:    NormalizeRuntimeType(RuntimeType(string(conn.Driver))),
		TLS: utils.TLSConfig{
			Enabled:            conn.TLS.Enabled,
			Verify:             conn.TLS.Verify,
			InsecureSkipVerify: conn.TLS.InsecureSkipVerify,
			CAFile:             conn.TLS.CAFile,
			CertFile:           conn.TLS.CertFile,
			KeyFile:            conn.TLS.KeyFile,
			ServerName:         conn.TLS.ServerName,
		},
	}
}

// Key returns the deduplication key for this runtime connection.
func (c ConnectionSpec) Key() string {
	return fmt.Sprintf("%s|api=%s|tls=%t|verify=%t|server=%s|ca=%s|cert=%s",
		ConnectionKey(c.Runtime, c.Host), c.APIVersion, c.TLS.Enabled,
		!c.TLS.InsecureSkipVerify, c.TLS.ServerName, c.TLS.CAFile, c.TLS.CertFile)
}

// ConnectionKey normalizes driver and endpoint into a stable deduplication key.
func ConnectionKey(runtime RuntimeType, endpoint string) string {
	runtime = NormalizeRuntimeType(runtime)
	endpoint = strings.TrimSpace(endpoint)
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" {
		return string(runtime) + "|" + endpoint
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	if parsed.Scheme == "unix" {
		parsed.Path = filepath.Clean(parsed.Path)
	}
	return string(runtime) + "|" + parsed.String()
}

// NormalizeRuntimeType canonicalizes runtime labels.
func NormalizeRuntimeType(runtime RuntimeType) RuntimeType {
	switch strings.ToLower(strings.TrimSpace(string(runtime))) {
	case string(RuntimePodman):
		return RuntimePodman
	default:
		return RuntimeDocker
	}
}
