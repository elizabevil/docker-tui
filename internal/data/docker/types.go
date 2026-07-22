package docker

import runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"

// Type aliases re-export runtime domain types so existing callers within the
// docker package continue to compile without import changes. New code should
// import the runtime package directly.

type ContainerListOptions = runtimeapi.ContainerListOptions
type ContainerSummary = runtimeapi.ContainerSummary
type PortBinding = runtimeapi.PortBinding
type ContainerProcesses = runtimeapi.ContainerProcesses

// Image types are aliases to the runtime domain model. The docker package
// retains these aliases for backward compatibility; the runtime package is
// the canonical location for these definitions.
type ManifestPlatform = runtimeapi.ManifestPlatform
type ImageManifestEntry = runtimeapi.ImageManifestEntry
type ImageSummary = runtimeapi.ImageSummary
type ImageRuntimeConfig = runtimeapi.ImageRuntimeConfig
type ImageHistoryLayer = runtimeapi.ImageHistoryLayer
type ImageHistorySource = runtimeapi.ImageHistorySource

// ImageHistorySource constants re-exported for callers that still reference
// the docker package path.
const (
	ImageHistoryPending  = runtimeapi.ImageHistoryPending
	ImageHistoryLayerAPI = runtimeapi.ImageHistoryLayerAPI
	ImageHistoryManifest = runtimeapi.ImageHistoryManifest
)

// ImageDetailData is an alias for runtime.ImageDetail. The name is preserved
// for backward compatibility with existing TUI code; the canonical type is
// runtime.ImageDetail.
type ImageDetailData = runtimeapi.ImageDetail
