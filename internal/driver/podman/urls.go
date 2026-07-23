package podman

import (
	"fmt"
	"net/url"
	"path"
)

// Podman Libpod REST API endpoint path constants.
const (
	PathSystemVersion = "/libpod/version"
	PathContainerList = "/libpod/containers/json"
	PathImageList     = "/libpod/images/json"
	PathImagePull     = "/libpod/images/pull"
	PathImagePrune    = "/libpod/images/prune"
	PathImageSave     = "/libpod/images/export"
	PathImageLoad     = "/libpod/images/load"
	PathVolumeList    = "/libpod/volumes/json"
	PathVolumeCreate  = "/libpod/volumes/create"
	PathVolumePrune   = "/libpod/volumes/prune"
	PathNetworkList   = "/libpod/networks/json"
	PathNetworkCreate = "/libpod/networks/create"
	PathNetworkPrune  = "/libpod/networks/prune"
	PathEvents        = "/libpod/events"
)

// Container lifecycle action constants.
const (
	ActionStart   = "start"
	ActionStop    = "stop"
	ActionRestart = "restart"
	ActionKill    = "kill"
	ActionPause   = "pause"
	ActionUnpause = "unpause"
	ActionRename  = "rename"
	ActionRemove  = "remove"
	ActionCreate  = "create"
	ActionExec    = "exec"
	ActionPrune   = "prune"
	ActionPull    = "pull"
	ActionTag     = "tag"
	ActionPush    = "push"
)

// Podman REST API resource path bases for URL construction.
var (
	ContainersURL = &url.URL{Path: "/libpod/containers"}
	ExecURL       = &url.URL{Path: "/libpod/exec"}
	ImagesURL     = &url.URL{Path: "/libpod/images"}
	VolumesURL    = &url.URL{Path: "/libpod/volumes"}
	NetworksURL   = &url.URL{Path: "/libpod/networks"}
)

// escape returns a single URL segment for id.
func escape(id string) string {
	return url.PathEscape(id)
}

// Containers
func ContainerPath(id, suffix string) string {
	return (&url.URL{Path: path.Join(ContainersURL.Path, escape(id)) + suffix}).Path
}
func ContainerInspectPath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s/json", escape(id))
}
func ContainerTopPath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s/top", escape(id))
}
func ContainerStatsPath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s/stats", escape(id))
}
func ContainerLogsPath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s/logs", escape(id))
}
func ContainerStartPath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s/start", escape(id))
}
func ContainerStopPath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s/stop", escape(id))
}
func ContainerRestartPath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s/restart", escape(id))
}
func ContainerKillPath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s/kill", escape(id))
}
func ContainerPausePath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s/pause", escape(id))
}
func ContainerUnpausePath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s/unpause", escape(id))
}
func ContainerRenamePath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s/rename", escape(id))
}
func ContainerRemovePath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s", escape(id))
}
func ContainerExecPath(id string) string {
	return fmt.Sprintf("/libpod/containers/%s/exec", escape(id))
}

// Exec
func ExecPath(id, suffix string) string {
	return (&url.URL{Path: path.Join(ExecURL.Path, escape(id)) + suffix}).Path
}
func ExecStartPath(id string) string {
	return fmt.Sprintf("/libpod/exec/%s/start", escape(id))
}
func ExecResizePath(id string) string {
	return fmt.Sprintf("/libpod/exec/%s/resize", escape(id))
}

// Images
func ImagePath(id, suffix string) string {
	return (&url.URL{Path: path.Join(ImagesURL.Path, escape(id)) + suffix}).Path
}
func ImageTagPath(id string) string {
	return fmt.Sprintf("/libpod/images/%s/tag", escape(id))
}
func ImagePushPath(id string) string {
	return fmt.Sprintf("/libpod/images/%s/push", escape(id))
}
func ImageRemovePath(id string) string {
	return fmt.Sprintf("/libpod/images/%s", escape(id))
}
func ImageInspectPath(id string) string {
	return fmt.Sprintf("/libpod/images/%s/json", escape(id))
}

// Volumes
func VolumePath(name, suffix string) string {
	return (&url.URL{Path: path.Join(VolumesURL.Path, escape(name)) + suffix}).Path
}
func VolumeInspectPath(name string) string {
	return fmt.Sprintf("/libpod/volumes/%s/json", escape(name))
}
func VolumeRemovePath(name string) string {
	return fmt.Sprintf("/libpod/volumes/%s", escape(name))
}

// Networks
func NetworkPath(id, suffix string) string {
	return (&url.URL{Path: path.Join(NetworksURL.Path, escape(id)) + suffix}).Path
}
func NetworkInspectPath(id string) string {
	return fmt.Sprintf("/libpod/networks/%s/json", escape(id))
}
func NetworkRemovePath(id string) string {
	return fmt.Sprintf("/libpod/networks/%s", escape(id))
}
