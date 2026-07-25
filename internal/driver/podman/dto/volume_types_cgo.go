//go:build cgo

package dto

import "go.podman.io/podman/v6/libpod/define"

// VolumeItem is an alias for the Podman native volume inspect/list DTO
// when CGO is enabled. In non-CGO builds it is the independent struct
// with matching JSON tags.
type VolumeItem = define.InspectVolumeData
