package keyboard

import (
	"context"

	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
	"github.com/elizabevil/docker-tui/internal/tui/state"

	"charm.land/bubbletea/v2"
)

func FetchContainers(client runtimeapi.Engine, all bool) tea.Cmd {
	return func() tea.Msg {
		containers, err := client.Containers().List(context.Background(), runtimeapi.ContainerListOptions{
			All: all,
		})
		return state.ContainersLoaded{
			Containers: containers,
			Error:      err,
		}
	}
}

func FetchImages(client runtimeapi.Engine) tea.Cmd {
	return func() tea.Msg {
		images, err := client.Images().List(context.Background(), runtimeapi.ImageListOptions{})
		return state.ImagesLoaded{
			Images: images,
			Error:  err,
		}
	}
}

func FetchVolumes(client runtimeapi.Engine) tea.Cmd {
	return func() tea.Msg {
		volumes, err := client.Volumes().List(context.Background(), runtimeapi.VolumeListOptions{})
		return state.VolumesLoaded{
			Volumes: volumes,
			Error:   err,
		}
	}
}

func FetchNetworks(client runtimeapi.Engine) tea.Cmd {
	return func() tea.Msg {
		networks, err := client.Networks().List(context.Background(), runtimeapi.NetworkListOptions{})
		return state.NetworksLoaded{
			Networks: networks,
			Error:    err,
		}
	}
}

func FetchAll(client runtimeapi.Engine) []tea.Cmd {
	return []tea.Cmd{
		FetchContainers(client, true),
		FetchImages(client),
		FetchVolumes(client),
		FetchNetworks(client),
	}
}
