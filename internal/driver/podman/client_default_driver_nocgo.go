//go:build !cgo

package podman

// defaultDriverFor returns nil in the non-CGO build so that all Client
// methods fall through to the REST transport.
func defaultDriverFor(host string) DriverBackend {
	return nil
}
