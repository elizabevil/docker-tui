package podman

import (
	"net/url"
	"path"
)

// Podman REST API path helpers.
// Each function builds a URL path segment with proper escaping.

func ContainerPath(id, suffix string) string {
	return path.Join("/containers", url.PathEscape(id)) + suffix
}

func ExecPath(id, suffix string) string {
	return path.Join("/exec", url.PathEscape(id)) + suffix
}

func ImagePath(id, suffix string) string {
	return path.Join("/images", url.PathEscape(id)) + suffix
}

func VolumePath(name, suffix string) string {
	return path.Join("/volumes", url.PathEscape(name)) + suffix
}

func NetworkPath(id, suffix string) string {
	return path.Join("/networks", url.PathEscape(id)) + suffix
}
