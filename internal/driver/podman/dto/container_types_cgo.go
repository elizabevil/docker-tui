//go:build cgo

package dto

import (
	commonTypes "go.podman.io/common/libnetwork/types"
	"go.podman.io/podman/v6/pkg/domain/entities/types"
)

// ContainerItem is an alias for the Podman native container list DTO when
// CGO is enabled. In non-CGO builds it is the independent struct with
// matching JSON tags.
type ContainerItem = types.ListContainer

// ContainerPort is an alias for the Podman native port mapping DTO when
// CGO is enabled.
type ContainerPort = commonTypes.PortMapping
