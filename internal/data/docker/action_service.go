package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/errdefs"
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
		return runtimeapi.ActionResult{}, mapRuntimeError(err, string(ref.Type)+"."+string(action), ref, s.client.RuntimeType)
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
			return runtimeapi.UnsupportedError("volume." + string(action))
		}
		if s.client.RuntimeType == RuntimePodman {
			return s.client.removeVolumePodman(ctx, ref.ID, options.Force)
		}
		return s.client.cli.VolumeRemove(ctx, ref.ID, options.Force)
	case runtimeapi.ResourceNetwork:
		if action != runtimeapi.ActionRemove {
			return runtimeapi.UnsupportedError("network." + string(action))
		}
		if s.client.RuntimeType == RuntimePodman {
			return s.client.removeNetworkPodman(ctx, ref.ID)
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
			return runtimeapi.NewError(runtimeapi.ErrorInvalid, "container.rename", id, fmt.Errorf("name is required"))
		}
		return s.client.cli.ContainerRename(ctx, id, options.Name)
	case runtimeapi.ActionRemove:
		return s.client.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: options.Force})
	default:
		return runtimeapi.UnsupportedError("container." + string(action))
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
		return runtimeapi.UnsupportedError("image." + string(action))
	}
}

func durationSeconds(timeout time.Duration) *int {
	if timeout <= 0 {
		return nil
	}
	seconds := int(timeout.Seconds())
	return &seconds
}

type statusCoder interface{ Code() int }

func mapRuntimeError(err error, operation string, ref runtimeapi.ResourceRef, driver RuntimeType) error {
	if err == nil {
		return nil
	}
	var existing *runtimeapi.Error
	if errors.As(err, &existing) {
		mapped := *existing
		if mapped.ResourceType == "" {
			mapped.ResourceType = string(ref.Type)
		}
		if mapped.Resource == "" {
			mapped.Resource = ref.ID
		}
		if mapped.Driver == "" {
			mapped.Driver = runtimeapi.Type(driver)
		}
		return &mapped
	}
	kind := runtimeapi.ClassifyContextError(err)
	if kind == runtimeapi.ErrorInternal {
		switch {
		case errors.Is(err, os.ErrNotExist):
			kind = runtimeapi.ErrorNotFound
		case errors.Is(err, os.ErrExist):
			kind = runtimeapi.ErrorConflict
		case errors.Is(err, os.ErrPermission):
			kind = runtimeapi.ErrorPermission
		case errdefs.IsNotFound(err):
			kind = runtimeapi.ErrorNotFound
		case errdefs.IsConflict(err):
			kind = runtimeapi.ErrorConflict
		case errdefs.IsInvalidParameter(err):
			kind = runtimeapi.ErrorInvalid
		case errdefs.IsUnauthorized(err):
			kind = runtimeapi.ErrorAuthentication
		case errdefs.IsForbidden(err):
			kind = runtimeapi.ErrorPermission
		case errdefs.IsNotImplemented(err):
			kind = runtimeapi.ErrorUnsupported
		case errdefs.IsUnavailable(err):
			kind = runtimeapi.ErrorUnavailable
		case errdefs.IsDeadline(err):
			kind = runtimeapi.ErrorTimeout
		case errdefs.IsCancelled(err):
			kind = runtimeapi.ErrorCanceled
		default:
			var coded statusCoder
			if errors.As(err, &coded) {
				kind = runtimeapi.ClassifyHTTPStatus(coded.Code())
			}
		}
	}
	return &runtimeapi.Error{Kind: kind, Operation: operation, ResourceType: string(ref.Type), Resource: ref.ID, Driver: runtimeapi.Type(driver), Retryable: runtimeapi.IsRetryableKind(kind), Err: err}
}
