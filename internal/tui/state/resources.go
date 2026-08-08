package state

import runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"

type ResourceState struct {
	Containers *ContainerListModel
	Images     *ImageListModel
	Volumes    *VolumeListModel
	Networks   *NetworkListModel

	// Compose holds the runtime-aggregated compose project summaries
	// (R08-03). Populated by FetchComposeProjects on every refresh;
	// when empty the compose panel falls back to label aggregation
	// over Containers.Items so first-paint is never blank.
	Compose []runtimeapi.ComposeProjectSummary
}

func NewResourceState() ResourceState {
	return ResourceState{
		Containers: NewContainerListModel(),
		Images:     NewImageListModel(),
		Volumes:    NewVolumeListModel(),
		Networks:   NewNetworkListModel(),
	}
}
