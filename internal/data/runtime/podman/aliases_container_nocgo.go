//go:build !cgo

package podman

import "github.com/elizabevil/docker-tui/internal/data/runtime/podman/dto"

// ContainerItem is an alias for dto.ContainerItem. The non-CGO path uses
// the independent REST DTO; the CGO path aliases the Podman native type.
type ContainerItem = dto.ContainerItem

// ContainerPort is an alias for dto.ContainerPort.
type ContainerPort = dto.ContainerPort
