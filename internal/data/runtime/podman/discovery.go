package podman

import (
	"fmt"
	"path/filepath"

	"github.com/elizabevil/docker-tui/internal/utils"
)

// DefaultPodmanSocket is the system-level Podman socket path.
const DefaultPodmanSocket = "/run/podman/podman.sock"

// PodmanUserEndpoint returns the Podman socket URI for the given UID.
// Root (uid 0) uses the system-level socket; non-root uses the user-level socket.
func PodmanUserEndpoint(uid int) string {
	if uid == 0 {
		return utils.SocketURI(DefaultPodmanSocket)
	}
	return utils.SocketURI(filepath.Join("/", "run", "user", fmt.Sprintf("%d", uid), "podman", "podman.sock"))
}
