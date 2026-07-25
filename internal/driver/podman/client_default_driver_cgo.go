//go:build cgo

package podman

// defaultDriverFor returns a DriverBackend that uses the official Podman
// Go bindings. Connections are opened lazily by the driver.
func defaultDriverFor(host string) DriverBackend {
	return newCGODriver(host)
}
