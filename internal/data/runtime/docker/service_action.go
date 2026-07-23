package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	runtimeapi "github.com/elizabevil/docker-tui/internal/data/runtime"
)

// resourceActionService implements runtime.ResourceActionService for the
// Docker/Podman adapter. It translates generic actions into SDK-specific calls.
type resourceActionService struct{ client *Client }

// Actions returns the resource action service facade.
func (c *Client) Actions() runtimeapi.ResourceActionService {
	return resourceActionService{client: c}
}

// Execute performs a lifecycle action on a resource. The action is routed to
// the appropriate SDK call based on the resource type and action kind.
func (s resourceActionService) Execute(ctx context.Context, ref runtimeapi.ResourceRef, action runtimeapi.Action, options runtimeapi.ActionOptions) (runtimeapi.ActionResult, error) {
	result := runtimeapi.ActionResult{Resource: ref, Action: action}
	err := s.execute(ctx, ref, action, options, &result)
	if err != nil {
		return runtimeapi.ActionResult{}, runtimeapi.MapRuntimeError(err, runtimeapi.Operation(ref.Type, string(action)), ref, runtimeapi.Docker)
	}
	return result, nil
}

func (s resourceActionService) execute(ctx context.Context, ref runtimeapi.ResourceRef, action runtimeapi.Action, options runtimeapi.ActionOptions, result *runtimeapi.ActionResult) error {
	switch ref.Type {
	case runtimeapi.ResourceContainer:
		return s.executeContainer(ctx, ref.ID, action, options)
	case runtimeapi.ResourceImage:
		return s.executeImage(ctx, ref.ID, action, options, result)
	case runtimeapi.ResourceVolume:
		if action != runtimeapi.ActionRemove {
			return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceVolume, string(action)))
		}
		return s.client.cli.VolumeRemove(ctx, ref.ID, options.Force)
	case runtimeapi.ResourceNetwork:
		if action != runtimeapi.ActionRemove {
			return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceNetwork, string(action)))
		}
		return s.client.cli.NetworkRemove(ctx, ref.ID)
	default:
		return runtimeapi.NewError(runtimeapi.ErrorInvalid, "resource.action", ref.ID, fmt.Errorf("unknown resource type %q", ref.Type))
	}
}

func (s resourceActionService) executeContainer(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions) error {
	switch action {
	case runtimeapi.ActionStart:
		return s.client.cli.ContainerStart(ctx, id, container.StartOptions{})
	case runtimeapi.ActionStop:
		return s.client.cli.ContainerStop(ctx, id, container.StopOptions{Timeout: durationSeconds(options.Timeout)})
	case runtimeapi.ActionRestart:
		return s.client.cli.ContainerRestart(ctx, id, container.StopOptions{Timeout: durationSeconds(options.Timeout)})
	case runtimeapi.ActionKill:
		return s.client.cli.ContainerKill(ctx, id, options.Signal)
	case runtimeapi.ActionPause:
		return s.client.cli.ContainerPause(ctx, id)
	case runtimeapi.ActionUnpause:
		return s.client.cli.ContainerUnpause(ctx, id)
	case runtimeapi.ActionRename:
		if options.Name == "" {
			return runtimeapi.NewError(runtimeapi.ErrorInvalid, runtimeapi.Operation(runtimeapi.ResourceContainer, string(action)), id, fmt.Errorf("name is required"))
		}
		return s.client.cli.ContainerRename(ctx, id, options.Name)
	case runtimeapi.ActionRemove:
		return s.client.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: options.Force})
	default:
		return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceContainer, string(action)))
	}
}

func (s resourceActionService) executeImage(ctx context.Context, id string, action runtimeapi.Action, options runtimeapi.ActionOptions, result *runtimeapi.ActionResult) error {
	switch action {
	case runtimeapi.ActionRemove:
		_, err := s.client.cli.ImageRemove(ctx, id, image.RemoveOptions{Force: options.Force})
		return err
	case runtimeapi.ActionPull:
		reader, err := s.client.cli.ImagePull(ctx, id, image.PullOptions{})
		if err != nil {
			return err
		}
		defer reader.Close()
		_, err = io.Copy(io.Discard, reader)
		return err
	case runtimeapi.ActionPrune:
		report, err := s.client.cli.ImagesPrune(ctx, filters.NewArgs())
		result.SpaceReclaimed = report.SpaceReclaimed
		return err
	default:
		return runtimeapi.UnsupportedError(runtimeapi.Operation(runtimeapi.ResourceImage, string(action)))
	}
}

func durationSeconds(timeout time.Duration) *int {
	if timeout <= 0 {
		return nil
	}
	seconds := int(timeout.Seconds())
	return &seconds
}
