package keyboard

import (
	"context"
	"github.com/elizabevil/docker-tui/internal/data/docker"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	"charm.land/bubbletea/v2"
)

func FetchContainers(client *docker.Client, all bool) tea.Cmd {
	return func() tea.Msg {
		containers, err := client.ListContainers(docker.ContainerListOptions{
			All: all,
		})
		return state.ContainersLoaded{
			Containers: containers,
			Error:      err,
		}
	}
}

func FetchImages(client *docker.Client) tea.Cmd {
	return func() tea.Msg {
		images, err := client.Images().List(context.Background(), runtimeapi.ImageListOptions{})
		return state.ImagesLoaded{
			Images: images,
			Error:  err,
		}
	}
}

func FetchVolumes(client *docker.Client) tea.Cmd {
	return func() tea.Msg {
		volumes, err := client.ListVolumes()
		return state.VolumesLoaded{
			Volumes: volumes,
			Error:   err,
		}
	}
}

func FetchNetworks(client *docker.Client) tea.Cmd {
	return func() tea.Msg {
		networks, err := client.ListNetworks()
		return state.NetworksLoaded{
			Networks: networks,
			Error:    err,
		}
	}
}

func FetchAll(client *docker.Client) []tea.Cmd {
	return []tea.Cmd{
		FetchContainers(client, true),
		FetchImages(client),
		FetchVolumes(client),
		FetchNetworks(client),
	}
}
