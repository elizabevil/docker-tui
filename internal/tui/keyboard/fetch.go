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
		FetchComposeProjects(client),
	}
}

// FetchComposeProjects asks the runtime's ComposeService for the
// current project list. R08-03 sets All=true so the panel matches
// `docker compose ls --all` semantics: running + stopped containers
// aggregate into projects; projects without any container are not
// synthesised (the aggregator's source-of-truth is the container
// label, see R08-03 §决策记录).
func FetchComposeProjects(client runtimeapi.Engine) tea.Cmd {
	return func() tea.Msg {
		projects, err := client.Compose().ListProjects(context.Background(), runtimeapi.ListProjectsOptions{All: true})
		return state.ComposeProjectsLoaded{Projects: projects, Error: err}
	}
}
