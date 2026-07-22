//go:build !cgo

package podman

// newCGODriver is unavailable in the non-CGO build. Callers should
// rely on the REST transport exposed by *RESTClient.
func newCGODriver(host string) any { return nil }

// closeCGODriver is a no-op in the non-CGO build.
func closeCGODriver(any) error { return nil }
