package podman

// Podman Libpod REST API endpoint path constants.
const (
	// System
	PathSystemVersion = "/libpod/version"

	// Containers
	PathContainerList    = "/containers/json"
	PathContainerInspect = "/containers/%s/json"
	PathContainerTop     = "/containers/%s/top"
	PathContainerStats   = "/containers/%s/stats"
	PathContainerLogs    = "/containers/%s/logs"
	PathContainerStart   = "/containers/%s/start"
	PathContainerStop    = "/containers/%s/stop"
	PathContainerRestart = "/containers/%s/restart"
	PathContainerKill    = "/containers/%s/kill"
	PathContainerPause   = "/containers/%s/pause"
	PathContainerUnpause = "/containers/%s/unpause"
	PathContainerRename  = "/containers/%s/rename"
	PathContainerRemove  = "/containers/%s"
	PathContainerExec    = "/containers/%s/exec"

	// Exec
	PathExecStart  = "/exec/%s/start"
	PathExecResize = "/exec/%s/resize"

	// Images
	PathImageList   = "/images/json"
	PathImagePull   = "/images/pull"
	PathImagePush   = "/images/%s/push"
	PathImageTag    = "/images/%s/tag"
	PathImageRemove = "/images/%s"
	PathImagePrune  = "/images/prune"
	PathImageSave   = "/images/export"
	PathImageLoad   = "/images/load"

	// Volumes
	PathVolumeList    = "/volumes/json"
	PathVolumeInspect = "/volumes/%s/json"
	PathVolumeCreate  = "/volumes/create"
	PathVolumeRemove  = "/volumes/%s"
	PathVolumePrune   = "/volumes/prune"

	// Networks
	PathNetworkList    = "/networks/json"
	PathNetworkInspect = "/networks/%s/json"
	PathNetworkCreate  = "/networks/create"
	PathNetworkRemove  = "/networks/%s"
	PathNetworkPrune   = "/networks/prune"

	// Events
	PathEvents = "/events"
)
