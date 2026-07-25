package state

type ResourceState struct {
	Containers *ContainerListModel
	Images     *ImageListModel
	Volumes    *VolumeListModel
	Networks   *NetworkListModel
}

func NewResourceState() ResourceState {
	return ResourceState{
		Containers: NewContainerListModel(),
		Images:     NewImageListModel(),
		Volumes:    NewVolumeListModel(),
		Networks:   NewNetworkListModel(),
	}
}
