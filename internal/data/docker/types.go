package docker

import runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"

type ContainerListOptions = runtimeapi.ContainerListOptions
type ContainerSummary = runtimeapi.ContainerSummary
type PortBinding = runtimeapi.PortBinding
type ContainerProcesses = runtimeapi.ContainerProcesses

type ManifestPlatform = runtimeapi.ManifestPlatform
type ImageManifestEntry = runtimeapi.ImageManifestEntry
type ImageSummary = runtimeapi.ImageSummary
type ImageRuntimeConfig = runtimeapi.ImageRuntimeConfig
type ImageHistoryLayer = runtimeapi.ImageHistoryLayer
type ImageHistorySource = runtimeapi.ImageHistorySource

const (
	ImageHistoryPending  = runtimeapi.ImageHistoryPending
	ImageHistoryLayerAPI = runtimeapi.ImageHistoryLayerAPI
	ImageHistoryManifest = runtimeapi.ImageHistoryManifest
)

type ImageDetailData = runtimeapi.ImageDetail
