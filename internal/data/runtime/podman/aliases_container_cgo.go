//go:build cgo

package podman

// Type aliases to Podman native types.
// The HTTP REST JSON responses use the same field names and types as these
// Go structs, so they can be deserialized directly without custom DTOs.
//
// Aliases avoid duplication: the CGO bindings path and the REST path share
// the same struct definitions from the Podman driver.

import (
	commonTypes "go.podman.io/common/libnetwork/types"
	"go.podman.io/podman/v6/pkg/domain/entities/types"
)

// ContainerItem is an alias for the Podman container list DTO.
// Verified compatible with Podman 5.4.2 /containers/json response.
type ContainerItem = types.ListContainer

// ContainerPort is an alias for the Podman port mapping DTO.
type ContainerPort = commonTypes.PortMapping
