package podman

import "fmt"

// ErrPodmanRESTNotReady is returned when a Podman REST operation is attempted
// without an initialized REST transport.
var ErrPodmanRESTNotReady = fmt.Errorf("podman REST transport is not initialized")

// Client wraps both the REST transport and an optional CGO driver
// connection. Methods on Client prefer the driver when available and
// fall back to REST otherwise, providing a unified API across build
// modes. The driver is wired up by NewClient; callers that need to
// replace it can do so before issuing the first request.
type Client struct {
	// Driver is set to a DriverBackend implementation when CGO is enabled
	// and the caller wants to use the official Podman bindings. It is nil
	// when the build does not have CGO available, in which case all
	// operations go through REST.
	Driver DriverBackend

	REST *RESTClient
	Host string
	TLS  bool
}

// NewClient creates a Podman client from a RESTClient and connection info.
// In CGO builds, NewClient also installs a cgoDriver as the preferred
// transport. In non-CGO builds, only the REST transport is configured.
func NewClient(rest *RESTClient, host string, tls bool) *Client {
	c := &Client{REST: rest, Host: host, TLS: tls}
	c.attachDefaultDriver()
	return c
}

// attachDefaultDriver wires up the build-appropriate default driver.
// Implementation lives in client_default_driver_cgo.go / _nocgo.go.
func (c *Client) attachDefaultDriver() {
	c.Driver = defaultDriverFor(c.Host)
}

// SetDriver replaces the active driver. Pass nil to fall back to REST only.
func (c *Client) SetDriver(d DriverBackend) { c.Driver = d }
