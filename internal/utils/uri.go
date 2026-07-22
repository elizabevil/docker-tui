package utils

import "net/url"

// SocketURI builds a unix:// URI from a filesystem socket path.
// The returned string is suitable for use as a Docker/Podman host endpoint.
func SocketURI(path string) string {
	return (&url.URL{Scheme: "unix", Path: path}).String()
}
