//go:build cgo

package dto

import "go.podman.io/podman/v6/pkg/domain/entities/types"

// ImageItem is an alias for the Podman native image list DTO when CGO is
// enabled. In non-CGO builds it is the independent struct with matching
// JSON tags.
type ImageItem = types.ImageSummary
