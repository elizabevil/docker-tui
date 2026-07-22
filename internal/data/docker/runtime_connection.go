package docker

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/elizabevil/docker-tui/internal/data/config"
	"github.com/elizabevil/docker-tui/internal/utils"
)

// ConnectionSpec is the normalized runtime connection model used by the
// startup flow, connection pool and runtime selector.
type ConnectionSpec struct {
	Name       string
	Host       string
	APIVersion string
	Runtime    RuntimeType
	TLS        TLSConfig
}

// HostEntry preserves the legacy name used by the connection pool API.
// It aliases ConnectionSpec so existing call sites continue to compile while
// the codebase moves to the normalized model.
type HostEntry = ConnectionSpec

// FromRuntimeConn converts the config model into the normalized runtime model.
func FromRuntimeConn(conn config.RuntimeConn) ConnectionSpec {
	return ConnectionSpec{
		Name:       conn.Name,
		Host:       conn.Endpoint,
		APIVersion: conn.APIVersion,
		Runtime:    RuntimeType(conn.Driver),
		TLS: TLSConfig{
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

// PodmanUserEndpoint returns the Podman socket URI for the given UID.
// Root (uid 0) uses the system-level socket; non-root uses the user-level socket.
func PodmanUserEndpoint(uid int) string {
	if uid == 0 {
		return utils.SocketURI(DefaultPodmanSocket)
	}
	return utils.SocketURI(filepath.Join("/", "run", "user", fmt.Sprintf("%d", uid), "podman", "podman.sock"))
}
